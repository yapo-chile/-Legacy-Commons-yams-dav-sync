package usecases

// ImageStatusRepo allows ImageStatusRepository's operations
type ImageStatusRepo interface {
	GetImageStatus(key string) (bool, error)
	SetImageStatus(listID string, adCachedStatus bool) error
}
