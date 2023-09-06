package task

import (
	"errors"
	"fmt"
	"github.com/MR5356/syncer/pkg/image/config"
	"github.com/MR5356/syncer/pkg/image/types"
	"github.com/MR5356/syncer/pkg/utils/imageutil"
	"github.com/containers/image/v5/manifest"
	types2 "github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type SyncTask struct {
	name        string
	source      string
	destination string

	ch chan struct{}

	getAuthFunc func(repo string) *config.Auth
}

type Sync struct {
	source      *types.ImageSource
	destination *types.ImageDestination
}

func NewSyncTask(source, destination string, getAuthFunc func(repo string) *config.Auth, ch chan struct{}) *SyncTask {
	return &SyncTask{
		name:        fmt.Sprintf("%s -> %s", source, destination),
		source:      source,
		destination: destination,

		ch: ch,

		getAuthFunc: getAuthFunc,
	}
}

func GenerateSyncTaskList(cfg *config.Config, ch chan struct{}) (*List, error) {
	list := NewTaskList()
	for source, dest := range cfg.Images {
		if destList, ok := dest.([]any); ok {
			if len(destList) == 0 {
				return nil, fmt.Errorf("empty destination for source: %s", source)
			}
			for _, d := range destList {
				if destStr, ok := d.(string); ok {
					logrus.Infof("generate sync task: %s -> %s", source, destStr)
					list.Add(NewSyncTask(source, destStr, cfg.GetAuth, ch))
				} else {
					return nil, fmt.Errorf("invalid destination type: %T", d)
				}
			}
		} else if destStr, ok := dest.(string); ok {
			if destStr == "" {
				return nil, fmt.Errorf("empty destination for source: %s", source)
			}
			logrus.Infof("generate sync task: %s -> %s", source, destStr)
			list.Add(NewSyncTask(source, destStr, cfg.GetAuth, ch))
		} else {
			return nil, fmt.Errorf("invalid destination, should be string or []string for source: %s", source)
		}
	}
	return list, nil
}

func (t *SyncTask) Name() string {
	return t.name
}

func (t *SyncTask) Run() error {
	// 支持的镜像同步规则
	// 源镜像【包含tag或digest】 -> 目标镜像【包含/不包含tag或digest】：镜像对应的tag或digest都会同步至目标镜像对应的tag或digest，不包含则表示使用源tag
	// 源镜像【不包含tag或digest】-> 目标镜像：镜像所有的tag都会同步至目标镜像
	srcImageInfo, err := imageutil.ParseImageInfo(t.source)
	if err != nil {
		return err
	}
	logrus.Debugf("source image info: %+v", srcImageInfo)

	destImageInfo, err := imageutil.ParseImageInfo(t.destination)
	if err != nil {
		return err
	}
	logrus.Debugf("destination image info: %+v", destImageInfo)

	syncList := make([]*Sync, 0)

	if srcImageInfo.TagOrDigest != "" {
		logrus.Debugf("source image info tag or digest: %s", srcImageInfo.TagOrDigest)
		srcAuth := t.getAuthFunc(srcImageInfo.Registry)
		srcRef, err := types.NewImageSource(srcImageInfo.Registry, srcImageInfo.GetRepo(), srcImageInfo.TagOrDigest, srcAuth.Username, srcAuth.Password, srcAuth.Insecure)
		if err != nil {
			return err
		}

		destAuth := t.getAuthFunc(destImageInfo.Registry)
		if destImageInfo.TagOrDigest == "" {
			destImageInfo.TagOrDigest = srcImageInfo.TagOrDigest
		}
		destRef, err := types.NewImageDestination(destImageInfo.Registry, destImageInfo.GetRepo(), destImageInfo.TagOrDigest, destAuth.Username, destAuth.Password, destAuth.Insecure)
		if err != nil {
			return err
		}

		syncList = append(syncList, &Sync{
			source:      srcRef,
			destination: destRef,
		})
	} else {
		logrus.Debugf("source image info tag or digest is empty")
		srcAuth := t.getAuthFunc(srcImageInfo.Registry)
		destAuth := t.getAuthFunc(destImageInfo.Registry)

		src, err := types.NewImageSource(srcImageInfo.Registry, srcImageInfo.GetRepo(), srcImageInfo.TagOrDigest, srcAuth.Username, srcAuth.Password, srcAuth.Insecure)
		if err != nil {
			return err
		}
		tags, err := src.GetTags()
		if err != nil {
			return err
		}
		logrus.Infof("source image tags: %+v", tags)

		group := new(errgroup.Group)

		for _, tag := range tags {
			t.ch <- struct{}{}

			tag := tag
			group.Go(func() error {
				defer func() {
					<-t.ch
				}()
				srcRef, err := types.NewImageSource(srcImageInfo.Registry, srcImageInfo.GetRepo(), tag, srcAuth.Username, srcAuth.Password, srcAuth.Insecure)
				if err != nil {
					return err
				}

				destRef, err := types.NewImageDestination(destImageInfo.Registry, destImageInfo.GetRepo(), tag, destAuth.Username, destAuth.Password, destAuth.Insecure)
				if err != nil {
					return err
				}

				syncList = append(syncList, &Sync{
					source:      srcRef,
					destination: destRef,
				})

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			logrus.Errorf("err: %+v", err)
			return err
		}
	}

	logrus.Debugf("s list: %+v", syncList)

	for _, s := range syncList {
		mf, manifestType, err := s.source.GetManifest()
		if err != nil {
			return err
		}
		logrus.Infof("parsing manifest...")
		mfObj, mfBytes, subMfs, err := GetManifests(mf, manifestType, s.source, nil)
		if err != nil {
			return err
		}
		if mfObj == nil {
			return errors.New("invalid manifest")
		}

		if len(subMfs) == 0 {
			blobInfos, err := s.source.GetBlobs(mfObj.(manifest.Manifest))
			if err != nil {
				return err
			}

			group := new(errgroup.Group)
			for _, info := range blobInfos {
				info := info

				t.ch <- struct{}{}
				group.Go(func() error {
					defer func() {
						<-t.ch
					}()
					return transBlob(s.source, s.destination, info)
				})
				if err := group.Wait(); err != nil {
					logrus.Errorf("err: %+v", err)
					return err
				}
			}
		} else {
			group := new(errgroup.Group)
			for _, mfInfo := range subMfs {
				mfInfo := mfInfo

				t.ch <- struct{}{}
				group.Go(func() error {
					defer func() {
						<-t.ch
					}()
					blobInfos, err := s.source.GetBlobs(mfInfo.Obj)
					if err != nil {
						return err
					}
					for _, info := range blobInfos {
						err := transBlob(s.source, s.destination, info)
						if err != nil {
							return err
						}
					}
					err = s.destination.PutManifest(mfInfo.Bytes, mfInfo.Digest)
					if err != nil {
						return err
					}
					return nil
				})
			}
			if err := group.Wait(); err != nil {
				logrus.Errorf("err: %+v", err)
				return err
			}
		}
		err = s.destination.PutManifest(mfBytes, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func transBlob(source *types.ImageSource, destination *types.ImageDestination, info types2.BlobInfo) error {
	logrus.Infof("trans blob: %s", info.Digest)
	exist, err := destination.CheckBlobExist(info)
	if err != nil {
		return err
	}
	if exist {
		logrus.Infof("blob %s already exist, skipping", info.Digest)
		return nil
	}
	blob, size, err := source.GetBlob(info)
	if err != nil {
		return err
	}
	info.Size = size
	if err := destination.PutBlob(blob, info); err != nil {
		return err
	}
	logrus.Infof("trans blob: %s success", info.Digest)
	return nil
}

type ManifestInfo struct {
	Obj    manifest.Manifest
	Digest *digest.Digest

	Bytes []byte
}

func GetManifests(manifestBytes []byte, manifestType string, source *types.ImageSource, parent *manifest.Schema2List) (interface{}, []byte, []*ManifestInfo, error) {
	switch manifestType {
	case manifest.DockerV2Schema2MediaType:
		manifestObj, err := manifest.Schema2FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}
		return manifestObj, manifestBytes, nil, nil
	case manifest.DockerV2Schema1MediaType, manifest.DockerV2Schema1SignedMediaType:
		manifestObj, err := manifest.Schema1FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}
		return manifestObj, manifestBytes, nil, nil
	case specsv1.MediaTypeImageManifest:
		manifestObj, err := manifest.OCI1FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}
		return manifestObj, manifestBytes, nil, nil
	case manifest.DockerV2ListMediaType:
		var subManifestInfoSlice []*ManifestInfo
		manifestObj, err := manifest.Schema2ListFromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}
		for index, manifestDesElem := range manifestObj.Manifests {
			mfBytes, mfType, err := source.GetSource().GetManifest(source.GetCtx(), &manifestDesElem.Digest)
			if err != nil {
				return nil, nil, nil, err
			}
			subManifest, _, _, err := GetManifests(mfBytes, mfType, source, manifestObj)
			if err != nil {
				return nil, nil, nil, err
			}
			if subManifest != nil {
				subManifestInfoSlice = append(subManifestInfoSlice, &ManifestInfo{
					Obj:    subManifest.(manifest.Manifest),
					Digest: &manifestObj.Manifests[index].Digest,
					Bytes:  mfBytes,
				})
			}
		}
		newManifestBytes, _ := manifestObj.Serialize()
		return manifestObj, newManifestBytes, subManifestInfoSlice, nil
	case specsv1.MediaTypeImageIndex:
		var subManifestInfoSlice []*ManifestInfo
		manifestObj, err := manifest.OCI1IndexFromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}
		for index, manifestDesElem := range manifestObj.Manifests {
			mfBytes, mfType, err := source.GetSource().GetManifest(source.GetCtx(), &manifestDesElem.Digest)
			if err != nil {
				return nil, nil, nil, err
			}
			subManifest, _, _, err := GetManifests(mfBytes, mfType, source, nil)
			if err != nil {
				return nil, nil, nil, err
			}
			if subManifest != nil {
				subManifestInfoSlice = append(subManifestInfoSlice, &ManifestInfo{
					Obj:    subManifest.(manifest.Manifest),
					Digest: &manifestObj.Manifests[index].Digest,
					Bytes:  mfBytes,
				})
			}
		}
		newManifestBytes, _ := manifestObj.Serialize()
		return manifestObj, newManifestBytes, subManifestInfoSlice, nil
	default:
		return nil, nil, nil, fmt.Errorf("invalid manifest type: %s", manifestType)
	}
}
