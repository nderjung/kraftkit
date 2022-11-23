// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package pack

import (
	"errors"
	"fmt"
	"os"

	"kraftkit.sh/initrd"
	"kraftkit.sh/log"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/utils"
)

// PackageOptions contains configuration for the Package
type PackageOptions struct {
	// Name of the package
	Name string

	// Type of package
	Type unikraft.ComponentType

	// Version of the package
	Version string

	// Architecture of the package if applicable
	Architecture *string

	// Platform of the package if applicable
	Platform *string

	// Kernel is the path on disk to the kernel
	Kernel string

	// Initrd contains the configuration for the initramfs file
	Initrd *initrd.InitrdConfig

	// Metadata represents other items that did not have appropriate annotations
	Metadata map[string]interface{}

	// RemoteLocation contains the remote location of the package.
	RemoteLocation string

	// LocalLocation contains the path to save or store the package on the host.
	LocalLocation string

	// Sha256
	Sha256 string

	// Access to a logger
	log log.Logger

	// workdir
	workdir string
}

type ContextKey string

type PackageOption func(opts *PackageOptions) error

// ArchPlatString returns the string representation of the architecture and
// platform combination for this package
func (opts *PackageOptions) ArchPlatString() string {
	plat := "*"
	arch := "*"

	if opts.Platform != nil {
		plat = *opts.Platform
	}

	if opts.Architecture != nil {
		arch = *opts.Architecture
	}

	return plat + "/" + arch
}

// NameVersion returns the string representation of name and version of this
// package
func (opts *PackageOptions) NameVersion() string {
	return opts.Name + ":" + opts.Version
}

// TypeNameVersion returns the string representation of the type, name and
// version of this package
func (opts *PackageOptions) TypeNameVersion() string {
	return opts.Type.Plural() + "/" + opts.Name + ":" + opts.Version
}

// NewPackageOptions creates PackageOptions
func NewPackageOptions(opts ...PackageOption) (*PackageOptions, error) {
	options := &PackageOptions{
		Metadata: make(map[string]interface{}),
	}

	for _, o := range opts {
		err := o(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func WithName(name string) PackageOption {
	return func(opts *PackageOptions) error {
		if len(name) == 0 {
			return fmt.Errorf("cannot set empty name")
		}

		opts.Name = name

		return nil
	}
}

func WithArchitecture(arch string) PackageOption {
	return func(opts *PackageOptions) error {
		opts.Architecture = &arch

		return nil
	}
}

func WithPlatform(plat string) PackageOption {
	return func(opts *PackageOptions) error {
		opts.Platform = &plat

		return nil
	}
}

func WithType(t unikraft.ComponentType) PackageOption {
	return func(opts *PackageOptions) error {
		opts.Type = t
		return nil
	}
}

func WithVersion(version string) PackageOption {
	return func(opts *PackageOptions) error {
		opts.Version = version
		return nil
	}
}

func WithMetadata(metadata map[string]interface{}) PackageOption {
	return func(opts *PackageOptions) error {
		opts.Metadata = metadata
		return nil
	}
}

// WithKerenl sets the metadata attribute path to the kernel image
func WithKernel(kernel string) PackageOption {
	return func(opts *PackageOptions) error {
		if len(kernel) == 0 {
			return fmt.Errorf("path to kernel cannot be empty")
		}
		if f, err := os.Stat(kernel); err != nil || errors.Is(err, os.ErrNotExist) || f.IsDir() {
			return fmt.Errorf("path to kernel does not exist: %s", kernel)
		}

		// TODO: Validate if a real kernel/ELF?
		opts.Kernel = kernel

		return nil
	}
}

// WithInitrdConfig sets the metadata attribute with the interface representing
// initrd configuration
func WithInitrdConfig(initrd *initrd.InitrdConfig) PackageOption {
	return func(opts *PackageOptions) error {
		if initrd == nil || len(initrd.Input) == 0 || initrd.Input == nil || len(initrd.Output) == 0 {
			return nil
		}

		opts.Initrd = initrd

		return nil
	}
}

// WithRemoteLocation sets the location of the package at its remote registry
func WithRemoteLocation(location string) PackageOption {
	return func(opts *PackageOptions) error {
		opts.RemoteLocation = location
		return nil
	}
}

// WithLocalLocation sets the location of the package to be stored locally.
func WithLocalLocation(location string, force bool) PackageOption {
	return func(opts *PackageOptions) error {
		if !force {
			if _, err := os.Stat(location); err == nil {
				return fmt.Errorf("local location already exists: %s", location)
			}
		}

		opts.LocalLocation = location
		return nil
	}
}

func WithLogger(l log.Logger) PackageOption {
	return func(opts *PackageOptions) error {
		opts.log = l
		return nil
	}
}

func WithWorkdir(dir string) PackageOption {
	return func(opts *PackageOptions) error {
		opts.workdir = dir
		return nil
	}
}

func (opts *PackageOptions) Log() log.Logger {
	return opts.log
}

func (opts *PackageOptions) Workdir() string {
	return opts.workdir
}

type PullPackageOptions struct {
	architectures     []string
	platforms         []string
	calculateChecksum bool
	onProgress        func(progress float64)
	workdir           string
	log               log.Logger
	useCache          bool
}

// OnProgress calls (if set) an embedded progress function which can be used to
// update an external progress bar, for example.
func (ppo *PullPackageOptions) OnProgress(progress float64) {
	if ppo.onProgress != nil {
		ppo.onProgress(progress)
	}
}

// Workdir returns the set working directory as part of the pull request
func (ppo *PullPackageOptions) Workdir() string {
	return ppo.workdir
}

// CalculateChecksum returns whether the pull request should perform a check of
// the resource sum.
func (ppo *PullPackageOptions) CalculateChecksum() bool {
	return ppo.calculateChecksum
}

// Log returns the available logger
func (ppo *PullPackageOptions) Log() log.Logger {
	return ppo.log
}

// UseCache returns whether the pull should redirect to using a local cache if
// available.
func (ppo *PullPackageOptions) UseCache() bool {
	return ppo.useCache
}

type PullPackageOption func(opts *PullPackageOptions) error

// NewPullPackageOptions creates PullPackageOptions
func NewPullPackageOptions(opts ...PullPackageOption) (*PullPackageOptions, error) {
	options := &PullPackageOptions{}

	for _, o := range opts {
		err := o(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

// WithPullArchitecture requests a given architecture (if applicable)
func WithPullArchitecture(archs ...string) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		for _, arch := range archs {
			if arch == "" {
				continue
			}

			if utils.Contains(opts.architectures, arch) {
				continue
			}

			opts.architectures = append(opts.architectures, archs...)
		}

		return nil
	}
}

// WithPullPlatform requests a given platform (if applicable)
func WithPullPlatform(plats ...string) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		for _, plat := range plats {
			if plat == "" {
				continue
			}

			if utils.Contains(opts.platforms, plat) {
				continue
			}

			opts.platforms = append(opts.platforms, plats...)
		}

		return nil
	}
}

// WithPullProgressFunc set an optional progress function which is used as a
// callback during the transmission of the package and the host
func WithPullProgressFunc(onProgress func(progress float64)) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		opts.onProgress = onProgress
		return nil
	}
}

// WithPullWorkdir set the working directory context of the pull such that the
// resources of the package are placed there appropriately
func WithPullWorkdir(workdir string) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		opts.workdir = workdir
		return nil
	}
}

// WithPullLogger set the use of a logger
func WithPullLogger(l log.Logger) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		opts.log = l
		return nil
	}
}

// WithPullChecksum to set whether to calculate and compare the checksum of the
// package
func WithPullChecksum(calc bool) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		opts.calculateChecksum = calc
		return nil
	}
}

// WithPullCache to set whether use cache if possible
func WithPullCache(cache bool) PullPackageOption {
	return func(opts *PullPackageOptions) error {
		opts.useCache = cache
		return nil
	}
}
