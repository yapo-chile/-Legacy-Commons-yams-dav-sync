package domain

const (
	// SWUpload send worker is uploading images for first time
	SWUpload = iota
	// SWRetry send worker retries to upload images because previous errors
	SWRetry
)
