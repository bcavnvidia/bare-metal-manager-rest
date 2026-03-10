/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managerapi

type AllocationExpansion interface{}

type AllocationInterface interface {
	Init()
	RegisterSubscriber() error
	RegisterPublisher() error
	RegisterCron() error
	GetState() []string
	AllocationExpansion
}
