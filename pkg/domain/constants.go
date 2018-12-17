package domain

const (
<<<<<<< HEAD
	// SWUpload send worker is uploading images for first time
	SWUpload = iota
=======
	// SWUplaod send worker is uploading images for first time
	SWUplaod = iota
>>>>>>> [feat/error-control] - Added sendWorker upload & retry constants
	// SWRetry send worker retries to upload images because previous errors
	SWRetry
)
