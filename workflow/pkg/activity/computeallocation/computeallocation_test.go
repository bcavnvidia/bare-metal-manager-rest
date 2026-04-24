/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package computeallocation

import (
	"context"
	"testing"
	"time"

	cwutil "github.com/NVIDIA/ncx-infra-controller-rest/common/pkg/util"
	cdb "github.com/NVIDIA/ncx-infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/ncx-infra-controller-rest/db/pkg/db/model"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	sc "github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/client/site"
	"github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/queue"
	cwu "github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	tmocks "go.temporal.io/sdk/mocks"
)

func TestManageComputeAllocation_UpdateComputeAllocationsInDBSkipsSiteOnlyAllocations(t *testing.T) {
	ctx := context.Background()

	dbSession := cwu.TestInitDB(t)
	defer dbSession.Close()

	cwu.TestSetupSchema(t, dbSession)

	user := cwu.TestBuildUser(t, dbSession, uuid.NewString(), []string{"test-provider-org"}, []string{"FORGE_PROVIDER_ADMIN"})
	ip := cwu.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cwu.TestBuildSite(t, dbSession, ip, "test-site", cdbm.SiteStatusRegistered, nil, user)
	tenant := cwu.TestBuildTenant(t, dbSession, "test-tenant-org", "Test Tenant", nil, user)
	instanceType := cwu.TestBuildInstanceType(t, dbSession, ip, site, "test-instance-type")

	manager := NewManageComputeAllocation(dbSession, nil)

	allocationID := uuid.NewString()
	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations: []*cwssaws.ComputeAllocation{
			{
				Id:                   &cwssaws.ComputeAllocationId{Value: allocationID},
				TenantOrganizationId: tenant.Org,
				Metadata: &cwssaws.Metadata{
					Name:        "allocation-from-site",
					Description: "created from inventory",
				},
				Attributes: &cwssaws.ComputeAllocationAttributes{
					InstanceTypeId: instanceType.ID.String(),
					Count:          4,
				},
			},
		},
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
	}

	err := manager.UpdateComputeAllocationsInDB(ctx, site.ID, inventory)
	require.NoError(t, err)

	allocationDAO := cdbm.NewAllocationDAO(dbSession)
	allocationUUID := uuid.MustParse(allocationID)
	_, err = allocationDAO.GetByID(ctx, nil, allocationUUID, nil)
	assert.Equal(t, cdb.ErrDoesNotExist, err)
}

func TestManageComputeAllocation_UpdateComputeAllocationsInDBUpdatesExistingAllocation(t *testing.T) {
	ctx := context.Background()

	dbSession := cwu.TestInitDB(t)
	defer dbSession.Close()

	cwu.TestSetupSchema(t, dbSession)

	user := cwu.TestBuildUser(t, dbSession, uuid.NewString(), []string{"test-provider-org"}, []string{"FORGE_PROVIDER_ADMIN"})
	ip := cwu.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cwu.TestBuildSite(t, dbSession, ip, "test-site", cdbm.SiteStatusRegistered, nil, user)
	tenant := cwu.TestBuildTenant(t, dbSession, "test-tenant-org", "Test Tenant", nil, user)
	instanceType := cwu.TestBuildInstanceType(t, dbSession, ip, site, "test-instance-type")
	allocation := cwu.TestBuildAllocation(t, dbSession, ip, tenant, site, "old-allocation")
	_ = cwu.TestBuildAllocationContraints(t, dbSession, allocation, cdbm.AllocationResourceTypeInstanceType, instanceType.ID, cdbm.AllocationConstraintTypeReserved, 2, user)

	manager := NewManageComputeAllocation(dbSession, nil)

	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations: []*cwssaws.ComputeAllocation{
			{
				Id:                   &cwssaws.ComputeAllocationId{Value: allocation.ID.String()},
				TenantOrganizationId: tenant.Org,
				Metadata: &cwssaws.Metadata{
					Name:        "allocation-from-site",
					Description: "created from inventory",
				},
				Attributes: &cwssaws.ComputeAllocationAttributes{
					InstanceTypeId: instanceType.ID.String(),
					Count:          4,
				},
			},
		},
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
	}

	err := manager.UpdateComputeAllocationsInDB(ctx, site.ID, inventory)
	require.NoError(t, err)

	allocationDAO := cdbm.NewAllocationDAO(dbSession)
	updatedAllocation, err := allocationDAO.GetByID(ctx, nil, allocation.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, "allocation-from-site", updatedAllocation.Name)
	require.NotNil(t, updatedAllocation.Description)
	assert.Equal(t, "created from inventory", *updatedAllocation.Description)
	assert.Equal(t, tenant.ID, updatedAllocation.TenantID)
	assert.Equal(t, cdbm.AllocationStatusPending, updatedAllocation.Status)

	allocationConstraintDAO := cdbm.NewAllocationConstraintDAO(dbSession)
	constraints, _, err := allocationConstraintDAO.GetAll(ctx, nil, []uuid.UUID{allocation.ID}, cdb.GetStrPtr(cdbm.AllocationResourceTypeInstanceType), nil, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, constraints, 1)
	assert.Equal(t, instanceType.ID, constraints[0].ResourceTypeID)
	assert.Equal(t, 4, constraints[0].ConstraintValue)

	tenantSiteDAO := cdbm.NewTenantSiteDAO(dbSession)
	_, err = tenantSiteDAO.GetByTenantIDAndSiteID(ctx, nil, tenant.ID, site.ID, nil)
	assert.Equal(t, cdb.ErrDoesNotExist, err)
}

func TestManageComputeAllocation_UpdateComputeAllocationsInDBBackfillsMissingAllocations(t *testing.T) {
	ctx := context.Background()

	dbSession := cwu.TestInitDB(t)
	defer dbSession.Close()

	cwu.TestSetupSchema(t, dbSession)

	user := cwu.TestBuildUser(t, dbSession, uuid.NewString(), []string{"test-provider-org"}, []string{"FORGE_PROVIDER_ADMIN"})
	ip := cwu.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cwu.TestBuildSite(t, dbSession, ip, "test-site", cdbm.SiteStatusRegistered, nil, user)
	tenant := cwu.TestBuildTenant(t, dbSession, "test-tenant-org", "Test Tenant", nil, user)
	instanceType := cwu.TestBuildInstanceType(t, dbSession, ip, site, "test-instance-type")
	allocation := cwu.TestBuildAllocation(t, dbSession, ip, tenant, site, "legacy-allocation")
	_ = cwu.TestBuildAllocationContraints(t, dbSession, allocation, cdbm.AllocationResourceTypeInstanceType, instanceType.ID, cdbm.AllocationConstraintTypeReserved, 2, user)

	// Age the allocation so the inventory lag guard allows backfill.
	_, err := dbSession.DB.Exec("UPDATE allocation SET created = ? WHERE id = ?", time.Now().Add(-2*time.Duration(cwutil.InventoryReceiptInterval)), allocation.ID)
	require.NoError(t, err)

	siteClient := &tmocks.Client{}
	workflowRun := &tmocks.WorkflowRun{}
	workflowRun.On("GetID").Return("test-workflow-id")
	workflowRun.On("Get", mock.Anything, mock.Anything).Return(nil).Once()
	siteClient.On("ExecuteWorkflow",
		mock.Anything,
		mock.MatchedBy(func(options client.StartWorkflowOptions) bool {
			return options.ID == "compute-allocation-create-"+allocation.ID.String() &&
				options.TaskQueue == queue.SiteTaskQueue
		}),
		"CreateComputeAllocation",
		mock.MatchedBy(func(request *cwssaws.CreateComputeAllocationRequest) bool {
			return request.GetId().GetValue() == allocation.ID.String() &&
				request.GetTenantOrganizationId() == tenant.Org &&
				request.GetMetadata().GetName() == allocation.Name &&
				request.GetAttributes().GetInstanceTypeId() == instanceType.ID.String() &&
				request.GetAttributes().GetCount() == 2
		}),
	).Return(workflowRun, nil).Once()

	siteClientPool := sc.NewClientPool(nil)
	siteClientPool.IDClientMap[site.ID.String()] = siteClient
	manager := NewManageComputeAllocation(dbSession, siteClientPool)

	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations:     []*cwssaws.ComputeAllocation{},
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
		InventoryPage:   &cwssaws.InventoryPage{CurrentPage: 1, TotalPages: 1, PageSize: 1, TotalItems: 0},
		StatusMsg:       "ok",
	}

	err = manager.UpdateComputeAllocationsInDB(ctx, site.ID, inventory)
	require.NoError(t, err)

	allocationDAO := cdbm.NewAllocationDAO(dbSession)
	_, err = allocationDAO.GetByID(ctx, nil, allocation.ID, nil)
	require.NoError(t, err)
	siteClient.AssertExpectations(t)
	workflowRun.AssertExpectations(t)
}
