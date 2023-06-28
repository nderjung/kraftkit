// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package tableprinter

import "fmt"

// TablePrinterOption ...
type TablePrinterOption func(*TablePrinter) error

// WithOutputFormat ...
func WithOutputFormat(format TableOutputFormat) TablePrinterOption {
	return func(opts *TablePrinter) error {
		opts.format = format
		return nil
	}
}

// WithOutputFormat ...
func WithOutputFormatFromString(format string) TablePrinterOption {
	return func(opts *TablePrinter) error {
		if format == "" {
			return fmt.Errorf("unsupported table printer format: %s", format)
		}
		opts.format = TableOutputFormat(format)
		return nil
	}
}

// WithTableDelimeter ...
func WithTableDelimeter(delim string) TablePrinterOption {
	return func(opts *TablePrinter) error {
		opts.delimeter = delim
		return nil
	}
}

func WithFieldTruncateFunc(truncateFunc func(int, string) string) TablePrinterOption {
	return func(opts *TablePrinter) error {
		opts.truncateFunc = truncateFunc
		return nil
	}
}

func WithMaxWidth(maxWidth int) TablePrinterOption {
	return func(opts *TablePrinter) error {
		opts.maxWidth = maxWidth
		return nil
	}
}
