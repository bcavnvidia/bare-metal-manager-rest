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

package activity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	cClient "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func TestManageComputeAllocation_CreateComputeAllocationOnSite(t *testing.T) {
	mockCarbide := cClient.NewMockCarbideClient()
	carbideAtomicClient := cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{})
	carbideAtomicClient.SwapClient(mockCarbide)

	request := &cwssaws.CreateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "test-tenant-org",
		Metadata: &cwssaws.Metadata{
			Name: "test-allocation",
		},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: uuid.NewString(),
			Count:          2,
		},
	}

	tests := []struct {
		name    string
		request *cwssaws.CreateComputeAllocationRequest
		wantErr bool
	}{
		{
			name:    "success",
			request: request,
			wantErr: false,
		},
		{
			name:    "missing id",
			request: &cwssaws.CreateComputeAllocationRequest{TenantOrganizationId: "test-tenant-org"},
			wantErr: true,
		},
		{
			name:    "missing tenant org",
			request: &cwssaws.CreateComputeAllocationRequest{Id: &cwssaws.ComputeAllocationId{Value: uuid.NewString()}},
			wantErr: true,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManageComputeAllocation(carbideAtomicClient)
			err := manager.CreateComputeAllocationOnSite(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageComputeAllocation_UpdateComputeAllocationOnSite(t *testing.T) {
	mockCarbide := cClient.NewMockCarbideClient()
	carbideAtomicClient := cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{})
	carbideAtomicClient.SwapClient(mockCarbide)

	request := &cwssaws.UpdateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "test-tenant-org",
		Metadata: &cwssaws.Metadata{
			Name: "updated-allocation",
		},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: uuid.NewString(),
			Count:          3,
		},
	}

	tests := []struct {
		name    string
		request *cwssaws.UpdateComputeAllocationRequest
		wantErr bool
	}{
		{
			name:    "success",
			request: request,
			wantErr: false,
		},
		{
			name:    "missing id",
			request: &cwssaws.UpdateComputeAllocationRequest{TenantOrganizationId: "test-tenant-org"},
			wantErr: true,
		},
		{
			name:    "missing tenant org",
			request: &cwssaws.UpdateComputeAllocationRequest{Id: &cwssaws.ComputeAllocationId{Value: uuid.NewString()}},
			wantErr: true,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManageComputeAllocation(carbideAtomicClient)
			err := manager.UpdateComputeAllocationOnSite(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManageComputeAllocation_DeleteComputeAllocationOnSite(t *testing.T) {
	mockCarbide := cClient.NewMockCarbideClient()
	carbideAtomicClient := cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{})
	carbideAtomicClient.SwapClient(mockCarbide)

	request := &cwssaws.DeleteComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "test-tenant-org",
	}

	tests := []struct {
		name    string
		request *cwssaws.DeleteComputeAllocationRequest
		wantErr bool
	}{
		{
			name:    "success",
			request: request,
			wantErr: false,
		},
		{
			name:    "missing id",
			request: &cwssaws.DeleteComputeAllocationRequest{TenantOrganizationId: "test-tenant-org"},
			wantErr: true,
		},
		{
			name:    "missing tenant org",
			request: &cwssaws.DeleteComputeAllocationRequest{Id: &cwssaws.ComputeAllocationId{Value: uuid.NewString()}},
			wantErr: true,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManageComputeAllocation(carbideAtomicClient)
			err := manager.DeleteComputeAllocationOnSite(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComputeAllocationInventoryHelpers(t *testing.T) {
	mockCarbide := cClient.NewMockCarbideClient()
	carbideAtomicClient := cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{})
	carbideAtomicClient.SwapClient(mockCarbide)
	carbideClient := carbideAtomicClient.GetClient()

	ctx := context.WithValue(context.Background(), "wantCount", 2)

	ids, err := computeAllocationFindIDs(ctx, carbideClient)
	assert.NoError(t, err)
	assert.Len(t, ids, 2)

	allocations, err := computeAllocationFindByIDs(context.Background(), carbideClient, ids)
	assert.NoError(t, err)
	assert.Len(t, allocations, 2)

	inventory := computeAllocationPagedInventory(ids, allocations, &pagedInventoryInput{
		totalItems:    2,
		totalPages:    1,
		pageSize:      2,
		pageNumber:    1,
		status:        cwssaws.InventoryStatus_INVENTORY_STATUS_SUCCESS,
		statusMessage: "ok",
	})

	assert.Len(t, inventory.GetAllocations(), 2)
	assert.NotNil(t, inventory.GetInventoryPage())
	assert.Len(t, inventory.GetInventoryPage().GetItemIds(), 2)
}
