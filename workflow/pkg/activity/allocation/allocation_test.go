/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	cdb "github.com/nvidia/bare-metal-manager-rest/db/pkg/db"
	cdbm "github.com/nvidia/bare-metal-manager-rest/db/pkg/db/model"
	cdbu "github.com/nvidia/bare-metal-manager-rest/db/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/extra/bundebug"
	"google.golang.org/protobuf/types/known/timestamppb"

	cwutil "github.com/nvidia/bare-metal-manager-rest/common/pkg/util"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func testAllocationInitDB(t *testing.T) *cdb.Session {
	dbSession := cdbu.GetTestDBSession(t, false)
	dbSession.DB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return dbSession
}

func testAllocationSetupSchema(t *testing.T, dbSession *cdb.Session) {
	err := dbSession.DB.ResetModel(context.Background(), (*cdbm.InfrastructureProvider)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.Site)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.Tenant)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.User)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.Allocation)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.AllocationConstraint)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.TenantSite)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.StatusDetail)(nil))
	assert.NoError(t, err)
	err = dbSession.DB.ResetModel(context.Background(), (*cdbm.InstanceType)(nil))
	assert.NoError(t, err)
}

func testAllocationBuildComputeAllocation(allocationID uuid.UUID, tenantOrg, name, description string, instanceTypeID uuid.UUID, count uint32) *cwssaws.ComputeAllocation {
	return &cwssaws.ComputeAllocation{
		Id: &cwssaws.ComputeAllocationId{
			Value: allocationID.String(),
		},
		TenantOrganizationId: tenantOrg,
		Metadata: &cwssaws.Metadata{
			Name:        name,
			Description: description,
		},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: instanceTypeID.String(),
			Count:          count,
		},
	}
}

func TestManageAllocation_UpdateAllocationsInDB(t *testing.T) {
	ctx := context.Background()
	dbSession := testAllocationInitDB(t)
	defer dbSession.Close()

	testAllocationSetupSchema(t, dbSession)

	user := cdbm.TestBuildUser(t, dbSession, uuid.NewString(), "test-provider-org", []string{"FORGE_PROVIDER_ADMIN"})
	ip := cdbm.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cdbm.TestBuildSite(t, dbSession, ip, "test-site", user)
	tenant := cdbm.TestBuildTenant(t, dbSession, "tenant-1", "test-tenant-org", user)
	otherTenant := cdbm.TestBuildTenant(t, dbSession, "tenant-2", "other-tenant-org", user)

	instanceType1 := cdbm.TestBuildInstanceType(t, dbSession, "test-instance-type-1", ip, site, user)
	instanceType2 := cdbm.TestBuildInstanceType(t, dbSession, "test-instance-type-2", ip, site, user)

	allocationDAO := cdbm.NewAllocationDAO(dbSession)
	constraintDAO := cdbm.NewAllocationConstraintDAO(dbSession)
	tenantSiteDAO := cdbm.NewTenantSiteDAO(dbSession)
	statusDAO := cdbm.NewStatusDetailDAO(dbSession)

	staleAllocation := cdbm.TestBuildAllocation(t, dbSession, "stale-allocation", site, tenant, user)
	staleConstraint, err := constraintDAO.CreateFromParams(ctx, nil, staleAllocation.ID, cdbm.AllocationResourceTypeInstanceType, instanceType1.ID, cdbm.AllocationConstraintTypeReserved, 3, nil, user.ID)
	require.NoError(t, err)
	_, err = dbSession.DB.Exec("UPDATE allocation SET created = ? WHERE id = ?", time.Now().Add(-(cwutil.InventoryReceiptInterval + 10*time.Second)), staleAllocation.ID)
	require.NoError(t, err)

	freshAllocation := cdbm.TestBuildAllocation(t, dbSession, "fresh-allocation", site, otherTenant, user)
	_, err = constraintDAO.CreateFromParams(ctx, nil, freshAllocation.ID, cdbm.AllocationResourceTypeInstanceType, instanceType1.ID, cdbm.AllocationConstraintTypeReserved, 9, nil, user.ID)
	require.NoError(t, err)

	updateAllocation := cdbm.TestBuildAllocation(t, dbSession, "existing-allocation", site, tenant, user)
	updateConstraint, err := constraintDAO.CreateFromParams(ctx, nil, updateAllocation.ID, cdbm.AllocationResourceTypeInstanceType, instanceType1.ID, cdbm.AllocationConstraintTypeReserved, 2, nil, user.ID)
	require.NoError(t, err)

	manager := NewManageAllocation(dbSession, nil)
	newAllocationID := uuid.New()
	inventory := &cwssaws.ComputeAllocationInventory{
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
		Timestamp:       timestamppb.Now(),
		Allocations: []*cwssaws.ComputeAllocation{
			testAllocationBuildComputeAllocation(updateAllocation.ID, tenant.Org, "updated-allocation", "updated-description", instanceType2.ID, 8),
			testAllocationBuildComputeAllocation(newAllocationID, tenant.Org, "new-allocation", "new-description", instanceType1.ID, 5),
		},
	}

	err = manager.UpdateAllocationsInDB(ctx, site.ID, inventory)
	require.NoError(t, err)

	_, err = allocationDAO.GetByID(ctx, nil, staleAllocation.ID, nil)
	assert.ErrorIs(t, err, cdb.ErrDoesNotExist)
	_, err = constraintDAO.GetByID(ctx, nil, staleConstraint.ID, nil)
	assert.ErrorIs(t, err, cdb.ErrDoesNotExist)

	freshRet, err := allocationDAO.GetByID(ctx, nil, freshAllocation.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, freshAllocation.ID, freshRet.ID)

	updatedRet, err := allocationDAO.GetByID(ctx, nil, updateAllocation.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, "updated-allocation", updatedRet.Name)
	require.NotNil(t, updatedRet.Description)
	assert.Equal(t, "updated-description", *updatedRet.Description)
	updatedConstraintRet, err := constraintDAO.GetByID(ctx, nil, updateConstraint.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, instanceType2.ID, updatedConstraintRet.ResourceTypeID)
	assert.Equal(t, 8, updatedConstraintRet.ConstraintValue)

	newRet, err := allocationDAO.GetByID(ctx, nil, newAllocationID, nil)
	require.NoError(t, err)
	assert.Equal(t, "new-allocation", newRet.Name)
	assert.Equal(t, cdbm.AllocationStatusRegistered, newRet.Status)
	require.NotNil(t, newRet.Description)
	assert.Equal(t, "new-description", *newRet.Description)
	newConstraints, total, err := constraintDAO.GetAll(ctx, nil, []uuid.UUID{newAllocationID}, cdb.GetStrPtr(cdbm.AllocationResourceTypeInstanceType), nil, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	assert.Equal(t, instanceType1.ID, newConstraints[0].ResourceTypeID)
	assert.Equal(t, 5, newConstraints[0].ConstraintValue)

	tenantSite, err := tenantSiteDAO.GetByTenantIDAndSiteID(ctx, nil, tenant.ID, site.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, tenant.Org, tenantSite.TenantOrg)

	statuses, total, err := statusDAO.GetAllByEntityID(ctx, nil, newAllocationID.String(), nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	assert.Equal(t, cdbm.AllocationStatusRegistered, statuses[0].Status)
}

func TestManageAllocation_UpdateAllocationsInDB_PagedDeletion(t *testing.T) {
	ctx := context.Background()
	dbSession := testAllocationInitDB(t)
	defer dbSession.Close()

	testAllocationSetupSchema(t, dbSession)

	user := cdbm.TestBuildUser(t, dbSession, uuid.NewString(), "test-provider-org", []string{"FORGE_PROVIDER_ADMIN"})
	ip := cdbm.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cdbm.TestBuildSite(t, dbSession, ip, "test-site", user)
	tenant := cdbm.TestBuildTenant(t, dbSession, "tenant-1", "test-tenant-org", user)
	instanceType := cdbm.TestBuildInstanceType(t, dbSession, "test-instance-type", ip, site, user)

	allocationDAO := cdbm.NewAllocationDAO(dbSession)
	constraintDAO := cdbm.NewAllocationConstraintDAO(dbSession)

	staleAllocation := cdbm.TestBuildAllocation(t, dbSession, "stale-allocation", site, tenant, user)
	_, err := constraintDAO.CreateFromParams(ctx, nil, staleAllocation.ID, cdbm.AllocationResourceTypeInstanceType, instanceType.ID, cdbm.AllocationConstraintTypeReserved, 3, nil, user.ID)
	require.NoError(t, err)
	_, err = dbSession.DB.Exec("UPDATE allocation SET created = ? WHERE id = ?", time.Now().Add(-(cwutil.InventoryReceiptInterval + 10*time.Second)), staleAllocation.ID)
	require.NoError(t, err)

	manager := NewManageAllocation(dbSession, nil)

	firstPageInventory := &cwssaws.ComputeAllocationInventory{
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
		Timestamp:       timestamppb.Now(),
		Allocations:     []*cwssaws.ComputeAllocation{},
		InventoryPage: &cwssaws.InventoryPage{
			CurrentPage: 1,
			TotalPages:  2,
			PageSize:    10,
			TotalItems:  1,
			ItemIds:     []string{uuid.NewString()},
		},
	}
	err = manager.UpdateAllocationsInDB(ctx, site.ID, firstPageInventory)
	require.NoError(t, err)

	_, err = allocationDAO.GetByID(ctx, nil, staleAllocation.ID, nil)
	require.NoError(t, err)

	lastPageInventory := &cwssaws.ComputeAllocationInventory{
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
		Timestamp:       timestamppb.Now(),
		Allocations:     []*cwssaws.ComputeAllocation{},
		InventoryPage: &cwssaws.InventoryPage{
			CurrentPage: 2,
			TotalPages:  2,
			PageSize:    10,
			TotalItems:  1,
			ItemIds:     []string{uuid.NewString()},
		},
	}
	err = manager.UpdateAllocationsInDB(ctx, site.ID, lastPageInventory)
	require.NoError(t, err)

	_, err = allocationDAO.GetByID(ctx, nil, staleAllocation.ID, nil)
	assert.ErrorIs(t, err, cdb.ErrDoesNotExist)
}

func TestManageAllocation_UpdateAllocationsInDB_NilAndFailedInventory(t *testing.T) {
	ctx := context.Background()
	dbSession := testAllocationInitDB(t)
	defer dbSession.Close()

	testAllocationSetupSchema(t, dbSession)

	user := cdbm.TestBuildUser(t, dbSession, uuid.NewString(), "test-provider-org", []string{"FORGE_PROVIDER_ADMIN"})
	ip := cdbm.TestBuildInfrastructureProvider(t, dbSession, "test-provider", "test-provider-org", user)
	site := cdbm.TestBuildSite(t, dbSession, ip, "test-site", user)

	manager := NewManageAllocation(dbSession, nil)

	err := manager.UpdateAllocationsInDB(ctx, site.ID, nil)
	require.Error(t, err)

	err = manager.UpdateAllocationsInDB(ctx, site.ID, &cwssaws.ComputeAllocationInventory{
		InventoryStatus: cwssaws.InventoryStatus_INVENTORY_STATUS_FAILED,
		StatusMsg:       "site failed inventory collection",
	})
	require.NoError(t, err)
}
