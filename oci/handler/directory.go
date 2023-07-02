// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package handler

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type DirectoryHandler struct {
	path string
}

func NewDirectoryHandler(ctx context.Context, path string) (context.Context, *DirectoryHandler, error) {
	return ctx, &DirectoryHandler{path: path}, nil
}

// DigestExists implements DigestResolver.
func (handle *DirectoryHandler) DigestExists(ctx context.Context, dgst digest.Digest) (exists bool, err error) {
	manifests, err := handle.ListManifests(ctx)
	if err != nil {
		return false, err
	}

	for _, manifest := range manifests {
		if manifest.Config.Digest == dgst {
			return true, nil
		}
	}

	return false, nil
}

// ListManifests implements DigestResolver.
func (handle *DirectoryHandler) ListManifests(ctx context.Context) (manifests []ocispec.Manifest, err error) {
	// Iterate over the manifest directory
	manifests_path := handle.path + "/manifests"

	// Open the directory
	dir, err := os.Open(manifests_path)
	if err != nil {
		return nil, err
	}

	// Close the directory
	defer dir.Close()

	// Read the directory
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	// Iterate over the files
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Skip files that don't end in .json
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read the manifest
		rawManifest, err := os.ReadFile(manifests_path + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		manifest := ocispec.Manifest{}
		err = json.Unmarshal(rawManifest, &manifest)
		if err != nil {
			return nil, err
		}

		// Append the manifest to the list
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

// PushDigest implements DigestPusher.
func (handle *DirectoryHandler) PushDigest(ctx context.Context, ref string, desc ocispec.Descriptor, reader io.Reader, onProgress func(float64)) (err error) {
	// Write to the content store
	// Just write an entry for the digest

	return nil
}

// ResolveImage implements ImageResolver.
func (handle *DirectoryHandler) ResolveImage(ctx context.Context, fullref string) (imgspec ocispec.Image, err error) {
	// Find the manifest of this image
	ref, err := name.ParseReference(fullref)
	if err != nil {
		return ocispec.Image{}, err
	}

	manifest_path := handle.path + "/manifests/" + strings.ReplaceAll(ref.Name(), "/", "-") + ".json"

	// Check whether the manifest exists
	if _, err := os.Stat(manifest_path); err != nil {
		return ocispec.Image{}, fmt.Errorf("manifest for %s does not exist", ref.Name())
	}

	// Read the manifest
	reader, err := os.Open(manifest_path)
	if err != nil {
		return ocispec.Image{}, err
	}
	raw_manifest, err := io.ReadAll(reader)
	if err != nil {
		return ocispec.Image{}, err
	}

	// Unmarshal the manifest
	manifest := ocispec.Manifest{}
	err = json.Unmarshal(raw_manifest, &manifest)
	if err != nil {
		return ocispec.Image{}, err
	}

	// Split the digest into algorithm and hex
	config_name := v1.Hash{
		Algorithm: manifest.Config.Digest.Algorithm().String(),
		Hex:       manifest.Config.Digest.Encoded(),
	}

	// Find the config
	config_path := handle.path + "/configs/" + config_name.Algorithm + "/" + config_name.Hex

	// Check whether the config exists
	if _, err := os.Stat(config_path); err != nil {
		return ocispec.Image{}, fmt.Errorf("config for %s does not exist", ref.Name())
	}

	// Read the config
	reader, err = os.Open(config_path)
	if err != nil {
		return ocispec.Image{}, err
	}
	raw_config, err := io.ReadAll(reader)
	if err != nil {
		return ocispec.Image{}, err
	}

	// Unmarshal the config
	config := ocispec.Image{}
	err = json.Unmarshal(raw_config, &config)
	if err != nil {
		return ocispec.Image{}, err
	}

	// Return the image
	return config, nil
}

// FetchImage implements ImageFetcher.
func (handle *DirectoryHandler) FetchImage(ctx context.Context, fullref string, onProgress func(float64)) (err error) {
	ref, err := name.ParseReference(fullref)

	if err != nil {
		return err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}

	// Write the manifest
	manifest, err := img.RawManifest()
	if err != nil {
		return err
	}

	manifest_path := handle.path + "/manifests/" + strings.ReplaceAll(ref.Name(), "/", "-") + ".json"
	// Recursively create the directory
	err = os.MkdirAll(manifest_path[:strings.LastIndex(manifest_path, "/")], 0755)
	if err != nil {
		return err
	}

	// Open a writer to the specified path
	writer, err := os.Create(manifest_path)
	if err != nil {
		return err
	}

	_, err = writer.Write(manifest)
	if err != nil {
		return err
	}

	config, err := img.RawConfigFile()
	if err != nil {
		return err
	}

	config_name, err := img.ConfigName()
	if err != nil {
		return err
	}

	config_dir := handle.path + "/configs/" + config_name.Algorithm
	config_path := config_dir + "/" + config_name.Hex

	// If the config already exists, skip it
	if _, err := os.Stat(config_path); err == nil {
		return nil
	}

	// Recursively create the directory
	err = os.MkdirAll(config_path[:strings.LastIndex(config_path, "/")], 0755)
	if err != nil {
		return err
	}

	writer, err = os.Create(config_path)
	if err != nil {
		return err
	}

	// Write the config
	_, err = writer.Write(config)
	if err != nil {
		return err
	}

	// Write the layers
	layers, err := img.Layers()
	if err != nil {
		return err
	}

	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return err
		}

		layer_dir := handle.path + "/layers/" + digest.Algorithm
		layer_path := layer_dir + "/" + digest.Hex

		// Recursively create the directory
		err = os.MkdirAll(layer_path[:strings.LastIndex(layer_path, "/")], 0755)
		if err != nil {
			return err
		}

		// If the layer already exists, skip it
		if _, err := os.Stat(layer_path); err == nil {
			continue
		}

		writer, err = os.Create(layer_path)
		if err != nil {
			return err
		}

		reader, err := layer.Compressed()
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}

	}

	return nil
}

// PushImage implements ImagePusher.
func (handle *DirectoryHandler) PushImage(ctx context.Context, ref string, target *ocispec.Descriptor) error {
	return fmt.Errorf("not implemented")
}

// UnpackImage implements ImageUnpacker.
func (handle *DirectoryHandler) UnpackImage(ctx context.Context, ref string, dest string) (err error) {
	img, err := handle.ResolveImage(ctx, ref)
	if err != nil {
		return err
	}

	// Iterate over the layers
	for _, layer := range img.RootFS.DiffIDs {
		// Get the digest
		digest, err := v1.NewHash(layer.String())
		if err != nil {
			return err
		}

		// Get the layer path
		layer_path := handle.path + "/layers/" + digest.Algorithm + "/" + digest.Hex

		// Layer path is a tarball, so we need to extract it
		reader, err := os.Open(layer_path)
		if err != nil {
			return err
		}
		defer reader.Close()

		tr := tar.NewReader(reader)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			// Write the file to the destination
			path := dest + "/" + hdr.Name

			// If the file is a directory, create it
			if hdr.Typeflag == tar.TypeDir {
				err = os.MkdirAll(path, 0755)
				if err != nil {
					return err
				}
				continue
			}

			// If the directory in the path doesn't exist, create it
			if _, err := os.Stat(path[:strings.LastIndex(path, "/")]); os.IsNotExist(err) {
				err = os.MkdirAll(path[:strings.LastIndex(path, "/")], 0755)
				if err != nil {
					return err
				}
			}

			// Otherwise, create the file
			writer, err := os.Create(path)
			if err != nil {
				return err
			}
			defer writer.Close()

			_, err = io.Copy(writer, tr)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

// FinalizeImage implements ImageFinalizer.
func (handle *DirectoryHandler) FinalizeImage(ctx context.Context, image ocispec.Image) error {
	return fmt.Errorf("not implemented: oci.handler.DirectoryHandler.FinalizeImage")
}
