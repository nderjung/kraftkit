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

	"github.com/cavaliergopher/cpio"
)

type directory struct {
	opts  InitrdOptions
	path  string
	files []string
}

// NewFromDirectory returns an instantiated Initrd interface which is is able to
// serialize a rootfs from a given directory.
func NewFromDirectory(_ context.Context, path string, opts ...InitrdOption) (Initrd, error) {
	rootfs := directory{
		opts: InitrdOptions{},
		path: path,
	}

	for _, opt := range opts {
		if err := opt(&rootfs.opts); err != nil {
			return nil, err
		}
	}

	fi, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	} else if err != nil {
		return nil, fmt.Errorf("could not check path: %w", err)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("supplied path is not a directory: %s", path)
	}

	return &rootfs, nil
}

// Build implements Initrd.
func (initrd *directory) Build(_ context.Context) (string, error) {
	if initrd.opts.output == "" {
		fi, err := os.CreateTemp("", "")
		if err != nil {
			return "", err
		}

		initrd.opts.output = fi.Name()
	}

	f, err := os.OpenFile(initrd.opts.output, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", fmt.Errorf("could not open initramfs file: %w", err)
	}

	defer f.Close()

	writer := cpio.NewWriter(f)
	defer writer.Close()

	if err := filepath.WalkDir(initrd.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		internal := strings.TrimPrefix(path, initrd.path)
		if internal == "" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.Type().IsDir() {
			if err := writer.WriteHeader(&cpio.Header{
				Name: internal,
				Mode: cpio.TypeDir,
			}); err != nil {
				return err
			}

			return nil
		}

		initrd.files = append(initrd.files, internal)

		var data []byte
		targetLink := ""
		if info, err := d.Info(); err == nil && info.Mode()&os.ModeSymlink != 0 {
			targetLink, err = os.Readlink(path)
			if err != nil {
				return err
			}
			data = []byte(targetLink)
		} else if d.Type().IsRegular() {
			data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}

		if err := writer.WriteHeader(&cpio.Header{
			Name:     internal,
			Linkname: targetLink,
			Mode:     cpio.FileMode(d.Type().Perm()),
			Size:     info.Size(),
		}); err != nil {
			return err
		}

		if _, err := writer.Write(data); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return initrd.opts.output, nil
}

// Files implements Initrd.
func (initrd *directory) Files() []string {
	return initrd.files
}
