// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package vfscore

import (
	"strings"

	"kraftkit.sh/unikraft/export/v0/ukargparse"
)

var ParamVfsAutomount = ukargparse.NewParamStrSlice("vfs", "automount", nil)

// ExportedParams returns the parameters available by this exported library.
func ExportedParams() []ukargparse.Param {
	return []ukargparse.Param{
		ParamVfsAutomount,
	}
}

// Automount is a vfscore mount entry.
type Automount struct {
	sourceDevice string
	mountTarget  string
	fsDriver     string
	options      string
}

// NewAutomount generates a structure that is representative of one of
// Unikraft's vfscore automounts.
func NewAutomount(sourceDevice, mountTarget, fsDriver, options string) Automount {
	return Automount{
		sourceDevice,
		mountTarget,
		fsDriver,
		options,
	}
}

// String implements fmt.Stringer and returns a valid vfs.automount-formatted
// entry.
func (automount Automount) String() string {
	return strings.Join([]string{
		automount.sourceDevice,
		automount.mountTarget,
		automount.fsDriver,
		automount.options,
	}, ":")
}
