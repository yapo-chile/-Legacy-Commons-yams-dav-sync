package usecases

// LastSyncRepository allows LastSyncRepository's operations
type LastSyncRepository interface {
	GetLastSync() (string, error)
	SetLastSync(value string) error
}

type ErrorControlRepository interface {
	GetErrorSync(nPage int) (result []string, err error)
	GetPagesQty() int
	DelErrorSync(imgPath string) error
	SetErrorSync(imagePath string) (err error)
	SetErrorCounter(imagePath string, count int) error
}
