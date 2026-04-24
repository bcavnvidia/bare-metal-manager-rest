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

	"github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/activity"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

// CreateComputeAllocation creates a ComputeAllocation using the Site Controller activity.
func CreateComputeAllocation(ctx workflow.Context, request *cwssaws.CreateComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Create").Str("ComputeAllocation ID", request.GetId().GetValue()).Logger()
	logger.Info().Msg("Starting workflow")

	// Configure retries for the Site Controller activity.
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var computeAllocationManager activity.ManageComputeAllocation

	// Execute the Site Controller activity.
	err := workflow.ExecuteActivity(ctx, computeAllocationManager.CreateComputeAllocationOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "CreateComputeAllocationOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")
	return nil
}

// UpdateComputeAllocation updates a ComputeAllocation using the Site Controller activity.
func UpdateComputeAllocation(ctx workflow.Context, request *cwssaws.UpdateComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Update").Str("ComputeAllocation ID", request.GetId().GetValue()).Logger()
	logger.Info().Msg("Starting workflow")

	// Configure retries for the Site Controller activity.
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var computeAllocationManager activity.ManageComputeAllocation

	// Execute the Site Controller activity.
	err := workflow.ExecuteActivity(ctx, computeAllocationManager.UpdateComputeAllocationOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "UpdateComputeAllocationOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")
	return nil
}

// DeleteComputeAllocation deletes a ComputeAllocation using the Site Controller activity.
func DeleteComputeAllocation(ctx workflow.Context, request *cwssaws.DeleteComputeAllocationRequest) error {
	logger := log.With().Str("Workflow", "ComputeAllocation").Str("Action", "Delete").Str("ComputeAllocation ID", request.GetId().GetValue()).Logger()
	logger.Info().Msg("Starting workflow")

	// Configure retries for the Site Controller activity.
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var computeAllocationManager activity.ManageComputeAllocation

	// Execute the Site Controller activity.
	err := workflow.ExecuteActivity(ctx, computeAllocationManager.DeleteComputeAllocationOnSite, request).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "DeleteComputeAllocationOnSite").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")
	return nil
}

// DiscoverComputeAllocationInventory discovers ComputeAllocation inventory and publishes it to Cloud.
func DiscoverComputeAllocationInventory(ctx workflow.Context) error {
	logger := log.With().Str("Workflow", "DiscoverComputeAllocationInventory").Logger()
	logger.Info().Msg("Starting workflow")

	// Configure retries for inventory collection.
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    2,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var computeAllocationInventoryManager activity.ManageComputeAllocationInventory

	// Execute the inventory collection activity.
	err := workflow.ExecuteActivity(ctx, computeAllocationInventoryManager.DiscoverComputeAllocationInventory).Get(ctx, nil)
	if err != nil {
		logger.Error().Err(err).Str("Activity", "DiscoverComputeAllocationInventory").Msg("Failed to execute activity from workflow")
		return err
	}

	logger.Info().Msg("Completing workflow")
	return nil
}
