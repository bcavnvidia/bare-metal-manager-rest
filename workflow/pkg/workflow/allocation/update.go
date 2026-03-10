/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	"fmt"
	"time"

	cwm "github.com/nvidia/bare-metal-manager-rest/workflow/internal/metrics"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
	allocationActivity "github.com/nvidia/bare-metal-manager-rest/workflow/pkg/activity/allocation"
)

func UpdateComputeAllocationInventory(ctx workflow.Context, siteID string, inventory *cwssaws.ComputeAllocationInventory) (err error) {
	logger := log.With().Str("Workflow", "UpdateComputeAllocationInventory").Str("Site ID", siteID).Logger()
	startTime := time.Now()

	parsedSiteID, err := uuid.Parse(siteID)
	if err != nil {
		logger.Warn().Err(err).Msg(fmt.Sprintf("workflow triggered with invalid site ID: %s", siteID))
		return err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
		},
	})

	var allocationManager allocationActivity.ManageAllocation
	err = workflow.ExecuteActivity(ctx, allocationManager.UpdateAllocationsInDB, parsedSiteID, inventory).Get(ctx, nil)

	var inventoryMetricsManager cwm.ManageInventoryMetrics
	_ = workflow.ExecuteActivity(ctx, inventoryMetricsManager.RecordLatency, parsedSiteID, "UpdateComputeAllocationInventory", err != nil, time.Since(startTime)).Get(ctx, nil)
	return err
}
