// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package initialize

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kraftcloud "sdk.kraft.cloud"
	kraftcloudinstances "sdk.kraft.cloud/instances"
	kraftcloudautoscale "sdk.kraft.cloud/services/autoscale"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/cli/kraft/cloud/utils"
	"kraftkit.sh/log"
	"kraftkit.sh/tui/selection"
)

type InitOptions struct {
	Auth           *config.AuthConfig    `noattribute:"true"`
	Client         kraftcloud.KraftCloud `noattribute:"true"`
	CooldownTimeMs uint                  `long:"cooldown-time-ms" short:"c" usage:"The cooldown time of the configuration" default:"1000"`
	Master         string                `long:"master" short:"i" usage:"The UUID or Name of the master instance"`
	MaxSize        uint                  `long:"max-size" short:"M" usage:"The maximum size of the configuration" default:"10"`
	Metro          string                `noattribute:"true"`
	MinSize        uint                  `long:"min-size" short:"m" usage:"The minimum size of the configuration"`
	WarmupTimeMs   uint                  `long:"warmup-time-ms" short:"w" usage:"The warmup time of the configuration" default:"1000"`
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&InitOptions{}, cobra.Command{
		Short:   "Initialize autoscale configuration for a service group",
		Use:     "initialize [FLAGS] NAME|UUID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"init", "initialise", "i"},
		Long:    "Initialize autoscale configuration for a service group.",
		Example: heredoc.Doc(`
			# Initialize an autoscale configuration
			kraft cloud scale initialize my-service-group \
				--master my-instance-name \
				--min-size 1 \
				--max-size 10 \
				--cooldown-time-ms 1000 \
				--warmup-time-ms 1000
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud-scale",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *InitOptions) Pre(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("specify a service group name or UUID")
	}

	opts.Metro = cmd.Flag("metro").Value.String()
	if opts.Metro == "" {
		opts.Metro = os.Getenv("KRAFTCLOUD_METRO")
	}
	if opts.Metro == "" {
		return fmt.Errorf("kraftcloud metro is unset")
	}

	log.G(cmd.Context()).WithField("metro", opts.Metro).Debug("using")

	return nil
}

func (opts *InitOptions) Run(ctx context.Context, args []string) error {
	var err error

	if opts.Auth == nil {
		opts.Auth, err = config.GetKraftCloudAuthConfigFromContext(ctx)
		if err != nil {
			return fmt.Errorf("could not retrieve credentials: %w", err)
		}
	}

	if opts.Client == nil {
		opts.Client = kraftcloud.NewClient(
			kraftcloud.WithToken(config.GetKraftCloudTokenAuthConfig(*opts.Auth)),
		)
	}

	// Look up the configuration by name
	if !utils.IsUUID(args[0]) {
		conf, err := opts.Client.Services().WithMetro(opts.Metro).GetByName(ctx, args[0])
		if err != nil {
			return fmt.Errorf("could not get configuration: %w", err)
		}

		args[0] = conf.UUID
	}

	master := &kraftcloudinstances.Instance{}

	if opts.Master == "" {
		if config.G[config.KraftKit](ctx).NoPrompt {
			return fmt.Errorf("specify an instance master UUID or name via --master")
		}

		instances, err := opts.Client.Instances().WithMetro(opts.Metro).List(ctx)
		if err != nil {
			return fmt.Errorf("could not list instances: %w", err)
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances found in service group")
		}

		if len(instances) == 1 {
			master.UUID = instances[0].UUID
		} else {
			var possible []kraftcloudinstances.Instance

			for _, instance := range instances {
				if instance.ServiceGroup == nil {
					continue
				}

				if instance.ServiceGroup.UUID != args[0] {
					continue
				}

				possible = append(possible, instance)
			}

			result, err := selection.Select[kraftcloudinstances.Instance](
				"select master instance",
				possible...,
			)
			if err != nil {
				return fmt.Errorf("could not select master instance: %w", err)
			}

			master.UUID = result.UUID
		}
	} else {
		if utils.IsUUID(opts.Master) {
			master.UUID = opts.Master
		} else {
			master.Name = opts.Master
		}
	}

	conf := kraftcloudautoscale.AutoscaleConfiguration{
		UUID:           args[0],
		MinSize:        opts.MinSize,
		MaxSize:        opts.MaxSize,
		WarmupTimeMs:   opts.WarmupTimeMs,
		CooldownTimeMs: opts.CooldownTimeMs,
		Master:         master,
	}

	if utils.IsUUID(args[0]) {
		_, err = opts.Client.
			Autoscale().
			WithMetro(opts.Metro).
			CreateConfigurationByUUID(ctx, args[0], conf)
	} else {
		_, err = opts.Client.
			Autoscale().
			WithMetro(opts.Metro).
			CreateConfigurationByName(ctx, args[0], conf)
	}
	if err != nil {
		return fmt.Errorf("could not create configuration: %w", err)
	}

	return nil
}