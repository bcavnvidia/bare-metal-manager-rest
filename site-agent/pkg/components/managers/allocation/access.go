/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package allocation

import (
	Manager "github.com/nvidia/bare-metal-manager-rest/site-agent/pkg/components/managers/managerapi"
	"github.com/nvidia/bare-metal-manager-rest/site-agent/pkg/datatypes/elektratypes"
)

var ManagerAccess *Manager.ManagerAccess

type API struct{}

func NewAllocationManager(superForge *elektratypes.Elektra, superAPI *Manager.ManagerAPI, superConf *Manager.ManagerConf) *API {
	ManagerAccess = &Manager.ManagerAccess{
		Data: &Manager.ManagerData{EB: superForge},
		API:  superAPI,
		Conf: superConf,
	}
	return &API{}
}
