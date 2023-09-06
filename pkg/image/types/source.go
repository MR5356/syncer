package types

import (
	"context"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache/none"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	"io"
)

type ImageSource struct {
	ref types.ImageReference

	source types.ImageSource
	ctx    context.Context
	sysCtx *types.SystemContext
}

type CTXKey string

func NewImageSource(registry, repository, tagOrDigest, username, password string, insecure bool) (*ImageSource, error) {
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

	ctx := context.WithValue(context.Background(), CTXKey("ImageSource"), repository)

	if username != "" && password != "" {
		sysCtx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	var source types.ImageSource

	if tagOrDigest != "" {
		source, err = srcRef.NewImageSource(ctx, sysCtx)
		if err != nil {
			return nil, err
		}
	}

	return &ImageSource{
		ref: srcRef,

		source: source,
		ctx:    ctx,
		sysCtx: sysCtx,
	}, nil

}

func (s *ImageSource) GetManifest() ([]byte, string, error) {
	return s.source.GetManifest(s.ctx, nil)
}

func (s *ImageSource) GetSource() types.ImageSource {
	return s.source
}

func (s *ImageSource) GetCtx() context.Context {
	return s.ctx
}

func (s *ImageSource) GetManifestWithDigest(digest *digest.Digest) ([]byte, string, error) {
	return s.source.GetManifest(s.ctx, digest)

}

func (s *ImageSource) GetTags() ([]string, error) {
	return docker.GetRepositoryTags(s.ctx, s.sysCtx, s.ref)
}

func (s *ImageSource) GetBlob(blobInfo types.BlobInfo) (io.ReadCloser, int64, error) {
	return s.source.GetBlob(s.ctx, types.BlobInfo{
		Digest: blobInfo.Digest,
		URLs:   blobInfo.URLs,
		Size:   -1,
	}, none.NoCache)
}

func (s *ImageSource) GetBlobs(manifests ...manifest.Manifest) ([]types.BlobInfo, error) {
	blobs := make([]types.BlobInfo, 0)
	for _, m := range manifests {
		blobInfo := m.LayerInfos()
		for _, blob := range blobInfo {
			blobs = append(blobs, blob.BlobInfo)
		}

		configBlob := m.ConfigInfo()
		if configBlob.Digest != "" {
			blobs = append(blobs, configBlob)
		}
	}
	return blobs, nil
}

func (s *ImageSource) Close() error {
	return s.source.Close()
}

func parseTagOrDigest(tagOrDigest string) string {
	if tagOrDigest == "" {
		return ""
	}
	d := digest.Digest(tagOrDigest)
	if err := d.Validate(); err != nil {
		return ":" + tagOrDigest
	}
	return "@" + tagOrDigest
}
