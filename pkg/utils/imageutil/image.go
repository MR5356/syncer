package imageutil

import "strings"

const (
	defaultRegistry  = "docker.io"
	defaultNamespace = "library"
)

type ImageInfo struct {
	Src string

	Registry    string
	Namespace   string
	Project     string
	TagOrDigest string
}

func (i *ImageInfo) GetRepo() string {
	repo := i.Namespace
	if repo == "" {
		repo = i.Project
	} else {
		repo += "/" + i.Project
	}
	return repo
}

func ParseImageInfo(src string) (*ImageInfo, error) {
	slice := strings.SplitN(src, "/", 3)

	var registry, namespace, project, tagOrDigest string

	// parse project, tagOrDigest
	lastPart := slice[len(slice)-1]
	if strings.Contains(lastPart, "@") {
		project = strings.Split(lastPart, "@")[0]
		tagOrDigest = strings.Split(lastPart, "@")[1]
	} else if strings.Contains(lastPart, ":") {
		project = strings.Split(lastPart, ":")[0]
		tagOrDigest = strings.Split(lastPart, ":")[1]
	} else {
		project = lastPart
		tagOrDigest = ""
	}

	// parse registry
	if len(slice) == 3 {
		registry = slice[0]
		namespace = slice[1]
	} else if len(slice) == 2 {
		if strings.Contains(slice[0], ".") {
			registry = slice[0]
			namespace = ""
		} else {
			registry = defaultRegistry
			namespace = slice[0]
		}
	} else {
		registry = defaultRegistry
		namespace = defaultNamespace
	}

	return &ImageInfo{
		Src:         src,
		Registry:    registry,
		Namespace:   namespace,
		Project:     project,
		TagOrDigest: tagOrDigest,
	}, nil
}
