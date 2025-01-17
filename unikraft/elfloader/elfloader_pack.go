// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package elfloader

import (
	"context"
	"fmt"

	"kraftkit.sh/initrd"
	"kraftkit.sh/oci"
	"kraftkit.sh/pack"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/unikraft/arch"
	"kraftkit.sh/unikraft/plat"
	"kraftkit.sh/unikraft/target"
)

var _ pack.Package = (*ELFLoader)(nil)

const (
	PrebuiltRegistry = "loaders.unikraft.org"
	DefaultPrebuilt  = "loaders.unikraft.org/default:latest"
)

// NewELFLoaderFromPrebuilt prepares a ELF Loader application that has been
// pre-built and is accessible from a remote registry.
func NewELFLoaderFromPrebuilt(ctx context.Context, pbopts ...ELFLoaderPrebuiltOption) (*ELFLoader, error) {
	elfloader := ELFLoader{}

	for _, opt := range pbopts {
		if err := opt(&elfloader); err != nil {
			return nil, err
		}
	}

	if defaultPrebuilt != "" {
		// Return early if the user provided a custom elfloader unikernel
		// application.
		if ok, _ := unikraft.IsFileUnikraftUnikernel(defaultPrebuilt); ok {
			elfloader.kernel = defaultPrebuilt
			return &elfloader, nil
		}

		elfloader.source = defaultPrebuilt
	} else if len(elfloader.kernel) > 0 {
		return &elfloader, nil
	} else if len(elfloader.name) == 0 {
		elfloader.name = DefaultPrebuilt
	}

	var err error
	elfloader.registry = packmanager.G(ctx)
	if elfloader.registry == nil {
		elfloader.registry, err = oci.NewOCIManager(ctx,
			oci.WithDetectHandler(),
		)
	} else {
		elfloader.registry, err = elfloader.registry.From(oci.OCIFormat)
	}
	if err != nil {
		return nil, err
	}

	if err := elfloader.registry.SetSources(ctx, PrebuiltRegistry); err != nil {
		return nil, err
	}

	// First try locally
	results, err := elfloader.registry.Catalog(ctx,
		packmanager.WithName(elfloader.name),
		packmanager.WithTypes(unikraft.ComponentTypeApp),
		packmanager.WithCache(true),
	)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		results, err = elfloader.registry.Catalog(ctx,
			packmanager.WithName(elfloader.name),
			packmanager.WithTypes(unikraft.ComponentTypeApp),
			packmanager.WithCache(false),
		)
		if err != nil {
			return nil, err
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("could not find elfloader")
	} else if len(results) > 1 {
		options := make([]string, len(results))
		for i, result := range results {
			options[i] = result.Name()
		}
		return nil, fmt.Errorf("too many options: %v", options)
	}

	elfloader.pack = results[0]
	elfloader.name = results[0].Name()
	elfloader.version = results[0].Version()
	elfloader.source = results[0].Name()

	return &elfloader, nil
}

// Metadata implements kraftkit.sh/pack.Package
func (elfloader *ELFLoader) Metadata() any {
	return elfloader.pack.Metadata()
}

// Push implements kraftkit.sh/pack.Package
func (elfloader *ELFLoader) Push(ctx context.Context, opts ...pack.PushOption) error {
	panic("not implemented: kraftkit.sh/unikraft/elfloader.ELFLoader.Push")
}

// Pull implements kraftkit.sh/pack.Package
func (elfloader *ELFLoader) Pull(ctx context.Context, opts ...pack.PullOption) error {
	return elfloader.pack.Pull(ctx, opts...)
}

// Format implements kraftkit.sh/unikraft.component.Component
func (elfloader *ELFLoader) Format() pack.PackageFormat {
	return elfloader.pack.Format()
}

// Architecture implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Architecture() arch.Architecture {
	return elfloader.pack.(target.Target).Architecture()
}

// Platform implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Platform() plat.Platform {
	return elfloader.pack.(target.Target).Platform()
}

// Kernel implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Kernel() string {
	if len(elfloader.kernel) > 0 {
		return elfloader.kernel
	}

	if t, ok := elfloader.pack.(target.Target); ok {
		return t.Kernel()
	}

	return ""
}

// KernelDbg implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) KernelDbg() string {
	return elfloader.pack.(target.Target).KernelDbg()
}

// Initrd implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Initrd() initrd.Initrd {
	return elfloader.pack.(target.Target).Initrd()
}

// Command implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Entrypoint() string {
	return elfloader.pack.(target.Target).Entrypoint()
}

// Command implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) Command() []string {
	return elfloader.pack.(target.Target).Command()
}

// ConfigFilename implements kraftkit.sh/unikraft.target.Target
func (elfloader *ELFLoader) ConfigFilename() string {
	return elfloader.pack.(target.Target).ConfigFilename()
}
