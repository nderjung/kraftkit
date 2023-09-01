// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package run

import (
	"context"
	"fmt"
	"os"

	machineapi "kraftkit.sh/api/machine/v1alpha1"
	"kraftkit.sh/config"
	"kraftkit.sh/log"
	"kraftkit.sh/pack"
	"kraftkit.sh/tui/paraprogress"
	"kraftkit.sh/unikraft/app"
	"kraftkit.sh/unikraft/elfloader"
	"kraftkit.sh/unikraft/target"
)

type runnerKraftfileRuntime struct {
	args    []string
	project app.Application
}

// String implements Runner.
func (runner *runnerKraftfileRuntime) String() string {
	return "kraftfile-runtime"
}

// Runnable implements Runner.
func (runner *runnerKraftfileRuntime) Runnable(ctx context.Context, opts *Run, args ...string) (bool, error) {
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("getting current working directory: %w", err)
	}

	if len(args) == 0 {
		opts.workdir = cwd
	} else {
		opts.workdir = cwd
		runner.args = args
		if f, err := os.Stat(args[0]); err == nil && f.IsDir() {
			opts.workdir = args[0]
			runner.args = args[1:]
		}
	}

	if !app.IsWorkdirInitialized(opts.workdir) {
		return false, fmt.Errorf("path is not project: %s", opts.workdir)
	}

	popts := []app.ProjectOption{
		app.WithProjectWorkdir(opts.workdir),
	}

	if len(opts.Kraftfile) > 0 {
		popts = append(popts, app.WithProjectKraftfile(opts.Kraftfile))
	} else {
		popts = append(popts, app.WithProjectDefaultKraftfiles())
	}

	runner.project, err = app.NewProjectFromOptions(ctx, popts...)
	if err != nil {
		return false, fmt.Errorf("could not instantiate project directory %s: %v", opts.workdir, err)
	}

	if runner.project.Runtime() == nil {
		return false, fmt.Errorf("cannot run project without runtime directive")
	}

	return true, nil
}

// Prepare implements Runner.
func (runner *runnerKraftfileRuntime) Prepare(ctx context.Context, opts *Run, machine *machineapi.Machine, args ...string) error {
	var err error

	// Filter project targets by any provided CLI options
	targets := target.Filter(
		runner.project.Targets(),
		opts.Architecture,
		opts.platform.String(),
		opts.Target,
	)

	var t target.Target

	switch {
	case len(targets) == 0:
		return fmt.Errorf("could not detect any project targets based on plat=\"%s\" arch=\"%s\"", opts.platform.String(), opts.Architecture)

	case len(targets) == 1:
		t = targets[0]

	case config.G[config.KraftKit](ctx).NoPrompt && len(targets) > 1:
		return fmt.Errorf("could not determine what to run based on provided CLI arguments")

	default:
		t, err = target.Select(targets)
		if err != nil {
			return fmt.Errorf("could not select target: %v", err)
		}
	}

	lopts := []elfloader.ELFLoaderPrebuiltOption{}

	if len(runner.project.Runtime().Name()) > 0 {
		lopts = append(lopts, elfloader.WithName(runner.project.Runtime().Name()))
	} else if len(runner.project.Runtime().Kernel()) > 0 {
		lopts = append(lopts, elfloader.WithKernel(runner.project.Runtime().Kernel()))
	}

	if runner.project.Rootfs() != "" && opts.Rootfs == "" {
		opts.Rootfs = runner.project.Rootfs()
	}

	loader, err := elfloader.NewELFLoaderFromPrebuilt(ctx, lopts...)
	if err != nil {
		return err
	}

	// Create a temporary directory where the image can be stored
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	paramodel, err := paraprogress.NewParaProgress(
		ctx,
		[]*paraprogress.Process{paraprogress.NewProcess(
			fmt.Sprintf("pulling %s", loader.Name()),
			func(ctx context.Context, w func(progress float64)) error {
				popts := []pack.PullOption{
					pack.WithPullWorkdir(dir),
					pack.WithPullPlatform(opts.platform.String()),
				}
				if log.LoggerTypeFromString(config.G[config.KraftKit](ctx).Log.Type) == log.FANCY {
					popts = append(popts, pack.WithPullProgressFunc(w))
				}

				return loader.Pull(
					ctx,
					popts...,
				)
			},
		)},
		paraprogress.IsParallel(false),
		paraprogress.WithRenderer(
			log.LoggerTypeFromString(config.G[config.KraftKit](ctx).Log.Type) != log.FANCY,
		),
		paraprogress.WithFailFast(true),
	)
	if err != nil {
		return err
	}

	if err := paramodel.Start(); err != nil {
		return err
	}

	machine.Spec.Architecture = loader.Architecture().Name()
	machine.Spec.Platform = loader.Platform().Name()
	machine.Spec.Kernel = fmt.Sprintf("%s://%s:%s", loader.Format(), loader.Name(), loader.Version())

	if len(t.Command()) > 0 {
		machine.Spec.ApplicationArgs = t.Command()
	} else if len(runner.project.Command()) > 0 {
		machine.Spec.ApplicationArgs = runner.project.Command()
	} else if len(loader.Command()) > 0 {
		machine.Spec.ApplicationArgs = loader.Command()
	}

	// Use the symbolic debuggable kernel image?
	if opts.WithKernelDbg {
		machine.Status.KernelPath = loader.KernelDbg()
	} else {
		machine.Status.KernelPath = loader.Kernel()
	}

	if opts.Rootfs == "" {
		if runner.project.Rootfs() != "" {
			opts.Rootfs = runner.project.Rootfs()
		} else if loader.Initrd() != nil {
			machine.Status.InitrdPath, err = loader.Initrd().Build(ctx)
			if err != nil {
				return err
			}
		}
	}

	if err := opts.parseKraftfileVolumes(ctx, runner.project, machine); err != nil {
		return err
	}

	return nil
}
