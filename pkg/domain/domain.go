package domain

import "time"

// Image is an object image representation
type Image struct {
	Metadata ImageMetadata
	FilePath string
}

// ImageMetadata is an image metadata respresentation
type ImageMetadata struct {
	ImageName string
	Size      int64
	ModTime   time.Time
}
