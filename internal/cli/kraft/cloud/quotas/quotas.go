// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package quotas

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/tableprinter"
	"kraftkit.sh/iostreams"
	"kraftkit.sh/log"

	kraftcloud "sdk.kraft.cloud"
)

type QuotasOptions struct {
	Output string `local:"true" long:"output" short:"o" usage:"Set output format" default:"table"`

	metro string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&QuotasOptions{}, cobra.Command{
		Short:   "View your resource quota on KraftCloud",
		Use:     "quotas",
		Aliases: []string{"q", "quota"},
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *QuotasOptions) Pre(cmd *cobra.Command, _ []string) error {
	opts.metro = cmd.Flag("metro").Value.String()
	if opts.metro == "" {
		opts.metro = os.Getenv("KRAFTCLOUD_METRO")
	}
	if opts.metro == "" {
		return fmt.Errorf("kraftcloud metro is unset")
	}
	log.G(cmd.Context()).WithField("metro", opts.metro).Debug("using")
	return nil
}

func (opts *QuotasOptions) Run(ctx context.Context, _ []string) error {
	auth, err := config.GetKraftCloudLoginFromContext(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve credentials: %w", err)
	}

	client := kraftcloud.NewUsersClient(
		kraftcloud.WithToken(auth.Token),
	)

	quotas, err := client.WithMetro(opts.metro).Quotas(ctx)
	if err != nil {
		return fmt.Errorf("could not get quotas: %w", err)
	}

	if err := iostreams.G(ctx).StartPager(); err != nil {
		log.G(ctx).Errorf("error starting pager: %v", err)
	}

	defer iostreams.G(ctx).StopPager()

	cs := iostreams.G(ctx).ColorScheme()
	table, err := tableprinter.NewTablePrinter(ctx,
		tableprinter.WithMaxWidth(iostreams.G(ctx).TerminalWidth()),
		tableprinter.WithOutputFormatFromString(opts.Output),
	)
	if err != nil {
		return err
	}

	table.AddField("LIVE INSTANCES", cs.Bold)
	table.AddField("TOTAL INSTANCES", cs.Bold)
	table.AddField("MAX TOTAL INSTANCES", cs.Bold)
	table.EndRow()

	table.AddField(fmt.Sprintf("%d", quotas.Used.Instances), nil)
	table.AddField(fmt.Sprintf("%d", quotas.Used.LiveInstances), nil)
	table.AddField(fmt.Sprintf("%d", quotas.Hard.Instances), nil)
	table.EndRow()

	return table.Render(iostreams.G(ctx).Out)
}
