/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
	allocationActivity "github.com/nvidia/bare-metal-manager-rest/workflow/pkg/activity/allocation"
)

type UpdateComputeAllocationSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateComputeAllocationSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateComputeAllocationSuite) TestSuccess() {
	var manager allocationActivity.ManageAllocation
	inventory := &cwssaws.ComputeAllocationInventory{}
	siteID := uuid.New()

	s.env.OnActivity(manager.UpdateAllocationsInDB, mock.Anything, siteID, inventory).Return(nil)
	s.env.ExecuteWorkflow(UpdateComputeAllocationInventory, siteID.String(), inventory)
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateComputeAllocationSuite) TestFailure() {
	var manager allocationActivity.ManageAllocation
	inventory := &cwssaws.ComputeAllocationInventory{}
	siteID := uuid.New()

	s.env.OnActivity(manager.UpdateAllocationsInDB, mock.Anything, siteID, inventory).Return(errors.New("boom"))
	s.env.ExecuteWorkflow(UpdateComputeAllocationInventory, siteID.String(), inventory)
	s.Error(s.env.GetWorkflowError())
}

func TestUpdateComputeAllocationSuite(t *testing.T) {
	suite.Run(t, new(UpdateComputeAllocationSuite))
}
