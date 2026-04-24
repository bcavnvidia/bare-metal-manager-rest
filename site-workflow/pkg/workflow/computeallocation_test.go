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
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	iActivity "github.com/NVIDIA/ncx-infra-controller-rest/site-workflow/pkg/activity"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

type ComputeAllocationWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *ComputeAllocationWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *ComputeAllocationWorkflowTestSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

func (s *ComputeAllocationWorkflowTestSuite) TestCreateComputeAllocationSuccess() {
	var manager iActivity.ManageComputeAllocation
	request := &cwssaws.CreateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "tenant-org",
	}

	s.env.RegisterActivity(manager.CreateComputeAllocationOnSite)
	s.env.OnActivity(manager.CreateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(CreateComputeAllocation, request)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *ComputeAllocationWorkflowTestSuite) TestCreateComputeAllocationFailure() {
	var manager iActivity.ManageComputeAllocation
	request := &cwssaws.CreateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "tenant-org",
	}

	s.env.RegisterActivity(manager.CreateComputeAllocationOnSite)
	s.env.OnActivity(manager.CreateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(errors.New("create failed"))

	s.env.ExecuteWorkflow(CreateComputeAllocation, request)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *ComputeAllocationWorkflowTestSuite) TestUpdateComputeAllocationSuccess() {
	var manager iActivity.ManageComputeAllocation
	request := &cwssaws.UpdateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "tenant-org",
	}

	s.env.RegisterActivity(manager.UpdateComputeAllocationOnSite)
	s.env.OnActivity(manager.UpdateComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(UpdateComputeAllocation, request)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *ComputeAllocationWorkflowTestSuite) TestDeleteComputeAllocationSuccess() {
	var manager iActivity.ManageComputeAllocation
	request := &cwssaws.DeleteComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
		TenantOrganizationId: "tenant-org",
	}

	s.env.RegisterActivity(manager.DeleteComputeAllocationOnSite)
	s.env.OnActivity(manager.DeleteComputeAllocationOnSite, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(DeleteComputeAllocation, request)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *ComputeAllocationWorkflowTestSuite) TestDiscoverComputeAllocationInventorySuccess() {
	var manager iActivity.ManageComputeAllocationInventory

	s.env.RegisterActivity(manager.DiscoverComputeAllocationInventory)
	s.env.OnActivity(manager.DiscoverComputeAllocationInventory, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(DiscoverComputeAllocationInventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestComputeAllocationWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(ComputeAllocationWorkflowTestSuite))
}
