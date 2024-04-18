// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package ninepfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/uuid"

	volumev1alpha1 "kraftkit.sh/api/volume/v1alpha1"
	"kraftkit.sh/config"
	"kraftkit.sh/log"
)

type v1alpha1Volume struct{}

func NewVolumeServiceV1alpha1(ctx context.Context, opts ...any) (volumev1alpha1.VolumeService, error) {
	return &v1alpha1Volume{}, nil
}

// Create implements kraftkit.sh/api/volume/v1alpha1.Create
func (*v1alpha1Volume) Create(ctx context.Context, volume *volumev1alpha1.Volume) (*volumev1alpha1.Volume, error) {
	if len(volume.Spec.Driver) == 0 {
		volume.Spec.Driver = "9pfs"
	} else if volume.Spec.Driver != "9pfs" {
		return volume, fmt.Errorf("cannot use 9pfs driver when driver set to %s", volume.Spec.Driver)
	}

	if volume.ObjectMeta.UID == "" {
		volume.ObjectMeta.UID = uuid.NewUUID()
	}

	if len(volume.Spec.Source) == 0 {
		// If no Source is specified, create a new volume entry in the runtime store
		log.G(ctx).Debugf("creating new volume entry in the runtime store %s", volume.ObjectMeta.UID)
		volume.Spec.Source = filepath.Join(config.G[config.KraftKit](ctx).RuntimeDir, "volumes", string(volume.ObjectMeta.UID))
	}

	// TODO turn into absolute path
	// volume.Spec.Source = volume.Spec.Source

	// Create the volume directory if it does not exist
	if err := os.MkdirAll(volume.Spec.Source, 0755); err != nil {
		return volume, fmt.Errorf("cannot create volume directory: %w", err)
	}

	if _, err := os.Stat(volume.Spec.Source); err != nil {
		return volume, fmt.Errorf("cannot stat host path volume: %w", err)
	}

	return volume, nil
}

// Delete implements kraftkit.sh/api/volume/v1alpha1.Delete
func (*v1alpha1Volume) Delete(_ context.Context, _ *volumev1alpha1.Volume) (*volumev1alpha1.Volume, error) {
	return nil, nil
}

// Get implements kraftkit.sh/api/volume/v1alpha1.Get
func (*v1alpha1Volume) Get(_ context.Context, volume *volumev1alpha1.Volume) (*volumev1alpha1.Volume, error) {
	if len(volume.Spec.Driver) == 0 || volume.Spec.Driver != "9pfs" {
		return nil, nil
	}

	if len(volume.Spec.Source) == 0 {
		return nil, nil
	}

	return volume, nil
}

// List implements kraftkit.sh/api/volume/v1alpha1.List
func (*v1alpha1Volume) List(_ context.Context, volumes *volumev1alpha1.VolumeList) (*volumev1alpha1.VolumeList, error) {
	return volumes, nil
}

// List implements kraftkit.sh/api/volume/v1alpha1.List
func (*v1alpha1Volume) Update(_ context.Context, volumes *volumev1alpha1.VolumeList) (*volumev1alpha1.VolumeList, error) {
	return volumes, nil
}

// Watch implements kraftkit.sh/api/volume/v1alpha1.Watch
func (*v1alpha1Volume) Watch(context.Context, *volumev1alpha1.Volume) (chan *volumev1alpha1.Volume, chan error, error) {
	panic("not implemented: kraftkit.sh/machine/volume/9pfs.v1alpha1Volume.Watch")
}
