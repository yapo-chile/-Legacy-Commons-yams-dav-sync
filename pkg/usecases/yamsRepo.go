package usecases

import (
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

// YamsRepositoryError erros that could happen in yams repo
type YamsRepositoryError struct {
	ErrorString string
}

// YamsGetResponse represents yams response for a list of objects
type YamsGetResponse struct {
	ContinuationToken string       `json:"continuation_token"`
	Images            []YamsObject `json:"objects"`
}

// YamsObject is a representation of objects contained by yams bucket
type YamsObject struct {
	ID           string `json:"object_id"`
	Md5          string `json:"md5"`
	Size         int    `json:"size"`
	LastModified int    `json:"last_modified"`
}

// Error parse the yams error response into string
func (pe *YamsRepositoryError) Error() string { return pe.ErrorString }

var (
	// ErrYamsDuplicate is returned by the Put method of YamsRepository
	// implementations to indicate that an object with the same name already
	// exists in Yams.
	ErrYamsDuplicate = &YamsRepositoryError{"object with the same name already exists"}

	// ErrYamsInternal is returned by any method of YamsRepository
	// implementations to indicate that an internal error has occured.
	ErrYamsInternal = &YamsRepositoryError{"internal error"}

	// ErrYamsImage is returned by the Put method of YamsRepository
	// implementations to indicate that it failed to read the image.
	ErrYamsImage = &YamsRepositoryError{"image error"}

	// ErrYamsConnection is returned by any method of YamsRepository
	// implementations to indicate that it failed to connect with Yams.
	ErrYamsConnection = &YamsRepositoryError{"connection error"}

	// ErrYamsUnauthorized is returned by any method of YamsRepository
	// implementations to indicate that it failed to authenticate with Yams.
	ErrYamsUnauthorized = &YamsRepositoryError{"unauthorized error"}

	// ErrYamsBucketNotFound is returned by any method of YamsRepository
	// implementations to indicate that it failed to locate the bucket.
	ErrYamsBucketNotFound = &YamsRepositoryError{"bucket not found"}

	// ErrYamsObjectNotFound is returned by any method of YamsRepository
	// implementations to indicate that it failed to locate the object.
	ErrYamsObjectNotFound = &YamsRepositoryError{"object not found"}
)

// YamsRepository interface that allows yams repository operations
type YamsRepository interface {
	GetImages() ([]YamsObject, *YamsRepositoryError)
	PutImage(domain.Image) *YamsRepositoryError
	HeadImage(imageName string) *YamsRepositoryError
	DeleteImage(imageName string, immediateRemoval bool) *YamsRepositoryError
}
