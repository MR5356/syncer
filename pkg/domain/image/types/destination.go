package types

import (
	"context"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/pkg/blobinfocache/none"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"io"
)

type ImageDestination struct {
	destination types.ImageDestination
	ctx         context.Context
	sysCtx      *types.SystemContext
}

func NewImageDestination(registry, repository, tagOrDigest, username, password string, insecure bool) (*ImageDestination, error) {
	srcRef, err := docker.ParseReference("//" + registry + "/" + repository + parseTagOrDigest(tagOrDigest))
	if err != nil {
		return nil, err
	}

	var sysCtx *types.SystemContext
	if insecure {
		sysCtx = &types.SystemContext{
			DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
		}
	} else {
		sysCtx = &types.SystemContext{}
	}

	ctx := context.WithValue(context.Background(), CTXKey("ImageDestination"), repository)

	if username != "" && password != "" {
		sysCtx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	destination, err := srcRef.NewImageDestination(ctx, sysCtx)
	return &ImageDestination{
		destination: destination,
		ctx:         ctx,
		sysCtx:      sysCtx,
	}, nil
}

func (i *ImageDestination) PutManifest(manifestBytes []byte, instanceDigest *digest.Digest) error {
	return i.destination.PutManifest(i.ctx, manifestBytes, instanceDigest)
}

func (i *ImageDestination) PutBlob(blob io.ReadCloser, blobInfo types.BlobInfo) error {
	exist, err := i.CheckBlobExist(blobInfo)
	if err != nil {
		return err
	}
	if exist {
		logrus.Infof("blob %s already exist, skipping", blobInfo.Digest)
		return nil
	}
	_, err = i.destination.PutBlob(i.ctx, blob, types.BlobInfo{
		Digest: blobInfo.Digest,
		Size:   blobInfo.Size,
	}, none.NoCache, true)
	defer blob.Close()

	return err
}

func (i *ImageDestination) CheckBlobExist(blobInfo types.BlobInfo) (bool, error) {
	exist, _, err := i.destination.TryReusingBlob(i.ctx, types.BlobInfo{
		Digest: blobInfo.Digest,
		Size:   blobInfo.Size,
	}, none.NoCache, false)
	return exist, err
}

func (i *ImageDestination) Close() error {
	return i.destination.Close()
}
