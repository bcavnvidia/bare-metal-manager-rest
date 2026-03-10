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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"

	swe "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/error"
	cClient "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type ManageAllocation struct {
	CarbideAtomicClient *cClient.CarbideAtomicClient
}

func (ma *ManageAllocation) CreateComputeAllocationOnSite(ctx context.Context, request *cwssaws.CreateComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "CreateComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty create ComputeAllocation request", swe.ErrTypeInvalidRequest, nil)
	case request.Id != nil && request.Id.Value == "":
		return temporal.NewNonRetryableApplicationError("received create ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, nil)
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received create ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, nil)
	case request.Attributes == nil || request.Attributes.InstanceTypeId == "":
		return temporal.NewNonRetryableApplicationError("received create ComputeAllocation request without attributes.instance_type_id", swe.ErrTypeInvalidRequest, nil)
	}

	_, err := ma.CarbideAtomicClient.GetClient().Compute().CreateComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

func (ma *ManageAllocation) UpdateComputeAllocationOnSite(ctx context.Context, request *cwssaws.UpdateComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "UpdateComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty update ComputeAllocation request", swe.ErrTypeInvalidRequest, nil)
	case request.Id == nil || request.Id.Value == "":
		return temporal.NewNonRetryableApplicationError("received update ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, nil)
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received update ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, nil)
	case request.Attributes == nil || request.Attributes.InstanceTypeId == "":
		return temporal.NewNonRetryableApplicationError("received update ComputeAllocation request without attributes.instance_type_id", swe.ErrTypeInvalidRequest, nil)
	}

	_, err := ma.CarbideAtomicClient.GetClient().Compute().UpdateComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

func (ma *ManageAllocation) DeleteComputeAllocationOnSite(ctx context.Context, request *cwssaws.DeleteComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "DeleteComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty delete ComputeAllocation request", swe.ErrTypeInvalidRequest, nil)
	case request.Id == nil || request.Id.Value == "":
		return temporal.NewNonRetryableApplicationError("received delete ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, nil)
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received delete ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, nil)
	}

	_, err := ma.CarbideAtomicClient.GetClient().Compute().DeleteComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

func NewManageAllocation(carbideClient *cClient.CarbideAtomicClient) ManageAllocation {
	return ManageAllocation{CarbideAtomicClient: carbideClient}
}

type ManageAllocationInventory struct {
	config ManageInventoryConfig
}

func (mai *ManageAllocationInventory) DiscoverComputeAllocationInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverComputeAllocationInventory").Logger()
	logger.Info().Msg("Starting activity")

	inventoryImpl := manageInventoryImpl[*cwssaws.ComputeAllocationId, *cwssaws.ComputeAllocation, *cwssaws.ComputeAllocationInventory]{
		itemType:               "ComputeAllocation",
		config:                 mai.config,
		internalFindIDs:        computeAllocationFindIDs,
		internalFindByIDs:      computeAllocationFindByIDs,
		internalPagedInventory: computeAllocationPagedInventory,
	}
	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

func NewManageAllocationInventory(config ManageInventoryConfig) ManageAllocationInventory {
	return ManageAllocationInventory{config: config}
}

func computeAllocationFindIDs(ctx context.Context, carbideClient *cClient.CarbideClient) ([]*cwssaws.ComputeAllocationId, error) {
	resp, err := carbideClient.Compute().FindComputeAllocationIDs(ctx, &cwssaws.FindComputeAllocationIdsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetIds(), nil
}

func computeAllocationFindByIDs(ctx context.Context, carbideClient *cClient.CarbideClient, ids []*cwssaws.ComputeAllocationId) ([]*cwssaws.ComputeAllocation, error) {
	resp, err := carbideClient.Compute().FindComputeAllocationsByIDs(ctx, &cwssaws.FindComputeAllocationsByIdsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	return resp.GetAllocations(), nil
}

func computeAllocationPagedInventory(allIDs []*cwssaws.ComputeAllocationId, allocations []*cwssaws.ComputeAllocation, pagedInput *pagedInventoryInput) *cwssaws.ComputeAllocationInventory {
	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations:     allocations,
		Timestamp:       timestamppb.New(time.Now()),
		InventoryStatus: pagedInput.status,
		StatusMsg:       pagedInput.statusMessage,
		InventoryPage:   pagedInput.buildPage(),
	}
	if inventory.InventoryPage != nil {
		for _, id := range allIDs {
			if id != nil {
				inventory.InventoryPage.ItemIds = append(inventory.InventoryPage.ItemIds, id.Value)
			}
		}
	}
	return inventory
}
