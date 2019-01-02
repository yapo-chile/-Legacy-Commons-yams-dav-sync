package usecases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	repo := YamsRepositoryError{ErrorString: "err"}
	result := repo.Error()
	assert.Equal(t, result, "err")
}
