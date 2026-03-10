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

	wflows "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

var (
	ErrInvalidComputeAllocationRequest = errors.New("gRPC-lib: ComputeAllocation - invalid request")
	ErrInvalidComputeAllocationID      = errors.New("gRPC-lib: ComputeAllocation - invalid id")
	ErrInvalidComputeAllocationTenant  = errors.New("gRPC-lib: ComputeAllocation - invalid tenant organization id")
	ErrInvalidComputeAllocationName    = errors.New("gRPC-lib: ComputeAllocation - invalid name")
	ErrInvalidComputeAllocationAttrs   = errors.New("gRPC-lib: ComputeAllocation - invalid attributes")
)

type AllocationInterface interface {
	CreateComputeAllocation(ctx context.Context, request *wflows.CreateComputeAllocationRequest) (*wflows.CreateComputeAllocationResponse, error)
	UpdateComputeAllocation(ctx context.Context, request *wflows.UpdateComputeAllocationRequest) (*wflows.UpdateComputeAllocationResponse, error)
	DeleteComputeAllocation(ctx context.Context, request *wflows.DeleteComputeAllocationRequest) (*wflows.DeleteComputeAllocationResponse, error)
	FindComputeAllocationIDs(ctx context.Context, request *wflows.FindComputeAllocationIdsRequest) (*wflows.FindComputeAllocationIdsResponse, error)
	FindComputeAllocationsByIDs(ctx context.Context, request *wflows.FindComputeAllocationsByIdsRequest) (*wflows.FindComputeAllocationsByIdsResponse, error)
}

func validateComputeAllocationMetadata(metadata *wflows.Metadata) error {
	if metadata != nil && metadata.Name == "" {
		return ErrInvalidComputeAllocationName
	}
	return nil
}

func validateComputeAllocationAttributes(attrs *wflows.ComputeAllocationAttributes) error {
	if attrs == nil || attrs.InstanceTypeId == "" {
		return ErrInvalidComputeAllocationAttrs
	}
	return nil
}

func (compute *compute) CreateComputeAllocation(ctx context.Context, request *wflows.CreateComputeAllocationRequest) (*wflows.CreateComputeAllocationResponse, error) {
	log.Info().Interface("request", request).Msg("CreateComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-CreateComputeAllocation")
	defer span.End()

	if request == nil {
		return nil, ErrInvalidComputeAllocationRequest
	}
	if request.Id != nil && request.Id.Value == "" {
		return nil, ErrInvalidComputeAllocationID
	}
	if request.TenantOrganizationId == "" {
		return nil, ErrInvalidComputeAllocationTenant
	}
	if err := validateComputeAllocationMetadata(request.Metadata); err != nil {
		return nil, err
	}
	if err := validateComputeAllocationAttributes(request.Attributes); err != nil {
		return nil, err
	}

	return compute.carbide.CreateComputeAllocation(ctx, request)
}

func (compute *compute) UpdateComputeAllocation(ctx context.Context, request *wflows.UpdateComputeAllocationRequest) (*wflows.UpdateComputeAllocationResponse, error) {
	log.Info().Interface("request", request).Msg("UpdateComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-UpdateComputeAllocation")
	defer span.End()

	if request == nil {
		return nil, ErrInvalidComputeAllocationRequest
	}
	if request.Id == nil || request.Id.Value == "" {
		return nil, ErrInvalidComputeAllocationID
	}
	if request.TenantOrganizationId == "" {
		return nil, ErrInvalidComputeAllocationTenant
	}
	if err := validateComputeAllocationMetadata(request.Metadata); err != nil {
		return nil, err
	}
	if err := validateComputeAllocationAttributes(request.Attributes); err != nil {
		return nil, err
	}

	return compute.carbide.UpdateComputeAllocation(ctx, request)
}

func (compute *compute) DeleteComputeAllocation(ctx context.Context, request *wflows.DeleteComputeAllocationRequest) (*wflows.DeleteComputeAllocationResponse, error) {
	log.Info().Interface("request", request).Msg("DeleteComputeAllocation: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-DeleteComputeAllocation")
	defer span.End()

	if request == nil {
		return nil, ErrInvalidComputeAllocationRequest
	}
	if request.Id == nil || request.Id.Value == "" {
		return nil, ErrInvalidComputeAllocationID
	}
	if request.TenantOrganizationId == "" {
		return nil, ErrInvalidComputeAllocationTenant
	}

	return compute.carbide.DeleteComputeAllocation(ctx, request)
}

func (compute *compute) FindComputeAllocationIDs(ctx context.Context, request *wflows.FindComputeAllocationIdsRequest) (*wflows.FindComputeAllocationIdsResponse, error) {
	log.Info().Interface("request", request).Msg("FindComputeAllocationIDs: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-FindComputeAllocationIDs")
	defer span.End()

	if request == nil {
		request = &wflows.FindComputeAllocationIdsRequest{}
	}

	return compute.carbide.FindComputeAllocationIds(ctx, request)
}

func (compute *compute) FindComputeAllocationsByIDs(ctx context.Context, request *wflows.FindComputeAllocationsByIdsRequest) (*wflows.FindComputeAllocationsByIdsResponse, error) {
	log.Info().Interface("request", request).Msg("FindComputeAllocationsByIDs: received request")
	ctx, span := otel.Tracer(os.Getenv("LS_SERVICE_NAME")).Start(ctx, "CarbideClient-FindComputeAllocationsByIDs")
	defer span.End()

	if request == nil {
		request = &wflows.FindComputeAllocationsByIdsRequest{}
	}

	return compute.carbide.FindComputeAllocationsByIds(ctx, request)
}
