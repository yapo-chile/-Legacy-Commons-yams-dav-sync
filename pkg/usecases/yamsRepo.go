package usecases

import "github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"

type YamsRepositoryError struct {
	ErrorString string
}

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

type YamsRepository interface {
	PutImage(domain.Image) *YamsRepositoryError
}
