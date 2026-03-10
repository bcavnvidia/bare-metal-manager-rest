/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import "fmt"

func (api *API) Init() {
	ManagerAccess.Data.EB.Log.Info().Msg("Allocation: Initializing")
}

func (api *API) GetState() []string {
	state := ManagerAccess.Data.EB.Managers.Workflow.AllocationState
	return []string{
		fmt.Sprintln("allocation_workflow_started", state.WflowStarted.Load()),
		fmt.Sprintln("allocation_workflow_activity_failed", state.WflowActFail.Load()),
		fmt.Sprintln("allocation_workflow_activity_succeeded", state.WflowActSucc.Load()),
		fmt.Sprintln("allocation_workflow_publishing_failed", state.WflowPubFail.Load()),
		fmt.Sprintln("allocation_workflow_publishing_succeeded", state.WflowPubSucc.Load()),
	}
}
