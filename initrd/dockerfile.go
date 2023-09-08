// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package initrd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
	"kraftkit.sh/config"
	"kraftkit.sh/log"

	"github.com/cavaliergopher/cpio"
	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progressui"
)

type dockerfile struct {
	opts       InitrdOptions
	dockerfile string
	files      []string
	workdir    string
}

// NewFromDockerfile accepts an input path which represents a Dockerfile that
// can be constructed via buildkit to become a CPIO archive.
func NewFromDockerfile(ctx context.Context, path string, opts ...InitrdOption) (Initrd, error) {
	if !strings.Contains(strings.ToLower(path), "dockerfile") {
		return nil, fmt.Errorf("file is not a Dockerfile")
	}

	initrd := dockerfile{
		opts:       InitrdOptions{},
		dockerfile: path,
		workdir:    filepath.Dir(path),
	}

	for _, opt := range opts {
		if err := opt(&initrd.opts); err != nil {
			return nil, err
		}
	}

	return &initrd, nil
}

// Build implements Initrd.
func (initrd *dockerfile) Build(ctx context.Context) (string, error) {
	if initrd.opts.output == "" {
		fi, err := os.CreateTemp("", "")
		if err != nil {
			return "", err
		}

		initrd.opts.output = fi.Name()
	}

	outputDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("could not make temporary directory: %w", err)
	}

	buildkitAddr := config.G[config.KraftKit](ctx).BuildKitHost

	c, err := client.New(ctx,
		buildkitAddr,
		client.WithFailFast(),
	)
	if err != nil {
		return "", fmt.Errorf("could not instantiate BuildKit client: %w", err)
	}

	var cacheExports []client.CacheOptionsEntry
	if len(initrd.opts.cacheDir) > 0 {
		cacheExports = []client.CacheOptionsEntry{
			{
				Type: "local",
				Attrs: map[string]string{
					"dest": initrd.opts.cacheDir,
				},
			},
		}
	}

	solveOpt := &client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type:      client.ExporterLocal,
				OutputDir: outputDir,
			},
		},
		CacheExports: cacheExports,
		LocalDirs: map[string]string{
			"context":    initrd.workdir,
			"dockerfile": filepath.Dir(initrd.dockerfile),
		},
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": filepath.Base(initrd.dockerfile),
		},
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := c.Solve(ctx, nil, *solveOpt, ch)
		if err != nil {
			return fmt.Errorf("could not solve: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var c console.Console
		if config.G[config.KraftKit](ctx).Log.Type == "fancy" && os.Getenv("GITHUB_ACTIONS") != "true" {
			if cn, err := console.ConsoleFromFile(os.Stderr); err == nil {
				c = cn
			}
		}

		_, err = progressui.DisplaySolveStatus(ctx, c, log.G(ctx).Writer(), ch)
		if err != nil {
			return fmt.Errorf("could not display output progress: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", fmt.Errorf("could not wait for err group: %w", err)
	}

	f, err := os.OpenFile(initrd.opts.output, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", fmt.Errorf("could not open initramfs file: %w", err)
	}

	defer f.Close()

	writer := cpio.NewWriter(f)
	defer writer.Close()

	// Recursively walk the output directory on successful build and serialize to
	// the output
	if err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("received error before parsing path: %w", err)
		}

		internal := strings.TrimPrefix(path, filepath.Clean(outputDir))
		if internal == "" {
			return nil // Do not archive empty paths
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("could not get directory entry info: %w", err)
		}

		if d.Type().IsDir() {
			if err := writer.WriteHeader(&cpio.Header{
				Name: internal,
				Mode: cpio.TypeDir,
			}); err != nil {
				return fmt.Errorf("could not write CPIO header: %w", err)
			}

			return nil
		}

		initrd.files = append(initrd.files, internal)

		log.G(ctx).
			WithField("file", internal).
			Trace("archiving")

		var data []byte
		targetLink := ""
		if info.Mode()&os.ModeSymlink != 0 {
			targetLink, err = os.Readlink(path)
			data = []byte(targetLink)
		} else if d.Type().IsRegular() {
			data, err = os.ReadFile(path)
		} else {
			log.G(ctx).Warnf("unsupported file: %s", path)
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}

		if err := writer.WriteHeader(&cpio.Header{
			Name:     internal,
			Linkname: targetLink,
			Mode:     cpio.FileMode(d.Type().Perm()),
			Size:     info.Size(),
		}); err != nil {
			return fmt.Errorf("could not write CPIO header: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("could not write CPIO data: %w", err)
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("could not walk output path: %w", err)
	}

	return initrd.opts.output, nil
}

// Files implements Initrd.
func (initrd *dockerfile) Files() []string {
	return initrd.files
}
