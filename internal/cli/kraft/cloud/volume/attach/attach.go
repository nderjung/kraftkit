// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package attach

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

type AttachOptions struct {
	At     string                           `long:"at" usage:"The path the volume should be mounted to"`
	Auth   *config.AuthConfig               `noattribute:"true"`
	Client kraftcloudvolumes.VolumesService `noattribute:"true"`
	To     string                           `long:"to" usage:"The instance the volume should be attached to"`

	metro string
}

// Attach a KraftCloud persistent volume to an instance.
func Attach(ctx context.Context, opts *AttachOptions, args ...string) (*kraftcloudvolumes.Volume, error) {
	var err error

	if opts == nil {
		opts = &AttachOptions{}
	}

	if opts.At == "" {
		return nil, fmt.Errorf("required to set the destination instance")
	}

	if opts.To == "" {
		return nil, fmt.Errorf("required to set the destination path in the instance")
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

	return opts.Client.WithMetro(opts.metro).Attach(ctx, args[0], kraftcloudvolumes.VolumeAttachRequest{
		AttachTo: opts.To,
		At:       opts.At,
	})
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&AttachOptions{}, cobra.Command{
		Short:   "Attach a persistent volume to an instance",
		Use:     "attach [FLAGS]",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"join", "mount"},
		Example: heredoc.Doc(`
		# Attach the volume <VOL_UUID> to the instance <INSTANCE_UUID> to the path /mnt/data
		$ kraft cloud vol attach <VOL_UUID> --to <INSTANCE_UUID> --at /mnt/data
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

func (opts *AttachOptions) Pre(cmd *cobra.Command, _ []string) error {
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

func (opts *AttachOptions) Run(ctx context.Context, args []string) error {
	volume, err := Attach(ctx, opts, args...)
	if err != nil {
		return fmt.Errorf("could not create volume: %w", err)
	}

	fmt.Fprintf(iostreams.G(ctx).Out, "%s\n", volume.UUID)

	return nil
}
