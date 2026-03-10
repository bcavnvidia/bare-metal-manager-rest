/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package activity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	cClient "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func TestManageAllocation_CreateUpdateDeleteOnSite(t *testing.T) {
	mockCarbide := cClient.NewMockCarbideClient()
	carbideAtomicClient := cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{})
	carbideAtomicClient.SwapClient(mockCarbide)

	manager := NewManageAllocation(carbideAtomicClient)
	id := &cwssaws.ComputeAllocationId{Value: uuid.NewString()}

	createReq := &cwssaws.CreateComputeAllocationRequest{
		Id:                   id,
		TenantOrganizationId: "tenant-org",
		Metadata:             &cwssaws.Metadata{Name: "allocation"},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: uuid.NewString(),
			Count:          2,
		},
	}
	assert.NoError(t, manager.CreateComputeAllocationOnSite(context.Background(), createReq))

	updateReq := &cwssaws.UpdateComputeAllocationRequest{
		Id:                   id,
		TenantOrganizationId: "tenant-org",
		Metadata:             &cwssaws.Metadata{Name: "allocation"},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: createReq.Attributes.InstanceTypeId,
			Count:          3,
		},
	}
	assert.NoError(t, manager.UpdateComputeAllocationOnSite(context.Background(), updateReq))

	deleteReq := &cwssaws.DeleteComputeAllocationRequest{
		Id:                   id,
		TenantOrganizationId: "tenant-org",
	}
	assert.NoError(t, manager.DeleteComputeAllocationOnSite(context.Background(), deleteReq))
}

func TestManageAllocation_InvalidRequests(t *testing.T) {
	manager := NewManageAllocation(cClient.NewCarbideAtomicClient(&cClient.CarbideClientConfig{}))

	assert.Error(t, manager.CreateComputeAllocationOnSite(context.Background(), nil))
	assert.Error(t, manager.UpdateComputeAllocationOnSite(context.Background(), &cwssaws.UpdateComputeAllocationRequest{}))
	assert.Error(t, manager.DeleteComputeAllocationOnSite(context.Background(), &cwssaws.DeleteComputeAllocationRequest{}))
}
