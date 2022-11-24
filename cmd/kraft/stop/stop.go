// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package stop

import (
	"context"
	"fmt"
	"sync"

	"kraftkit.sh/config"
	"kraftkit.sh/log"
	"kraftkit.sh/machine"
	machinedriver "kraftkit.sh/machine/driver"
	"kraftkit.sh/machine/driveropts"
	"kraftkit.sh/packmanager"

	"kraftkit.sh/internal/cmdfactory"
	"kraftkit.sh/internal/cmdutil"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type stopOptions struct {
	PackageManager func(opts ...packmanager.PackageManagerOption) (packmanager.PackageManager, error)
	ConfigManager  func() (*config.ConfigManager, error)
}

func StopCmd(f *cmdfactory.Factory) *cobra.Command {
	cmd, err := cmdutil.NewCmd(f, "stop")
	if err != nil {
		panic("could not initialize 'kraft stop' command")
	}

	opts := &stopOptions{
		PackageManager: f.PackageManager,
		ConfigManager:  f.ConfigManager,
	}

	cmd.Short = "Stop one or more running unikernels"
	cmd.Hidden = true
	cmd.Use = "stop [FLAGS] MACHINE [MACHINE [...]]"
	cmd.Args = cobra.MinimumNArgs(1)
	cmd.Long = heredoc.Doc(`
		Stop one or more running unikernels`)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runStop(opts, args...)
	}

	return cmd
}

type machineWaitGroup struct {
	lock sync.RWMutex
	mids []machine.MachineID
}

func (mwg *machineWaitGroup) Add(mid machine.MachineID) {
	mwg.lock.Lock()
	defer mwg.lock.Unlock()

	if mwg.Contains(mid) {
		return
	}

	mwg.mids = append(mwg.mids, mid)
}

func (mwg *machineWaitGroup) Done(needle machine.MachineID) {
	mwg.lock.Lock()
	defer mwg.lock.Unlock()

	if !mwg.Contains(needle) {
		return
	}

	for i, mid := range mwg.mids {
		if mid == needle {
			mwg.mids = append(mwg.mids[:i], mwg.mids[i+1:]...)
			return
		}
	}
}

func (mwg *machineWaitGroup) Wait() {
	for {
		if len(mwg.mids) == 0 {
			break
		}
	}
}

func (mwg *machineWaitGroup) Contains(needle machine.MachineID) bool {
	for _, mid := range mwg.mids {
		if mid == needle {
			return true
		}
	}

	return false
}

var (
	observations = new(machineWaitGroup)
	drivers      = make(map[machinedriver.DriverType]machinedriver.Driver)
)

func runStop(opts *stopOptions, args ...string) error {
	var err error

	cfgm, err := opts.ConfigManager()
	if err != nil {
		return err
	}

	ctx := context.Background()
	store, err := machine.NewMachineStoreFromPath(cfgm.Config.RuntimeDir)
	if err != nil {
		return fmt.Errorf("could not access machine store: %v", err)
	}

	allMids, err := store.ListAllMachineIDs()
	if err != nil {
		return fmt.Errorf("could not list machines: %v", err)
	}

	var mids []machine.MachineID

	for _, mid1 := range args {
		found := false
		for _, mid2 := range allMids {
			if mid1 == mid2.ShortString() || mid1 == mid2.String() {
				mids = append(mids, mid2)
				found = true
			}
		}

		if !found {
			return fmt.Errorf("could not find machine %s", mid1)
		}
	}

	for _, mid := range mids {
		mid := mid // loop closure

		if observations.Contains(mid) {
			continue
		}

		observations.Add(mid)

		go func() {
			observations.Add(mid)

			log.G(ctx).Infof("stopping %s...", mid.ShortString())

			state, err := store.LookupMachineState(mid)
			if err != nil {
				log.G(ctx).Errorf("could not look up machine state: %v", err)
				observations.Done(mid)
				return
			}

			switch state {
			case machine.MachineStateDead, machine.MachineStateExited:
				log.G(ctx).Errorf("%s has exited", mid.ShortString())
				observations.Done(mid)
				return
			}

			mcfg := &machine.MachineConfig{}
			if err := store.LookupMachineConfig(mid, mcfg); err != nil {
				log.G(ctx).Errorf("could not look up machine config: %v", err)
				observations.Done(mid)
				return
			}

			driverType := machinedriver.DriverTypeFromName(mcfg.DriverName)

			if _, ok := drivers[driverType]; !ok {
				driver, err := machinedriver.New(driverType,
					driveropts.WithMachineStore(store),
					driveropts.WithRuntimeDir(cfgm.Config.RuntimeDir),
				)
				if err != nil {
					log.G(ctx).Errorf("could not instantiate machine driver for %s: %v", mid.ShortString(), err)
					observations.Done(mid)
					return
				}

				drivers[driverType] = driver
			}

			driver := drivers[driverType]

			if err := driver.Stop(ctx, mid); err != nil {
				log.G(ctx).Errorf("could not stop machine %s: %v", mid.ShortString(), err)
			} else {
				log.G(ctx).Infof("stopped %s", mid.ShortString())
			}

			observations.Done(mid)
		}()
	}

	observations.Wait()

	return nil
}
