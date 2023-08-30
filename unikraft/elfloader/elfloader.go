// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package elfloader

import (
	"kraftkit.sh/kconfig"
	"kraftkit.sh/pack"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft"
)

type ELFLoader struct {
	registry packmanager.PackageManager

	// Path to the kernel of the ELF loader.
	kernel string

	// The package representing the ELF Loader.
	pack pack.Package

	// The name of the elfloader.
	name string

	// The version of the elfloader.
	version string

	// The source of the elfloader (can be either remote or local, this attribute
	// is ultimately handled by the packmanager).
	source string

	// List of kconfig key-values specific to this core.
	kconfig kconfig.KeyValueMap

	// The rootfs (initramfs) of the ELF loader.
	rootfs string
}

var _ unikraft.Nameable = (*ELFLoader)(nil)

// Type implements kraftkit.sh/unikraft.Nameable
func (elfloader *ELFLoader) Type() unikraft.ComponentType {
	return unikraft.ComponentTypeApp
}

// Name implements kraftkit.sh/unikraft.Nameable
func (elfloader *ELFLoader) Name() string {
	return elfloader.name
}

// Version implements kraftkit.sh/unikraft.Nameable
func (elfloader *ELFLoader) Version() string {
	return elfloader.version
}

// Source of the ELF Loader runtime.
func (elfloader *ELFLoader) Source() string {
	return elfloader.source
}
