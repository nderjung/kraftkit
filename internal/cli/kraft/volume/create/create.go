// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file expect in compliance with the License.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	volumeapi "kraftkit.sh/api/volume/v1alpha1"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/machine/volume"
)

type CreateOptions struct {
	Driver string `noattribute:"true"`
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&CreateOptions{}, cobra.Command{
		Short: "Create a machine volume",
		Use:   "create VOLUME",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "volume",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *CreateOptions) Pre(cmd *cobra.Command, _ []string) error {
	opts.Driver = cmd.Flag("driver").Value.String()
	return nil
}

func (opts *CreateOptions) Run(ctx context.Context, args []string) error {
	var err error

	strategy, ok := volume.Strategies()[opts.Driver]
	if !ok {
		return fmt.Errorf("unsupported network driver strategy: %v (contributions welcome!)", opts.Driver)
	}

	controller, err := strategy.NewVolumeV1alpha1(ctx)
	if err != nil {
		return err
	}

	if _, err := controller.Get(ctx, &volumeapi.Volume{
		ObjectMeta: v1.ObjectMeta{
			Name: args[0],
		},
	}); err == nil {
		return fmt.Errorf("volume %s already exists", args[0])
	}

	_, err = controller.Create(ctx, &volumeapi.Volume{
		ObjectMeta: v1.ObjectMeta{
			Name: args[0],
		},
		Spec: volumeapi.VolumeSpec{
			Driver: opts.Driver,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
