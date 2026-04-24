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
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/timestamppb"

	swe "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/error"
	cClient "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/grpc/client"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// ManageComputeAllocation is an activity wrapper for ComputeAllocation management tasks.
type ManageComputeAllocation struct {
	CarbideAtomicClient *cClient.CarbideAtomicClient
}

// CreateComputeAllocationOnSite creates a ComputeAllocation on the Site Controller.
func (mca *ManageComputeAllocation) CreateComputeAllocationOnSite(ctx context.Context, request *cwssaws.CreateComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "CreateComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	// Validate the request before calling the Site Controller.
	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty create ComputeAllocation request", swe.ErrTypeInvalidRequest, errors.New("received empty create ComputeAllocation request"))
	case request.Id == nil || request.Id.GetValue() == "":
		return temporal.NewNonRetryableApplicationError("received create ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, errors.New("received create ComputeAllocation request without ID"))
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received create ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, errors.New("received create ComputeAllocation request without tenant organization ID"))
	}

	// Call the Site Controller gRPC endpoint.
	carbideClient := mca.CarbideAtomicClient.GetClient()
	if carbideClient == nil {
		return cClient.ErrClientNotConnected
	}

	_, err := carbideClient.Compute().CreateComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

// UpdateComputeAllocationOnSite updates a ComputeAllocation on the Site Controller.
func (mca *ManageComputeAllocation) UpdateComputeAllocationOnSite(ctx context.Context, request *cwssaws.UpdateComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "UpdateComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	// Validate the request before calling the Site Controller.
	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty update ComputeAllocation request", swe.ErrTypeInvalidRequest, errors.New("received empty update ComputeAllocation request"))
	case request.Id == nil || request.Id.GetValue() == "":
		return temporal.NewNonRetryableApplicationError("received update ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, errors.New("received update ComputeAllocation request without ID"))
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received update ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, errors.New("received update ComputeAllocation request without tenant organization ID"))
	}

	// Call the Site Controller gRPC endpoint.
	carbideClient := mca.CarbideAtomicClient.GetClient()
	if carbideClient == nil {
		return cClient.ErrClientNotConnected
	}

	_, err := carbideClient.Compute().UpdateComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to update ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

// DeleteComputeAllocationOnSite deletes a ComputeAllocation on the Site Controller.
func (mca *ManageComputeAllocation) DeleteComputeAllocationOnSite(ctx context.Context, request *cwssaws.DeleteComputeAllocationRequest) error {
	logger := log.With().Str("Activity", "DeleteComputeAllocationOnSite").Logger()
	logger.Info().Msg("Starting activity")

	// Validate the request before calling the Site Controller.
	switch {
	case request == nil:
		return temporal.NewNonRetryableApplicationError("received empty delete ComputeAllocation request", swe.ErrTypeInvalidRequest, errors.New("received empty delete ComputeAllocation request"))
	case request.Id == nil || request.Id.GetValue() == "":
		return temporal.NewNonRetryableApplicationError("received delete ComputeAllocation request without ID", swe.ErrTypeInvalidRequest, errors.New("received delete ComputeAllocation request without ID"))
	case request.TenantOrganizationId == "":
		return temporal.NewNonRetryableApplicationError("received delete ComputeAllocation request without tenant organization ID", swe.ErrTypeInvalidRequest, errors.New("received delete ComputeAllocation request without tenant organization ID"))
	}

	// Call the Site Controller gRPC endpoint.
	carbideClient := mca.CarbideAtomicClient.GetClient()
	if carbideClient == nil {
		return cClient.ErrClientNotConnected
	}

	_, err := carbideClient.Compute().DeleteComputeAllocation(ctx, request)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to delete ComputeAllocation using Site Controller API")
		return swe.WrapErr(err)
	}

	logger.Info().Msg("Completed activity")
	return nil
}

// NewManageComputeAllocation returns a new ManageComputeAllocation activity.
func NewManageComputeAllocation(carbideClient *cClient.CarbideAtomicClient) ManageComputeAllocation {
	return ManageComputeAllocation{
		CarbideAtomicClient: carbideClient,
	}
}

// ManageComputeAllocationInventory is an activity wrapper for ComputeAllocation inventory collection and publishing.
type ManageComputeAllocationInventory struct {
	config ManageInventoryConfig
}

// DiscoverComputeAllocationInventory collects ComputeAllocation inventory and publishes it to Cloud.
func (mcai *ManageComputeAllocationInventory) DiscoverComputeAllocationInventory(ctx context.Context) error {
	logger := log.With().Str("Activity", "DiscoverComputeAllocationInventory").Logger()
	logger.Info().Msg("Starting activity")

	inventoryImpl := manageInventoryImpl[*cwssaws.ComputeAllocationId, *cwssaws.ComputeAllocation, *cwssaws.ComputeAllocationInventory]{
		itemType:               "ComputeAllocation",
		config:                 mcai.config,
		internalFindIDs:        computeAllocationFindIDs,
		internalFindByIDs:      computeAllocationFindByIDs,
		internalPagedInventory: computeAllocationPagedInventory,
	}

	return inventoryImpl.CollectAndPublishInventory(ctx, &logger)
}

// NewManageComputeAllocationInventory returns a ManageInventory implementation for ComputeAllocation activity.
func NewManageComputeAllocationInventory(config ManageInventoryConfig) ManageComputeAllocationInventory {
	return ManageComputeAllocationInventory{
		config: config,
	}
}

// computeAllocationFindIDs retrieves the ComputeAllocation IDs from Carbide.
func computeAllocationFindIDs(ctx context.Context, carbideClient *cClient.CarbideClient) ([]*cwssaws.ComputeAllocationId, error) {
	idList, err := carbideClient.Compute().FindComputeAllocationIDs(ctx, &cwssaws.FindComputeAllocationIdsRequest{})
	if err != nil {
		return nil, err
	}

	return idList.GetIds(), nil
}

// computeAllocationFindByIDs retrieves the ComputeAllocation resources for the given IDs from Carbide.
func computeAllocationFindByIDs(ctx context.Context, carbideClient *cClient.CarbideClient, ids []*cwssaws.ComputeAllocationId) ([]*cwssaws.ComputeAllocation, error) {
	list, err := carbideClient.Compute().FindComputeAllocationsByIDs(ctx, &cwssaws.FindComputeAllocationsByIdsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	return list.GetAllocations(), nil
}

// computeAllocationPagedInventory builds a paged inventory payload for ComputeAllocation inventory publishing.
func computeAllocationPagedInventory(allItemIDs []*cwssaws.ComputeAllocationId, pagedItems []*cwssaws.ComputeAllocation, input *pagedInventoryInput) *cwssaws.ComputeAllocationInventory {
	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations: pagedItems,
		Timestamp: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		InventoryStatus: input.status,
		StatusMsg:       input.statusMessage,
		InventoryPage:   input.buildPage(),
	}

	if inventory.InventoryPage != nil {
		for _, id := range allItemIDs {
			if id != nil {
				inventory.InventoryPage.ItemIds = append(inventory.InventoryPage.ItemIds, id.GetValue())
			}
		}
	}

	return inventory
}
