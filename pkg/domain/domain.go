package domain

import "time"

type Image struct {
	Metadata ImageMetadata
	FilePath string
}

type ImageMetadata struct {
	ImageName string
	Size      int64
	ModTime   time.Time
}
