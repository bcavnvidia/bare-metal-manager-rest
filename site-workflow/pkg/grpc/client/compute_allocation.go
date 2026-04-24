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

package client

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"

	wflows "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

var (
	ErrInvalidComputeAllocationRequest = errors.New("gRPC-lib: ComputeAllocation - invalid request")
	ErrInvalidComputeAllocationID      = errors.New("gRPC-lib: ComputeAllocation - invalid allocation id")
	ErrInvalidComputeAllocationTenant  = errors.New("gRPC-lib: ComputeAllocation - invalid tenant organization id")
)

type ComputeAllocationInterface interface {
	// ComputeAllocation Interfaces
	CreateComputeAllocation(ctx context.Context, request *wflows.CreateComputeAllocationRequest) (response *wflows.CreateComputeAllocationResponse, err error)
	UpdateComputeAllocation(ctx context.Context, request *wflows.UpdateComputeAllocationRequest) (response *wflows.UpdateComputeAllocationResponse, err error)
	DeleteComputeAllocation(ctx context.Context, request *wflows.DeleteComputeAllocationRequest) (response *wflows.DeleteComputeAllocationResponse, err error)
	FindComputeAllocationIDs(ctx context.Context, request *wflows.FindComputeAllocationIdsRequest) (response *wflows.FindComputeAllocationIdsResponse, err error)
	FindComputeAllocationsByIDs(ctx context.Context, request *wflows.FindComputeAllocationsByIdsRequest) (response *wflows.FindComputeAllocationsByIdsResponse, err error)
}

// CreateComputeAllocation creates a ComputeAllocation in Carbide.
func (computeAllocation *compute) CreateComputeAllocation(ctx context.Context, request *wflows.CreateComputeAllocationRequest) (response *wflows.CreateComputeAllocationResponse, err error) {
	log.Info().Interface("request", request).Msg("CreateComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-CreateComputeAllocation")
	defer span.End()

	// Validate the request before making the gRPC call.
	switch {
	case request == nil:
		return response, ErrInvalidComputeAllocationRequest
	case request.Id == nil || request.Id.GetValue() == "":
		return response, ErrInvalidComputeAllocationID
	case request.TenantOrganizationId == "":
		return response, ErrInvalidComputeAllocationTenant
	}

	response, err = computeAllocation.carbide.CreateComputeAllocation(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("CreateComputeAllocation: error")
		return nil, err
	}

	return response, nil
}

// UpdateComputeAllocation updates a ComputeAllocation in Carbide.
func (computeAllocation *compute) UpdateComputeAllocation(ctx context.Context, request *wflows.UpdateComputeAllocationRequest) (response *wflows.UpdateComputeAllocationResponse, err error) {
	log.Info().Interface("request", request).Msg("UpdateComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-UpdateComputeAllocation")
	defer span.End()

	// Validate the request before making the gRPC call.
	switch {
	case request == nil:
		return response, ErrInvalidComputeAllocationRequest
	case request.Id == nil || request.Id.GetValue() == "":
		return response, ErrInvalidComputeAllocationID
	case request.TenantOrganizationId == "":
		return response, ErrInvalidComputeAllocationTenant
	}

	response, err = computeAllocation.carbide.UpdateComputeAllocation(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("UpdateComputeAllocation: error")
		return nil, err
	}

	return response, nil
}

// DeleteComputeAllocation deletes a ComputeAllocation in Carbide.
func (computeAllocation *compute) DeleteComputeAllocation(ctx context.Context, request *wflows.DeleteComputeAllocationRequest) (response *wflows.DeleteComputeAllocationResponse, err error) {
	log.Info().Interface("request", request).Msg("DeleteComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-DeleteComputeAllocation")
	defer span.End()

	// Validate the request before making the gRPC call.
	switch {
	case request == nil:
		return response, ErrInvalidComputeAllocationRequest
	case request.Id == nil || request.Id.GetValue() == "":
		return response, ErrInvalidComputeAllocationID
	case request.TenantOrganizationId == "":
		return response, ErrInvalidComputeAllocationTenant
	}

	response, err = computeAllocation.carbide.DeleteComputeAllocation(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("DeleteComputeAllocation: error")
		return nil, err
	}

	return response, nil
}

// FindComputeAllocationIDs finds ComputeAllocation IDs in Carbide.
func (computeAllocation *compute) FindComputeAllocationIDs(ctx context.Context, request *wflows.FindComputeAllocationIdsRequest) (response *wflows.FindComputeAllocationIdsResponse, err error) {
	log.Info().Interface("request", request).Msg("FindComputeAllocationIDs: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-FindComputeAllocationIDs")
	defer span.End()

	if request == nil {
		request = &wflows.FindComputeAllocationIdsRequest{}
	}

	response, err = computeAllocation.carbide.FindComputeAllocationIds(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("FindComputeAllocationIDs: error")
		return nil, err
	}

	return response, nil
}

// FindComputeAllocationsByIDs finds ComputeAllocations by ID in Carbide.
func (computeAllocation *compute) FindComputeAllocationsByIDs(ctx context.Context, request *wflows.FindComputeAllocationsByIdsRequest) (response *wflows.FindComputeAllocationsByIdsResponse, err error) {
	log.Info().Interface("request", request).Msg("FindComputeAllocationsByIDs: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-FindComputeAllocationsByIDs")
	defer span.End()

	if request == nil {
		request = &wflows.FindComputeAllocationsByIdsRequest{}
	}

	response, err = computeAllocation.carbide.FindComputeAllocationsByIds(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("FindComputeAllocationsByIDs: error")
		return nil, err
	}

	return response, nil
}
