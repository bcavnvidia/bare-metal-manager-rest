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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	cwutil "github.com/NVIDIA/ncx-infra-controller-rest/common/pkg/util"
	cdb "github.com/NVIDIA/ncx-infra-controller-rest/db/pkg/db"
	cdbm "github.com/NVIDIA/ncx-infra-controller-rest/db/pkg/db/model"
	cdbp "github.com/NVIDIA/ncx-infra-controller-rest/db/pkg/db/paginator"
	cwssaws "github.com/NVIDIA/ncx-infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
	sc "github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/client/site"
	"github.com/NVIDIA/ncx-infra-controller-rest/workflow/pkg/queue"
	temporalClient "go.temporal.io/sdk/client"
)

// ManageComputeAllocation manages ComputeAllocation inventory reconciliation in Cloud.
type ManageComputeAllocation struct {
	dbSession      *cdb.Session
	siteClientPool *sc.ClientPool
}

// UpdateComputeAllocationsInDB reconciles ComputeAllocation inventory into the allocation tables.
func (mca ManageComputeAllocation) UpdateComputeAllocationsInDB(ctx context.Context, siteID uuid.UUID, inventory *cwssaws.ComputeAllocationInventory) error {
	logger := log.With().Str("Activity", "UpdateComputeAllocationsInDB").Str("Site ID", siteID.String()).Logger()
	logger.Info().Msg("starting activity")

	// Validate the inventory input up front.
	if inventory == nil {
		logger.Error().Msg("UpdateComputeAllocationsInDB called with nil inventory")
		return errors.New("UpdateComputeAllocationsInDB called with nil inventory")
	}

	siteDAO := cdbm.NewSiteDAO(mca.dbSession)
	tenantDAO := cdbm.NewTenantDAO(mca.dbSession)
	instanceTypeDAO := cdbm.NewInstanceTypeDAO(mca.dbSession)
	allocationDAO := cdbm.NewAllocationDAO(mca.dbSession)
	allocationConstraintDAO := cdbm.NewAllocationConstraintDAO(mca.dbSession)

	// Resolve the site and stop early for failed inventory snapshots.
	site, err := siteDAO.GetByID(ctx, nil, siteID, nil, false)
	if err != nil {
		if err == cdb.ErrDoesNotExist {
			logger.Warn().Err(err).Msg("received ComputeAllocation inventory for unknown or deleted Site")
		} else {
			logger.Error().Err(err).Msg("failed to retrieve Site from DB")
		}
		return err
	}

	if inventory.InventoryStatus == cwssaws.InventoryStatus_INVENTORY_STATUS_FAILED {
		logger.Warn().Msg("received failed inventory status from Site Agent, skipping inventory processing")
		return nil
	}

	// Load the existing instance-type allocations for the site.
	existingAllocations, _, err := allocationDAO.GetAll(ctx, nil, cdbm.AllocationFilterInput{
		SiteIDs:       []uuid.UUID{site.ID},
		ResourceTypes: []string{cdbm.AllocationResourceTypeInstanceType},
	}, cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)}, []string{cdbm.TenantRelationName})
	if err != nil {
		logger.Error().Err(err).Msg("failed to get Allocations for Site from DB")
		return err
	}

	existingAllocationMap := make(map[string]*cdbm.Allocation, len(existingAllocations))
	existingAllocationIDs := make([]uuid.UUID, 0, len(existingAllocations))
	for i := range existingAllocations {
		allocation := existingAllocations[i]
		existingAllocationMap[allocation.ID.String()] = &existingAllocations[i]
		existingAllocationIDs = append(existingAllocationIDs, allocation.ID)
	}

	constraintsByAllocationID := map[uuid.UUID]*cdbm.AllocationConstraint{}
	if len(existingAllocationIDs) > 0 {
		constraints, _, serr := allocationConstraintDAO.GetAll(
			ctx,
			nil,
			existingAllocationIDs,
			cdb.GetStrPtr(cdbm.AllocationResourceTypeInstanceType),
			nil,
			nil,
			nil,
			nil,
			nil,
			cdb.GetIntPtr(cdbp.TotalLimit),
			nil,
		)
		if serr != nil {
			logger.Error().Err(serr).Msg("failed to get Allocation Constraints for Site from DB")
			return serr
		}

		for i := range constraints {
			constraint := constraints[i]
			constraintsByAllocationID[constraint.AllocationID] = &constraints[i]
		}
	}

	tenantByOrg := map[string]*cdbm.Tenant{}
	instanceTypeByID := map[string]*cdbm.InstanceType{}
	reportedAllocationIDs := map[string]bool{}

	// Reconcile reported ComputeAllocations with existing Cloud state.
	for _, controllerAllocation := range inventory.Allocations {
		if controllerAllocation == nil || controllerAllocation.Id == nil || controllerAllocation.Id.GetValue() == "" {
			logger.Warn().Msg("skipping ComputeAllocation without a valid ID")
			continue
		}

		allocationID, serr := uuid.Parse(controllerAllocation.Id.GetValue())
		if serr != nil {
			logger.Warn().Err(serr).Str("ComputeAllocation ID", controllerAllocation.Id.GetValue()).Msg("skipping ComputeAllocation with invalid ID")
			continue
		}
		reportedAllocationIDs[allocationID.String()] = true

		existingAllocation := existingAllocationMap[allocationID.String()]
		existingConstraint := constraintsByAllocationID[allocationID]
		if existingAllocation == nil {
			logger.Warn().
				Str("ComputeAllocation ID", allocationID.String()).
				Str("Tenant Org", controllerAllocation.GetTenantOrganizationId()).
				Msg("ComputeAllocation exists on Site but not in REST DB; cannot import because it cannot be tied to a REST Tenant ID")
			continue
		}

		instanceType, serr := mca.getInstanceTypeByID(ctx, instanceTypeDAO, instanceTypeByID, controllerAllocation.GetAttributes().GetInstanceTypeId())
		if serr != nil {
			logger.Warn().Err(serr).Str("InstanceType ID", controllerAllocation.GetAttributes().GetInstanceTypeId()).Msg("skipping ComputeAllocation with unknown InstanceType")
			continue
		}

		if instanceType.SiteID == nil || *instanceType.SiteID != site.ID {
			logger.Warn().Str("ComputeAllocation ID", allocationID.String()).Msg("skipping ComputeAllocation for InstanceType not associated with inventory Site")
			continue
		}

		if serr = mca.updateComputeAllocationInCloud(ctx, tenantDAO, tenantByOrg, instanceType, existingAllocation, existingConstraint, controllerAllocation); serr != nil {
			logger.Error().Err(serr).Str("ComputeAllocation ID", allocationID.String()).Msg("failed to update ComputeAllocation in DB")
		}
	}

	// Fill the reported ID set from the paging metadata when it is present.
	if inventory.InventoryPage != nil {
		for _, itemID := range inventory.InventoryPage.ItemIds {
			reportedAllocationIDs[itemID] = true
		}
	}

	// Backfill missing site allocations once the final page has been received.
	if inventory.InventoryPage == nil || inventory.InventoryPage.TotalPages == 0 || inventory.InventoryPage.CurrentPage == inventory.InventoryPage.TotalPages {
		for _, existingAllocation := range existingAllocationMap {
			if reportedAllocationIDs[existingAllocation.ID.String()] {
				continue
			}

			// Allow for inventory lag before backfilling records that may have just been created.
			if time.Since(existingAllocation.Created) < cwutil.InventoryReceiptInterval+(5*time.Second) {
				continue
			}

			// TODO: Replace this backfill with an explicit missing-on-site state after native rollout is complete.
			if serr := mca.addComputeAllocationToSite(ctx, existingAllocation, constraintsByAllocationID[existingAllocation.ID]); serr != nil {
				logger.Error().Err(serr).Str("ComputeAllocation ID", existingAllocation.ID.String()).Msg("failed to backfill ComputeAllocation to Site")
			}
		}
	}

	return nil
}

// getTenantByOrg resolves a tenant by org and caches the lookup result.
func (mca ManageComputeAllocation) getTenantByOrg(ctx context.Context, tenantDAO cdbm.TenantDAO, cache map[string]*cdbm.Tenant, tenantOrg string) (*cdbm.Tenant, error) {
	if cachedTenant, ok := cache[tenantOrg]; ok {
		return cachedTenant, nil
	}

	tenants, err := tenantDAO.GetAllByOrg(ctx, nil, tenantOrg, nil)
	if err != nil {
		return nil, err
	}
	if len(tenants) == 0 {
		return nil, cdb.ErrDoesNotExist
	}

	cache[tenantOrg] = &tenants[0]
	return &tenants[0], nil
}

// getInstanceTypeByID resolves an InstanceType by ID and caches the lookup result.
func (mca ManageComputeAllocation) getInstanceTypeByID(ctx context.Context, instanceTypeDAO cdbm.InstanceTypeDAO, cache map[string]*cdbm.InstanceType, instanceTypeID string) (*cdbm.InstanceType, error) {
	if cachedInstanceType, ok := cache[instanceTypeID]; ok {
		return cachedInstanceType, nil
	}

	parsedInstanceTypeID, err := uuid.Parse(instanceTypeID)
	if err != nil {
		return nil, err
	}

	instanceType, err := instanceTypeDAO.GetByID(ctx, nil, parsedInstanceTypeID, nil)
	if err != nil {
		return nil, err
	}

	cache[instanceTypeID] = instanceType
	return instanceType, nil
}

// updateComputeAllocationInCloud updates existing cloud allocation rows from Site-reported data.
func (mca ManageComputeAllocation) updateComputeAllocationInCloud(ctx context.Context, tenantDAO cdbm.TenantDAO, tenantByOrg map[string]*cdbm.Tenant, instanceType *cdbm.InstanceType, existingAllocation *cdbm.Allocation, existingConstraint *cdbm.AllocationConstraint, controllerAllocation *cwssaws.ComputeAllocation) error {
	if existingConstraint == nil {
		return fmt.Errorf("missing instance-type Allocation Constraint for Allocation %s", existingAllocation.ID)
	}
	if existingAllocation.Tenant == nil {
		return fmt.Errorf("missing Tenant relation for Allocation %s", existingAllocation.ID)
	}

	updateInput := cdbm.AllocationUpdateInput{
		AllocationID: existingAllocation.ID,
	}
	allocationNeedsUpdate := false
	clearDescription := false

	if existingAllocation.Name != controllerAllocation.GetMetadata().GetName() {
		updateInput.Name = cdb.GetStrPtr(controllerAllocation.GetMetadata().GetName())
		allocationNeedsUpdate = true
	}

	if existingAllocation.Tenant.Org != controllerAllocation.GetTenantOrganizationId() {
		tenant, err := mca.getTenantByOrg(ctx, tenantDAO, tenantByOrg, controllerAllocation.GetTenantOrganizationId())
		if err != nil {
			return fmt.Errorf("failed to resolve Site-reported tenant organization %q: %w", controllerAllocation.GetTenantOrganizationId(), err)
		}
		updateInput.TenantID = &tenant.ID
		allocationNeedsUpdate = true
	}

	controllerDescription := descriptionFromMetadata(controllerAllocation.GetMetadata())
	switch {
	case existingAllocation.Description == nil && controllerDescription != nil:
		updateInput.Description = controllerDescription
		allocationNeedsUpdate = true
	case existingAllocation.Description != nil && controllerDescription == nil:
		clearDescription = true
	case existingAllocation.Description != nil && controllerDescription != nil && *existingAllocation.Description != *controllerDescription:
		updateInput.Description = controllerDescription
		allocationNeedsUpdate = true
	}

	constraintValue := int(controllerAllocation.GetAttributes().GetCount())
	constraintNeedsUpdate := existingConstraint.ResourceTypeID != instanceType.ID || existingConstraint.ConstraintValue != constraintValue

	if !allocationNeedsUpdate && !clearDescription && !constraintNeedsUpdate {
		return nil
	}

	tx, err := cdb.BeginTx(ctx, mca.dbSession, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction to update Allocation in DB: %w", err)
	}

	txCommitted := false
	defer func(dbTx *cdb.Tx, committed *bool) {
		if committed != nil && !*committed {
			dbTx.Rollback()
		}
	}(tx, &txCommitted)

	allocationDAO := cdbm.NewAllocationDAO(mca.dbSession)
	allocationConstraintDAO := cdbm.NewAllocationConstraintDAO(mca.dbSession)

	if allocationNeedsUpdate {
		if _, err = allocationDAO.Update(ctx, tx, updateInput); err != nil {
			return fmt.Errorf("failed to update Allocation in DB: %w", err)
		}
	}

	if clearDescription {
		if _, err = allocationDAO.Clear(ctx, tx, cdbm.AllocationClearInput{AllocationID: existingAllocation.ID, Description: true}); err != nil {
			return fmt.Errorf("failed to clear Allocation description in DB: %w", err)
		}
	}

	if constraintNeedsUpdate {
		_, err = allocationConstraintDAO.UpdateFromParams(
			ctx,
			tx,
			existingConstraint.ID,
			nil,
			nil,
			&instanceType.ID,
			nil,
			&constraintValue,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to update Allocation Constraint in DB: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit Allocation update transaction: %w", err)
	}

	txCommitted = true
	return nil
}

// addComputeAllocationToSite runs the site workflow to create a missing ComputeAllocation.
func (mca ManageComputeAllocation) addComputeAllocationToSite(ctx context.Context, existingAllocation *cdbm.Allocation, existingConstraint *cdbm.AllocationConstraint) error {
	if mca.siteClientPool == nil {
		return errors.New("missing Site client pool for ComputeAllocation backfill")
	}
	if existingAllocation.Tenant == nil {
		return fmt.Errorf("missing Tenant relation for Allocation %s", existingAllocation.ID)
	}
	if existingConstraint == nil {
		return fmt.Errorf("missing instance-type Allocation Constraint for Allocation %s", existingAllocation.ID)
	}

	count, err := existingConstraint.ComputeAllocationCount()
	if err != nil {
		return err
	}

	// Build the site-native create request from the existing cloud allocation.
	metadata := &cwssaws.Metadata{
		Name:        existingAllocation.Name,
		Description: "",
	}
	if existingAllocation.Description != nil {
		metadata.Description = *existingAllocation.Description
	}

	request := &cwssaws.CreateComputeAllocationRequest{
		Id:                   &cwssaws.ComputeAllocationId{Value: existingAllocation.ID.String()},
		TenantOrganizationId: existingAllocation.Tenant.Org,
		Metadata:             metadata,
		Attributes: &cwssaws.ComputeAllocationAttributes{
			InstanceTypeId: existingConstraint.ResourceTypeID.String(),
			Count:          count,
		},
		CreatedBy: cdb.GetStrPtr(existingAllocation.CreatedBy.String()),
	}

	// Execute the site workflow synchronously before considering the backfill complete.
	tc, err := mca.siteClientPool.GetClientByID(existingAllocation.SiteID)
	if err != nil {
		return fmt.Errorf("failed to retrieve Temporal client for Site: %w", err)
	}

	workflowOptions := temporalClient.StartWorkflowOptions{
		ID:                       "compute-allocation-create-" + existingAllocation.ID.String(),
		TaskQueue:                queue.SiteTaskQueue,
		WorkflowExecutionTimeout: cwutil.WorkflowExecutionTimeout,
	}

	workflowRun, err := tc.ExecuteWorkflow(ctx, workflowOptions, "CreateComputeAllocation", request)
	if err != nil {
		return fmt.Errorf("failed to start CreateComputeAllocation site workflow: %w", err)
	}

	if err = workflowRun.Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to execute CreateComputeAllocation site workflow: %w", err)
	}

	return nil
}

// descriptionFromMetadata normalizes an empty metadata description to nil for DB storage.
func descriptionFromMetadata(metadata *cwssaws.Metadata) *string {
	if metadata == nil || metadata.GetDescription() == "" {
		return nil
	}

	return cdb.GetStrPtr(metadata.GetDescription())
}

// NewManageComputeAllocation returns a new ManageComputeAllocation activity.
func NewManageComputeAllocation(dbSession *cdb.Session, siteClientPool *sc.ClientPool) ManageComputeAllocation {
	return ManageComputeAllocation{
		dbSession:      dbSession,
		siteClientPool: siteClientPool,
	}
}
