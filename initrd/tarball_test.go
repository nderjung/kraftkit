// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package initrd_test

import (
	"context"
	"io"
	"os"
	"testing"

	"kraftkit.sh/cpio"
	"kraftkit.sh/initrd"
)

func TestNewFromTarball(t *testing.T) {
	const rootfsTarball = "testdata/rootfs.tar.gz"

	ctx := context.Background()

	ird, err := initrd.NewFromTarball(ctx, rootfsTarball)
	if err != nil {
		t.Fatal("NewFromTarball:", err)
	}

	irdPath, err := ird.Build(ctx)
	if err != nil {
		t.Fatal("Build:", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(irdPath); err != nil {
			t.Fatal("Failed to remove initrd file:", err)
		}
	})

	r := cpio.NewReader(openFile(t, irdPath))

	var gotFiles []string

	for {
		hdr, _, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal("Failed to read next cpio header:", err)
		}

		if hdr.Name == "/" {
			continue
		}

		gotFiles = append(gotFiles, hdr.Name)

		expectHdr, ok := expectHeaders[hdr.Name]
		if !ok {
			t.Error("Encountered unexpected file in cpio archive:", hdr.Name)
			continue
		}

		if gotMode := hdr.Mode & cpio.ModeType; gotMode != expectHdr.Mode {
			t.Errorf("file [%s]: got mode %s, expected %s", hdr.Name, gotMode, expectHdr.Mode)
		}
		if hdr.Linkname != expectHdr.Linkname {
			t.Errorf("file [%s]: got linkname %q, expected %q", hdr.Name, hdr.Linkname, expectHdr.Linkname)
		}
		if hdr.Size != expectHdr.Size {
			t.Errorf("file [%s]: got size %d, expected %d", hdr.Name, hdr.Size, expectHdr.Size)
		}
	}

	if len(gotFiles) != len(expectHeaders) {
		t.Errorf("Expected %d files, got %d: %#v", len(expectHeaders), len(gotFiles), gotFiles)
	}
}