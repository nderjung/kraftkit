// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2012 Alex Ogier.
// Copyright (c) 2012 The Go Authors.
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package cli

import (
	"github.com/spf13/pflag"
)

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string { return string(*s) }

// StringVarP is like StringVar, but accepts a shorthand letter that can be used
// after a single dash.
func StringVarP(p *string, name, shorthand string, value string, usage string) *pflag.Flag {
	return VarPF(newStringValue(value, p), name, shorthand, usage)
}
