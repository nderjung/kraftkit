// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package list

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kraftcloud "sdk.kraft.cloud"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/tableprinter"
	"kraftkit.sh/iostreams"
	"kraftkit.sh/log"
)

type ListOptions struct {
	Output string `long:"output" short:"o" usage:"Set output format" default:"table"`
	Watch  bool   `long:"watch" short:"w" usage:"After listing watch for changes."`

	metro string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&ListOptions{}, cobra.Command{
		Short:   "List all volumes at a metro for your account",
		Use:     "ls",
		Aliases: []string{"list"},
		Long: heredoc.Doc(`
			List all volumes in your account.
		`),
		Example: heredoc.Doc(`
			# List all volumes in your account.
			$ kraft cloud vol ls
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud-volume",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *ListOptions) Pre(cmd *cobra.Command, _ []string) error {
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

func (opts *ListOptions) Run(ctx context.Context, args []string) error {
	auth, err := config.GetKraftCloudLoginFromContext(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve credentials: %w", err)
	}

	client := kraftcloud.NewVolumesClient(
		kraftcloud.WithToken(auth.Token),
	)

	volumes, err := client.WithMetro(opts.metro).List(ctx)
	if err != nil {
		return fmt.Errorf("could not list volumes: %w", err)
	}

	err = iostreams.G(ctx).StartPager()
	if err != nil {
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

	// printed := make(map[string]bool, len(volumes))

	// Header row
	table.AddField("UUID", cs.Bold)
	table.AddField("STATUS", cs.Bold)
	table.AddField("PERSISTENT", cs.Bold)
	table.AddField("SIZE", cs.Bold)
	table.AddField("ATTACHED", cs.Bold)
	table.EndRow()

	for _, volume := range volumes {
		table.AddField(volume.UUID, nil)
		table.AddField(string(volume.Status), nil)
		table.AddField(fmt.Sprintf("%v", volume.Persistent), nil)
		table.AddField(fmt.Sprintf("%d MB", volume.SizeMB), nil)
		table.AddField(fmt.Sprintf("%s", volume.AttachedTo), nil)

		table.EndRow()
	}

	if err := table.Render(iostreams.G(ctx).Out); err != nil {
		return err
	}

	// for ok := opts.Watch; ok; {
	// 	images, err := client.WithMetro(opts.metro).List(ctx, map[string]interface{}{})
	// 	if err != nil {
	// 		return fmt.Errorf("could not list images: %w", err)
	// 	}

	// 	table, err := tableprinter.NewTablePrinter(ctx,
	// 		tableprinter.WithMaxWidth(iostreams.G(ctx).TerminalWidth()),
	// 		tableprinter.WithOutputFormatFromString(opts.Output),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	for _, image := range images {
	// 		if _, ok := printed[image.Digest]; ok {
	// 			continue
	// 		}

	// 		printed[image.Digest] = true

	// 		if len(image.Tags) > 0 {
	// 			table.AddField(image.Tags[0], nil)
	// 		} else {
	// 			table.AddField(image.Digest, nil)
	// 		}
	// 		table.AddField(fmt.Sprintf("%v", image.Public), nil)
	// 		table.AddField(fmt.Sprintf("%v", image.Initrd), nil)
	// 		table.AddField(strings.TrimSpace(fmt.Sprintf("%s -- %s", image.KernelArgs, image.Args)), nil)
	// 		table.AddField(humanize.Bytes(uint64(image.SizeInBytes)), nil)
	// 		table.EndRow()
	// 	}

	// 	if err := table.Render(iostreams.G(ctx).Out); err != nil {
	// 		return err
	// 	}

	// 	time.Sleep(time.Second)
	// }

	return nil
}
