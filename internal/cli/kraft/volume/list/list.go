// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	volumeapi "kraftkit.sh/api/volume/v1alpha1"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/internal/tableprinter"
	"kraftkit.sh/iostreams"
	"kraftkit.sh/log"
	"kraftkit.sh/machine/volume"
)

type List struct {
	Long   bool `long:"long" short:"l" usage:"Show more information"`
	driver string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&List{}, cobra.Command{
		Short:   "List machine volumes",
		Use:     "ls [FLAGS]",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "volume",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *List) Pre(cmd *cobra.Command, _ []string) error {
	opts.driver = cmd.Flag("driver").Value.String()
	return nil
}

func (opts *List) Run(ctx context.Context, args []string) error {
	var err error

	strategy, ok := volume.Strategies()[opts.driver]
	if !ok {
		return fmt.Errorf("unsupported volume driver strategy: %s", opts.driver)
	}

	controller, err := strategy.NewVolumeV1alpha1(ctx)
	if err != nil {
		return err
	}

	volumes, err := controller.List(ctx, &volumeapi.VolumeList{})
	if err != nil {
		return err
	}

	type volTable struct {
		id     string
		source string
		driver string
		status volumeapi.VolumeState
	}

	var items []volTable

	for _, volume := range volumes.Items {
		items = append(items, volTable{
			id:     string(volume.Name),
			driver: opts.driver,
			source: volume.Spec.Source,
			status: volume.Status.State,
		})
	}

	err = iostreams.G(ctx).StartPager()
	if err != nil {
		log.G(ctx).Errorf("error starting pager: %v", err)
	}

	defer iostreams.G(ctx).StopPager()

	cs := iostreams.G(ctx).ColorScheme()
	table, err := tableprinter.NewTablePrinter(ctx)
	if err != nil {
		return err
	}

	// Header row
	table.AddField("VOLUME ID", cs.Bold)
	table.AddField("SOURCE", cs.Bold)
	table.AddField("DRIVER", cs.Bold)
	table.AddField("STATUS", cs.Bold)
	table.EndRow()

	for _, item := range items {
		table.AddField(item.id, nil)
		table.AddField(item.source, nil)
		table.AddField(item.driver, nil)
		table.AddField(item.status.String(), nil)
		table.EndRow()
	}

	return table.Render(iostreams.G(ctx).Out)
}
