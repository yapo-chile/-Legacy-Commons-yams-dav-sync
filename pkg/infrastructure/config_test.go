package infrastructure

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Nested struct {
	F bool `env:"LE_F"`
}

type TestConf struct {
	I  int    `env:"LE_I"`
	S  string `env:"LE_S"`
	F  string `env:"FROM"`
	N  Nested `env:"NESTED_"`
	D  string `env:"DEF" envDefault:"default_conf"`
	OF string `env:"OTHERFILE"`
}

func TestConfigLoad(t *testing.T) {
	env := map[string]string{
		"LE_I":           "42",
		"LE_S":           "Don't panic",
		"NESTED_LE_F":    "true",
		"FROM_FILE":      "testdata/from.data",
		"OTHERFILE_FILE": "testdata/not.data",
	}
	// Setup environment
	for k, v := range env {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	var conf TestConf
	LoadFromEnv(&conf)

	expected := TestConf{
		I: 42,
		S: "Don't panic",
		F: "fullhd",
		N: Nested{
			F: true,
		},
		D: "default_conf",
	}

	assert.Equal(t, expected, conf)
}
