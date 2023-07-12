// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package lib

import (
	"github.com/spf13/cobra"

	"kraftkit.sh/cmd/kraft/lib/create"
	"kraftkit.sh/cmdfactory"
)

type Lib struct{}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Lib{}, cobra.Command{
		Short:   "Manage and maintain Unikraft microlibraries",
		Use:     "lib SUBCOMMAND",
		Aliases: []string{"library"},
		Hidden:  true,
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "lib",
		},
	})
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(create.New())
	// cmd.AddCommand(down.New())
	// cmd.AddCommand(inspect.New())
	// cmd.AddCommand(list.New())
	// cmd.AddCommand(remove.New())
	// cmd.AddCommand(up.New())

	return cmd
}

func (opts *Lib) Run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
