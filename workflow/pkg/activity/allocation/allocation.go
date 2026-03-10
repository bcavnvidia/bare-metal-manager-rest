/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	cwutil "github.com/nvidia/bare-metal-manager-rest/common/pkg/util"
	cdb "github.com/nvidia/bare-metal-manager-rest/db/pkg/db"
	cdbm "github.com/nvidia/bare-metal-manager-rest/db/pkg/db/model"
	cdbp "github.com/nvidia/bare-metal-manager-rest/db/pkg/db/paginator"
	cwssaws "github.com/nvidia/bare-metal-manager-rest/workflow-schema/schema/site-agent/workflows/v1"
	sc "github.com/nvidia/bare-metal-manager-rest/workflow/pkg/client/site"
)

type ManageAllocation struct {
	dbSession      *cdb.Session
	siteClientPool *sc.ClientPool
}

type existingComputeAllocation struct {
	allocation *cdbm.Allocation
	constraint *cdbm.AllocationConstraint
}

func (ma ManageAllocation) UpdateAllocationsInDB(ctx context.Context, siteID uuid.UUID, inventory *cwssaws.ComputeAllocationInventory) error {
	logger := log.With().Str("Activity", "UpdateAllocationsInDB").Str("Site ID", siteID.String()).Logger()
	logger.Info().Msg("starting activity")

	if inventory == nil {
		return errors.New("UpdateAllocationsInDB called with nil inventory")
	}
	if inventory.InventoryStatus == cwssaws.InventoryStatus_INVENTORY_STATUS_FAILED {
		logger.Warn().Msg("received failed inventory status from Site Agent, skipping inventory processing")
		return nil
	}

	siteDAO := cdbm.NewSiteDAO(ma.dbSession)
	site, err := siteDAO.GetByID(ctx, nil, siteID, nil, false)
	if err != nil {
		return err
	}

	allocationDAO := cdbm.NewAllocationDAO(ma.dbSession)
	constraintDAO := cdbm.NewAllocationConstraintDAO(ma.dbSession)
	tenantDAO := cdbm.NewTenantDAO(ma.dbSession)
	instanceTypeDAO := cdbm.NewInstanceTypeDAO(ma.dbSession)
	tenantSiteDAO := cdbm.NewTenantSiteDAO(ma.dbSession)
	statusDAO := cdbm.NewStatusDetailDAO(ma.dbSession)

	dbAllocations, _, err := allocationDAO.GetAll(ctx, nil, cdbm.AllocationFilterInput{SiteIDs: []uuid.UUID{site.ID}}, cdbp.PageInput{Limit: cdb.GetIntPtr(cdbp.TotalLimit)}, nil)
	if err != nil {
		return err
	}

	allocationIDs := make([]uuid.UUID, 0, len(dbAllocations))
	allocationByID := make(map[uuid.UUID]*cdbm.Allocation, len(dbAllocations))
	for i := range dbAllocations {
		allocationIDs = append(allocationIDs, dbAllocations[i].ID)
		allocationByID[dbAllocations[i].ID] = &dbAllocations[i]
	}

	dbConstraints, _, err := constraintDAO.GetAll(ctx, nil, allocationIDs, cdb.GetStrPtr(cdbm.AllocationResourceTypeInstanceType), nil, nil, nil, nil, nil, cdb.GetIntPtr(cdbp.TotalLimit), nil)
	if err != nil {
		return err
	}

	existingByID := map[uuid.UUID]*existingComputeAllocation{}
	for i := range dbConstraints {
		constraint := &dbConstraints[i]
		allocation := allocationByID[constraint.AllocationID]
		if allocation == nil {
			continue
		}
		existingByID[allocation.ID] = &existingComputeAllocation{
			allocation: allocation,
			constraint: constraint,
		}
	}

	reportedIDs := map[uuid.UUID]bool{}
	if inventory.InventoryPage != nil {
		for _, itemID := range inventory.InventoryPage.ItemIds {
			if parsedID, parseErr := uuid.Parse(itemID); parseErr == nil {
				reportedIDs[parsedID] = true
			}
		}
	}

	for _, reported := range inventory.Allocations {
		if reported == nil || reported.Id == nil {
			continue
		}

		allocationID, parseErr := uuid.Parse(reported.Id.Value)
		if parseErr != nil {
			logger.Warn().Err(parseErr).Str("Allocation ID", reported.Id.Value).Msg("skipping ComputeAllocation with invalid ID")
			continue
		}
		reportedIDs[allocationID] = true

		tenantMatches, serr := tenantDAO.GetAllByOrg(ctx, nil, reported.TenantOrganizationId, nil)
		if serr != nil || len(tenantMatches) == 0 {
			logger.Warn().Err(serr).Str("Tenant Org", reported.TenantOrganizationId).Msg("skipping ComputeAllocation with unknown tenant")
			continue
		}
		tenant := tenantMatches[0]

		instanceTypeID, parseErr := uuid.Parse(reported.GetAttributes().GetInstanceTypeId())
		if parseErr != nil {
			logger.Warn().Err(parseErr).Str("Instance Type ID", reported.GetAttributes().GetInstanceTypeId()).Msg("skipping ComputeAllocation with invalid InstanceType ID")
			continue
		}
		instanceType, serr := instanceTypeDAO.GetByID(ctx, nil, instanceTypeID, nil)
		if serr != nil {
			logger.Warn().Err(serr).Str("Instance Type ID", instanceTypeID.String()).Msg("skipping ComputeAllocation with unknown InstanceType")
			continue
		}

		tx, serr := cdb.BeginTx(ctx, ma.dbSession, nil)
		if serr != nil {
			return serr
		}
		txCommitted := false
		defer func() {
			if !txCommitted {
				tx.Rollback()
			}
		}()

		existing := existingByID[allocationID]
		name := reported.GetMetadata().GetName()
		if name == "" {
			name = allocationID.String()
		}
		description := cdb.GetStrPtr(reported.GetMetadata().GetDescription())
		if reported.GetMetadata().GetDescription() == "" {
			description = nil
		}

		var allocation *cdbm.Allocation
		if existing == nil {
			allocation = &cdbm.Allocation{ID: allocationID}
			created, createErr := cdb.GetIDB(tx, ma.dbSession).NewInsert().Model(&cdbm.Allocation{
				ID:                       allocationID,
				Name:                     name,
				Description:              description,
				InfrastructureProviderID: site.InfrastructureProviderID,
				TenantID:                 tenant.ID,
				SiteID:                   site.ID,
				Status:                   cdbm.AllocationStatusRegistered,
				CreatedBy:                site.CreatedBy,
			}).Exec(ctx)
			_ = created
			if createErr != nil {
				return createErr
			}
			allocation, serr = allocationDAO.GetByID(ctx, tx, allocationID, nil)
			if serr != nil {
				return serr
			}
			_, serr = constraintDAO.CreateFromParams(ctx, tx, allocation.ID, cdbm.AllocationResourceTypeInstanceType, instanceType.ID, cdbm.AllocationConstraintTypeReserved, int(reported.GetAttributes().GetCount()), nil, site.CreatedBy)
			if serr != nil {
				return serr
			}
			if _, serr = tenantSiteDAO.GetByTenantIDAndSiteID(ctx, tx, tenant.ID, site.ID, nil); serr == cdb.ErrDoesNotExist {
				_, serr = tenantSiteDAO.Create(ctx, tx, cdbm.TenantSiteCreateInput{
					TenantID:  tenant.ID,
					TenantOrg: tenant.Org,
					SiteID:    site.ID,
					CreatedBy: site.CreatedBy,
				})
			}
			if serr != nil && serr != cdb.ErrDoesNotExist {
				return serr
			}
			_, serr = statusDAO.CreateFromParams(ctx, tx, allocation.ID.String(), cdbm.AllocationStatusRegistered, cdb.GetStrPtr("allocation synchronized from site inventory"))
			if serr != nil {
				return serr
			}
		} else {
			allocation = existing.allocation
			_, serr = allocationDAO.Update(ctx, tx, cdbm.AllocationUpdateInput{
				AllocationID: allocation.ID,
				Name:         &name,
				Description:  description,
			})
			if serr != nil {
				return serr
			}
			_, serr = constraintDAO.UpdateFromParams(ctx, tx, existing.constraint.ID, nil, nil, &instanceType.ID, nil, cdb.GetIntPtr(int(reported.GetAttributes().GetCount())), nil)
			if serr != nil {
				return serr
			}
		}

		if serr = tx.Commit(); serr != nil {
			return serr
		}
		txCommitted = true
	}

	if inventory.InventoryPage == nil || inventory.InventoryPage.TotalPages == 0 || inventory.InventoryPage.CurrentPage == inventory.InventoryPage.TotalPages {
		for _, existing := range existingByID {
			if reportedIDs[existing.allocation.ID] {
				continue
			}
			if time.Since(existing.allocation.Created) < cwutil.InventoryReceiptInterval+(5*time.Second) {
				continue
			}

			tx, serr := cdb.BeginTx(ctx, ma.dbSession, nil)
			if serr != nil {
				return serr
			}
			if serr = constraintDAO.DeleteByID(ctx, tx, existing.constraint.ID); serr != nil && serr != cdb.ErrDoesNotExist {
				tx.Rollback()
				return serr
			}
			if serr = allocationDAO.Delete(ctx, tx, existing.allocation.ID); serr != nil && serr != cdb.ErrDoesNotExist {
				tx.Rollback()
				return serr
			}
			if serr = tx.Commit(); serr != nil {
				return serr
			}
		}
	}

	return nil
}

func NewManageAllocation(dbSession *cdb.Session, siteClientPool *sc.ClientPool) ManageAllocation {
	return ManageAllocation{dbSession: dbSession, siteClientPool: siteClientPool}
}
