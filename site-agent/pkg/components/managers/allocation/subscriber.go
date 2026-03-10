/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	swa "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/activity"
	sww "github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/workflow"
)

func (api *API) RegisterSubscriber() error {
	manager := swa.NewManageAllocation(ManagerAccess.Data.EB.Managers.Carbide.Client)

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateComputeAllocation)
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateComputeAllocation)
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteComputeAllocation)

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(manager.CreateComputeAllocationOnSite)
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(manager.UpdateComputeAllocationOnSite)
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(manager.DeleteComputeAllocationOnSite)
	return nil
}
