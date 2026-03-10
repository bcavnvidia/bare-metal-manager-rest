/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workflow

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	"github.com/nvidia/bare-metal-manager-rest/site-workflow/pkg/activity"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type ComputeAllocationWorkflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *ComputeAllocationWorkflowSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *ComputeAllocationWorkflowSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

func (s *ComputeAllocationWorkflowSuite) TestCRUDAndInventory() {
	var manager activity.ManageAllocation
	var inventory activity.ManageAllocationInventory
	id := &cwssaws.ComputeAllocationId{Value: uuid.NewString()}

	createReq := &cwssaws.CreateComputeAllocationRequest{
		Id:                   id,
		TenantOrganizationId: "tenant-org",
		Metadata:             &cwssaws.Metadata{Name: "allocation"},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: uuid.NewString(),
			Count:          1,
		},
	}

	s.env.RegisterActivity(manager.CreateComputeAllocationOnSite)
	s.env.OnActivity(manager.CreateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)
	s.env.ExecuteWorkflow(CreateComputeAllocation, createReq)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(manager.UpdateComputeAllocationOnSite)
	s.env.OnActivity(manager.UpdateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)
	s.env.ExecuteWorkflow(UpdateComputeAllocation, &cwssaws.UpdateComputeAllocationRequest{
		Id:                   id,
		TenantOrganizationId: "tenant-org",
		Metadata:             &cwssaws.Metadata{Name: "allocation"},
		Attributes:           createReq.Attributes,
	})
	s.NoError(s.env.GetWorkflowError())

	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(manager.DeleteComputeAllocationOnSite)
	s.env.OnActivity(manager.DeleteComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)
	s.env.ExecuteWorkflow(DeleteComputeAllocation, &cwssaws.DeleteComputeAllocationRequest{Id: id, TenantOrganizationId: "tenant-org"})
	s.NoError(s.env.GetWorkflowError())

	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(inventory.DiscoverComputeAllocationInventory)
	s.env.OnActivity(inventory.DiscoverComputeAllocationInventory, mock.Anything).Return(nil)
	s.env.ExecuteWorkflow(DiscoverComputeAllocationInventory)
	s.NoError(s.env.GetWorkflowError())
}

func (s *ComputeAllocationWorkflowSuite) TestCreateFailure() {
	var manager activity.ManageAllocation
	s.env.RegisterActivity(manager.CreateComputeAllocationOnSite)
	s.env.OnActivity(manager.CreateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(errors.New("boom"))
	s.env.ExecuteWorkflow(CreateComputeAllocation, &cwssaws.CreateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "tenant-org",
		Metadata:             &cwssaws.Metadata{Name: "allocation"},
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: uuid.NewString(),
			Count:          1,
		},
	})
	s.Error(s.env.GetWorkflowError())
}

func TestComputeAllocationWorkflowSuite(t *testing.T) {
	suite.Run(t, new(ComputeAllocationWorkflowSuite))
}
