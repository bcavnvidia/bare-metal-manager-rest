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

package workflow

import (
	"time"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/activity"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

func CreateComputeAllocation(ctx workflow.Context, request *cwssaws.CreateComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Create").Logger()
	logger.Info().Msg("Starting workflow")

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    2,
		},
	})

	var allocationManager activity.ManageAllocation
	return workflow.ExecuteActivity(ctx, allocationManager.CreateComputeAllocationOnSite, request).Get(ctx, nil)
}

func UpdateComputeAllocation(ctx workflow.Context, request *cwssaws.UpdateComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Update").Logger()
	logger.Info().Msg("Starting workflow")

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    2,
		},
	})

	var allocationManager activity.ManageAllocation
	return workflow.ExecuteActivity(ctx, allocationManager.UpdateComputeAllocationOnSite, request).Get(ctx, nil)
}

func DeleteComputeAllocation(ctx workflow.Context, request *cwssaws.DeleteComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Delete").Logger()
	logger.Info().Msg("Starting workflow")

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    2,
		},
	})

	var allocationManager activity.ManageAllocation
	return workflow.ExecuteActivity(ctx, allocationManager.DeleteComputeAllocationOnSite, request).Get(ctx, nil)
}

func DiscoverComputeAllocationInventory(ctx workflow.Context) error {
	logger := log.With().Str("Workflow", "DiscoverComputeAllocationInventory").Logger()
	logger.Info().Msg("Starting workflow")

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
		},
	})

	var inventoryManager activity.ManageAllocationInventory
	return workflow.ExecuteActivity(ctx, inventoryManager.DiscoverComputeAllocationInventory).Get(ctx, nil)
}
