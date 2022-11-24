// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package processtree

type ProcessTreeOption func(pt *ProcessTree) error

func WithVerb(verb string) ProcessTreeOption {
	return func(pt *ProcessTree) error {
		pt.verb = verb
		return nil
	}
}

func WithRenderer(norender bool) ProcessTreeOption {
	return func(pt *ProcessTree) error {
		pt.norender = norender
		return nil
	}
}

func IsParallel(parallel bool) ProcessTreeOption {
	return func(pt *ProcessTree) error {
		pt.parallel = parallel
		return nil
	}
}

func WithFailFast(failFast bool) ProcessTreeOption {
	return func(pt *ProcessTree) error {
		pt.failFast = failFast
		return nil
	}
}
