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
	"errors"
	"testing"

	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	cwm "github.com/NVIDIA/ncx-infra-controller-rest/workflow/internal/metrics"
	computeAllocationActivity "github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/activity/computeallocation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

type UpdateComputeAllocationTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UpdateComputeAllocationTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UpdateComputeAllocationTestSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

func (s *UpdateComputeAllocationTestSuite) TestUpdateComputeAllocationInventorySuccess() {
	var manager computeAllocationActivity.ManageComputeAllocation
	var inventoryMetricsManager cwm.ManageInventoryMetrics

	siteID := uuid.New()
	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations: []*cwssaws.ComputeAllocation{
			{
				Id: &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
			},
		},
	}

	s.env.RegisterActivity(manager.UpdateComputeAllocationsInDB)
	s.env.OnActivity(manager.UpdateComputeAllocationsInDB, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterActivity(inventoryMetricsManager.RecordLatency)
	s.env.OnActivity(inventoryMetricsManager.RecordLatency, mock.Anything, siteID, "UpdateComputeAllocationInventory", false, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(UpdateComputeAllocationInventory, siteID.String(), inventory)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UpdateComputeAllocationTestSuite) TestUpdateComputeAllocationInventoryActivityFails() {
	var manager computeAllocationActivity.ManageComputeAllocation
	var inventoryMetricsManager cwm.ManageInventoryMetrics

	siteID := uuid.New()
	inventory := &cwssaws.ComputeAllocationInventory{
		Allocations: []*cwssaws.ComputeAllocation{
			{
				Id: &cwssaws.ComputeAllocationId{Value: uuid.NewString()},
			},
		},
	}

	s.env.RegisterActivity(manager.UpdateComputeAllocationsInDB)
	s.env.OnActivity(manager.UpdateComputeAllocationsInDB, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("UpdateComputeAllocationInventory failure"))

	s.env.RegisterActivity(inventoryMetricsManager.RecordLatency)
	s.env.OnActivity(inventoryMetricsManager.RecordLatency, mock.Anything, siteID, "UpdateComputeAllocationInventory", true, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(UpdateComputeAllocationInventory, siteID.String(), inventory)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)

	var applicationErr *temporal.ApplicationError
	s.True(errors.As(err, &applicationErr))
	s.Equal("UpdateComputeAllocationInventory failure", applicationErr.Error())
}

func TestUpdateComputeAllocationSuite(t *testing.T) {
	suite.Run(t, new(UpdateComputeAllocationTestSuite))
}
