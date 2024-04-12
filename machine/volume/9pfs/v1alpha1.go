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

	zip "api.zip"
	"k8s.io/apimachinery/pkg/util/uuid"

	volumev1alpha1 "kraftkit.sh/api/volume/v1alpha1"
	"kraftkit.sh/config"
	"kraftkit.sh/log"
	"kraftkit.sh/store"
)

type v1alpha1Volume struct{}

func NewVolumeServiceV1alpha1(ctx context.Context, opts ...any) (volumev1alpha1.VolumeService, error) {
	embeddedStore, err := store.NewEmbeddedStore[volumev1alpha1.VolumeSpec, volumev1alpha1.VolumeStatus](
		filepath.Join(
			config.G[config.KraftKit](ctx).RuntimeDir,
			"volumev1alpha1",
		),
	)

	if err != nil {
		return nil, err
	}

	return volumev1alpha1.NewVolumeServiceHandler(ctx, &v1alpha1Volume{}, zip.WithStore[volumev1alpha1.VolumeSpec, volumev1alpha1.VolumeStatus](embeddedStore, zip.StoreRehydrationNever))
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
		// Create the volume directory if it does not exist
		if err := os.MkdirAll(volume.Spec.Source, 0755); err != nil {
			return volume, fmt.Errorf("cannot create volume directory: %w", err)
		}
	}

	if _, err := os.Stat(volume.Spec.Source); err != nil {
		return volume, fmt.Errorf("cannot stat host path volume: %w", err)
	}
	volume.Status.State = volumev1alpha1.VolumeStateBound

	return volume, nil
}

// Delete implements kraftkit.sh/api/volume/v1alpha1.Delete
func (*v1alpha1Volume) Delete(_ context.Context, _ *volumev1alpha1.Volume) (*volumev1alpha1.Volume, error) {
	return nil, nil
}

// Get implements kraftkit.sh/api/volume/v1alpha1.Get
func (*v1alpha1Volume) Get(_ context.Context, volume *volumev1alpha1.Volume) (*volumev1alpha1.Volume, error) {
	return volume, nil
}

// List implements kraftkit.sh/api/volume/v1alpha1.List
func (*v1alpha1Volume) List(_ context.Context, volumes *volumev1alpha1.VolumeList) (*volumev1alpha1.VolumeList, error) {
	return volumes, nil
}

// Watch implements kraftkit.sh/api/volume/v1alpha1.Watch
func (*v1alpha1Volume) Watch(context.Context, *volumev1alpha1.Volume) (chan *volumev1alpha1.Volume, chan error, error) {
	panic("not implemented: kraftkit.sh/machine/volume/9pfs.v1alpha1Volume.Watch")
}
