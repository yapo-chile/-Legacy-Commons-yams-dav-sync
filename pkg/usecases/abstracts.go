package usecases

// ImageStatusRepository allows ImageStatusRepository's operations
type ImageStatusRepository interface {
	GetImageStatus(key string) (string, error)
	SetImageStatus(key, value string) error
	DelImageStatus(key string) error
}
