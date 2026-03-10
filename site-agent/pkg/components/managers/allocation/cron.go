/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	"context"

	"go.temporal.io/sdk/client"

	sww "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/workflow"
)

const (
	InventoryCarbidePageSize = 100
	InventoryCloudPageSize   = 25
	InventoryDefaultSchedule = "@every 3m"
)

func (api *API) RegisterCron() error {
	workflowOptions := client.StartWorkflowOptions{
		ID:           "inventory-allocation-" + ManagerAccess.Conf.EB.Temporal.TemporalSubscribeNamespace,
		TaskQueue:    ManagerAccess.Conf.EB.Temporal.TemporalSubscribeQueue,
		CronSchedule: InventoryDefaultSchedule,
	}
	if ManagerAccess.Conf.EB.Temporal.TemporalInventorySchedule != "" {
		workflowOptions.CronSchedule = ManagerAccess.Conf.EB.Temporal.TemporalInventorySchedule
	}
	_, err := ManagerAccess.Data.EB.Managers.Workflow.Temporal.Subscriber.ExecuteWorkflow(context.Background(), workflowOptions, sww.DiscoverComputeAllocationInventory)
	return err
}
