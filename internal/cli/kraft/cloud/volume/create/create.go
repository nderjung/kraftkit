// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package create

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kraftcloud "sdk.kraft.cloud"
	kraftcloudvolumes "sdk.kraft.cloud/volumes"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/iostreams"
	"kraftkit.sh/log"
)

type CreateOptions struct {
	Auth   *config.AuthConfig               `noattribute:"true"`
	Client kraftcloudvolumes.VolumesService `noattribute:"true"`
	SizeMB int                              `local:"true" long:"size" short:"s" usage:"Size in MB"`

	metro string
}

// Create a KraftCloud persistent volume.
func Create(ctx context.Context, opts *CreateOptions, args ...string) (*kraftcloudvolumes.Volume, error) {
	var err error

	if opts == nil {
		opts = &CreateOptions{}
	}

	if opts.Auth == nil {
		opts.Auth, err = config.GetKraftCloudLoginFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve credentials: %w", err)
		}
	}

	if opts.Client == nil {
		opts.Client = kraftcloud.NewVolumesClient(
			kraftcloud.WithToken(opts.Auth.Token),
		)
	}

	return opts.Client.WithMetro(opts.metro).Create(ctx, kraftcloudvolumes.VolumeCreateRequest{
		SizeMB: opts.SizeMB,
	})
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&CreateOptions{}, cobra.Command{
		Short:   "Create a persistent volume",
		Use:     "create [FLAGS]",
		Args:    cobra.NoArgs,
		Aliases: []string{"new"},
		Example: heredoc.Doc(`
			# Create a new persistent 100MiB volume 
			$ kraft cloud volume create --size 100
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

func (opts *CreateOptions) Pre(cmd *cobra.Command, _ []string) error {
	if opts.SizeMB == 0 {
		return fmt.Errorf("must specify --size flag")
	}

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

func (opts *CreateOptions) Run(ctx context.Context, args []string) error {
	volume, err := Create(ctx, opts, args...)
	if err != nil {
		return fmt.Errorf("could not create volume: %w", err)
	}

	fmt.Fprintf(iostreams.G(ctx).Out, "%s\n", volume.UUID)

	return nil
}
