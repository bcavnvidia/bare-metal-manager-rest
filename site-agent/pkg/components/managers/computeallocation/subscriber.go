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

package computeallocation

import (
	swa "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/activity"
	sww "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/workflow"
)

// RegisterSubscriber registers the ComputeAllocation workflows and activities with the Temporal client.
func (api *API) RegisterSubscriber() error {
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: Registering the subscribers")

	computeAllocationManager := swa.NewManageComputeAllocation(ManagerAccess.Data.EB.Managers.Carbide.Client)

	// Register the sync workflows.
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.CreateComputeAllocation)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Create ComputeAllocation workflow")

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.UpdateComputeAllocation)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Update ComputeAllocation workflow")

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflow(sww.DeleteComputeAllocation)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Delete ComputeAllocation workflow")

	// Register the sync activities.
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(computeAllocationManager.CreateComputeAllocationOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Create ComputeAllocation activity")

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(computeAllocationManager.UpdateComputeAllocationOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Update ComputeAllocation activity")

	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivity(computeAllocationManager.DeleteComputeAllocationOnSite)
	ManagerAccess.Data.EB.Log.Info().Msg("ComputeAllocation: successfully registered Delete ComputeAllocation activity")

	return nil
}
