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
