package domain

const (
	// SWUpload send worker is uploading images for first time
	SWUpload = iota
	// SWRetry send worker retries to upload images because previous errors
	SWRetry
)

const (
	// YAMSForceRemoval force inmediate Image removal. Image won't be recoverable.
	YAMSForceRemoval = true
	// YAMSSoftRemoval soft image removal from yams bucket
	YAMSSoftRemoval = false
)

// Metrics exposer constants
const (
	// SentImage represents sent images stat
	SentImages = iota
	// ProcessedImages represents processed images stat
	ProcessedImages
	// Skipeed represents skipped images stat
	SkippedImages
	// NotFoundImages represents not found images stat
	NotFoundImages
	// FailedUploads represents failed uploads  stat
	FailedUploads
	// DuplicatedImages represents duplicated images stat
	DuplicatedImages
	// RecoveredImages represents recovered images stat
	RecoveredImages
	// Total represents total images stat
	TotalImages
)
