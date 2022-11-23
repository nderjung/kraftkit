// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gobwas/glob"

	"kraftkit.sh/config"
	"kraftkit.sh/internal/cmdutil"
	"kraftkit.sh/pack"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft"
)

type ManifestManager struct {
	opts *packmanager.PackageManagerOptions
}

// useGit is a local variable used within the context of the manifest package
// and is dynamically injected as a CLI option.
var useGit = false

func init() {
	options, err := packmanager.NewPackageManagerOptions(
		context.TODO(),
	)
	if err != nil {
		panic(fmt.Sprintf("could not register package manager options: %s", err))
	}

	manager, err := NewManifestPackageManagerFromOptions(options)
	if err != nil {
		panic(fmt.Sprintf("could not register package manager: %s", err))
	}

	// Register a new pack.Package type
	packmanager.RegisterPackageManager(ManifestContext, manager)

	// Register additional command-line flags
	cmdutil.RegisterFlag(
		"kraft pkg pull",
		cmdutil.BoolVarP(
			&useGit,
			"git", "g",
			false,
			"Use Git when pulling sources",
		),
	)
}

func NewManifestPackageManagerFromOptions(opts *packmanager.PackageManagerOptions) (packmanager.PackageManager, error) {
	return ManifestManager{
		opts: opts,
	}, nil
}

// NewPackage initializes a new package
func (mm ManifestManager) NewPackageFromOptions(ctx context.Context, opts *pack.PackageOptions) ([]pack.Package, error) {
	p, err := NewPackageFromOptions(ctx, opts, nil)
	return []pack.Package{p}, err
}

// Options allows you to view the current options.
func (mm ManifestManager) Options() *packmanager.PackageManagerOptions {
	return mm.opts
}

func (mm ManifestManager) ApplyOptions(pmopts ...packmanager.PackageManagerOption) error {
	for _, opt := range pmopts {
		if err := opt(mm.opts); err != nil {
			return err
		}
	}

	return nil
}

// update retrieves and returns a cache of the upstream manifest registry
func (mm ManifestManager) update() (*ManifestIndex, error) {
	cfm := mm.opts.ConfigManager
	if len(cfm.Config.Unikraft.Manifests) == 0 {
		return nil, fmt.Errorf("no manifests specified in config")
	}

	localIndex := &ManifestIndex{
		LastUpdated: time.Now(),
	}

	mopts := []ManifestOption{
		WithAuthConfig(mm.Options().ConfigManager.Config.Auth),
		WithSourcesRootDir(mm.Options().ConfigManager.Config.Paths.Sources),
		WithLogger(mm.Options().Log),
	}

	for _, manipath := range cfm.Config.Unikraft.Manifests {
		// If the path of the manipath is the same as the current manifest or it
		// resides in the same directory as KraftKit's configured path for manifests
		// then we can skip this since we don't want to update ourselves.
		// if manipath == mm.LocalManifestIndex() || filepath.Dir(manipath) == mm.LocalManifestsDir() {
		// 	mm.opts.Log.Debugf("skipping: %s", manipath)
		// 	continue
		// }

		mm.opts.Log.Infof("fetching %s", manipath)

		manifests, err := FindManifestsFromSource(manipath, mopts...)
		if err != nil {
			mm.opts.Log.Warnf("%s", err)
		}

		localIndex.Manifests = append(localIndex.Manifests, manifests...)
	}

	return localIndex, nil
}

func (mm ManifestManager) Update() error {
	// Create parent directories if not present
	if err := os.MkdirAll(filepath.Dir(mm.LocalManifestIndex()), 0o771); err != nil {
		return err
	}

	localIndex, err := mm.update()
	if err != nil {
		return err
	}

	// TODO: Partition directories when there is a large number of manifests
	// TODO: Merge manifests of same name and type?

	// Create a file for each manifest
	for i, manifest := range localIndex.Manifests {
		filename := manifest.Name + ".yaml"

		if manifest.Type != unikraft.ComponentTypeCore {
			filename = manifest.Type.Plural() + "/" + filename
		}

		fileloc := filepath.Join(mm.LocalManifestsDir(), filename)
		if err := os.MkdirAll(filepath.Dir(fileloc), 0o771); err != nil {
			return err
		}

		mm.opts.Log.Infof("saving %s", fileloc)
		if err := manifest.WriteToFile(fileloc); err != nil {
			mm.opts.Log.Errorf("could not save manifest: %s", err)
		}

		// Replace manifest with relative path
		localIndex.Manifests[i] = &Manifest{
			Name:     manifest.Name,
			Type:     manifest.Type,
			Manifest: "./" + filename,
		}
	}

	return localIndex.WriteToFile(mm.LocalManifestIndex())
}

func (mm ManifestManager) AddSource(source string) error {
	cfm := mm.opts.ConfigManager
	cfg := cfm.Config

	for _, manifest := range cfg.Unikraft.Manifests {
		if source == manifest {
			mm.opts.Log.Warnf("manifest already saved: %s", source)
			return nil
		}
	}

	mm.opts.Log.Infof("adding to list of manifests: %s", source)
	cfg.Unikraft.Manifests = append(cfg.Unikraft.Manifests, source)
	return cfm.Write(true)
}

func (mm ManifestManager) RemoveSource(source string) error {
	cfm := mm.opts.ConfigManager
	cfg := cfm.Config
	manifests := []string{}

	for _, manifest := range cfg.Unikraft.Manifests {
		if source != manifest {
			manifests = append(manifests, manifest)
		}
	}

	mm.opts.Log.Infof("removing from list of manifests: %s", source)
	cfg.Unikraft.Manifests = manifests
	return cfm.Write(false)
}

// Push the resulting package to the supported registry of the implementation.
func (mm ManifestManager) Push(path string) error {
	return fmt.Errorf("not implemented pack.ManifestManager.Pushh")
}

// Pull a package from the support registry of the implementation.
func (mm ManifestManager) Pull(path string, opts *pack.PullPackageOptions) ([]pack.Package, error) {
	if _, err := mm.IsCompatible(path); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("not implemented pack.ManifestManager.Pull")
}

func (um ManifestManager) From(sub string) (packmanager.PackageManager, error) {
	return nil, fmt.Errorf("method not applicable to manifest manager")
}

func (mm ManifestManager) Catalog(query packmanager.CatalogQuery, popts ...pack.PackageOption) ([]pack.Package, error) {
	var err error
	var index *ManifestIndex
	var allManifests []*Manifest

	mopts := []ManifestOption{
		WithAuthConfig(mm.Options().ConfigManager.Config.Auth),
		WithSourcesRootDir(mm.Options().ConfigManager.Config.Paths.Sources),
		WithLogger(mm.Options().Log),
	}

	if len(query.Source) > 0 && query.NoCache {
		manifest, err := FindManifestsFromSource(query.Source, mopts...)
		if err != nil {
			return nil, err
		}

		allManifests = append(allManifests, manifest...)
	} else if query.NoCache {
		index, err = mm.update()
		if err != nil {
			return nil, err
		}

	} else {
		index, err = NewManifestIndexFromFile(mm.LocalManifestIndex())
		if err != nil {
			return nil, err
		}
	}

	if index != nil {
		for _, manifest := range index.Manifests {
			if len(manifest.Manifest) > 0 {
				manifests, err := findManifestsFromSource(mm.LocalManifestsDir(), manifest.Manifest, mopts)
				if err != nil {
					return nil, err
				}

				allManifests = append(allManifests, manifests...)
			} else {
				allManifests = append(allManifests, manifest)
			}
		}
	}

	var packages []pack.Package
	var g glob.Glob

	if len(query.Name) > 0 {
		t, n, v, err := unikraft.GuessTypeNameVersion(query.Name)

		// Overwrite additional attributes if pattern-matchable
		if err == nil {
			query.Name = n
			query.Version = v
			if t != unikraft.ComponentTypeUnknown {
				query.Types = append(query.Types, t)
			}
		}

		g = glob.MustCompile(query.Name)
	}

	for _, manifest := range allManifests {
		if len(query.Types) > 0 {
			found := false
			for _, t := range query.Types {
				if manifest.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if len(query.Source) > 0 && manifest.Origin != query.Source {
			continue
		}

		if len(query.Name) > 0 && !g.Match(manifest.Name) {
			continue
		}

		var versions []string
		if len(query.Version) > 0 {
			for _, version := range manifest.Versions {
				if version.Version == query.Version {
					versions = append(versions, version.Version)
					break
				}
			}
			if len(versions) == 0 {
				for _, channel := range manifest.Channels {
					if channel.Name == query.Version {
						versions = append(versions, channel.Name)
						break
					}
				}
			}

			if len(versions) == 0 {
				break
			}
		}

		if len(versions) > 0 {
			for _, version := range versions {
				p, err := NewPackageWithVersion(mm.opts.Context(), manifest, version, popts...)
				if err != nil {
					mm.opts.Log.Warnf("%v", err)
					continue
					// TODO: Config option for fast-fail?
					// return nil, err
				}

				packages = append(packages, p)
			}
		} else {
			packs, err := NewPackageFromManifest(mm.opts.Context(), manifest, popts...)
			if err != nil {
				mm.opts.Log.Warnf("%v", err)
				continue
				// TODO: Config option for fast-fail?
				// return nil, err
			}

			packages = append(packages, packs)
		}
	}

	for i := range packages {
		packages[i].ApplyOptions(
			pack.WithLogger(mm.Options().Log),
		)
	}

	return packages, nil
}

func (mm ManifestManager) IsCompatible(source string) (packmanager.PackageManager, error) {
	if _, err := NewProvider(source); err != nil {
		return nil, fmt.Errorf("incompatible source")
	}

	return mm, nil
}

// LocalManifestDir returns the user configured path to all the manifests
func (mm ManifestManager) LocalManifestsDir() string {
	if len(mm.opts.ConfigManager.Config.Paths.Manifests) > 0 {
		return mm.opts.ConfigManager.Config.Paths.Manifests
	}

	return filepath.Join(config.DataDir(), "manifests")
}

// LocalManifestIndex returns the user configured path to the manifest index
func (mm ManifestManager) LocalManifestIndex() string {
	return filepath.Join(mm.LocalManifestsDir(), "index.yaml")
}

func (mm ManifestManager) Format() string {
	return string(ManifestContext)
}
