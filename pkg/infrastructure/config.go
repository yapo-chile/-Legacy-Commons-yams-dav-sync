package infrastructure

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// LoggerConf holds configuration for logging
// LogLevel definition:
//   0 - Debug
//   1 - Info
//   2 - Warning
//   3 - Error
//   4 - Critic
type LoggerConf struct {
	SyslogIdentity string `env:"SYSLOG_IDENTITY" envDefault:"true"`
	SyslogEnabled  bool   `env:"SYSLOG_ENABLED" envDefault:"false"`
	StdlogEnabled  bool   `env:"STDLOG_ENABLED" envDefault:"true"`
	LogLevel       int    `env:"LOG_LEVEL" envDefault:"0"`
}

// Config holds all configuration for the service
type Config struct {
	LoggerConf       LoggerConf   `env:"LOGGER_"`
	LocalStorageConf LocalStorage `env:"LOCAL_"`
	YamsConf         YamsConf     `env:"YAMS_"`
}

type LocalStorage struct {
	Path string `env:""`
}

type YamsConf struct {
	MgmtURL        string `env:"MGMT_URL" envDefault:"mgmt-us-east-1-yams.schibsted.com"`
	AccessKeyID    string `env:"ACCESS_KEY_ID" envDefault:""`
	TenantID       string `env:"TENTAND_ID" envDefault:""`
	DomainID       string `env:"DOMAIN_ID" envDefault:""`
	BucketID       string `env:"BUCKET_ID" envDefault:""`
	PrivateKeyFile string `env:"PRIVATE_KEY" envDefault:"writer-key.rsa"`
}

// LoadFromEnv loads the config data from the environment variables
func LoadFromEnv(data interface{}) {
	load(reflect.ValueOf(data), "", "")
}

// recursiveExpandEnv recursively expands any nested env variables
// present in the `s` variable, a nested env variable can be of the form
// "ENDPOINT_PATH=${BASE_PATH}/endpoint" for example
func recursiveExpandEnv(s string) (r string) {
	r = os.ExpandEnv(s)
	if r != s {
		r = recursiveExpandEnv(r)
	}
	return
}

// valueFromEnv lookup the best value for a variable on the environment
func valueFromEnv(envTag, envDefault string) string {
	// Maybe it's a secret and <envTag>_FILE points to a file with the value
	// https://rancher.com/docs/rancher/v1.6/en/cattle/secrets/#docker-hub-images
	if fileName, ok := os.LookupEnv(fmt.Sprintf("%s_FILE", envTag)); ok {
		b, err := ioutil.ReadFile(fileName) // nolint
		if err == nil {
			return string(b)
		}
		fmt.Print(err)
	}
	// The value might be set directly on the environment
	if value, ok := os.LookupEnv(envTag); ok {
		return value
	}
	// Nothing to do, return the default
	return envDefault
}

// load the variable defined in the envTag into Value
func load(conf reflect.Value, envTag, envDefault string) {
	if conf.Kind() == reflect.Ptr {
		reflectedConf := reflect.Indirect(conf)
		// Only attempt to set writeable variables
		if reflectedConf.IsValid() && reflectedConf.CanSet() {
			value := valueFromEnv(envTag, envDefault)
			// Print message if config is missing
			if envTag != "" && value == "" && !strings.HasSuffix(envTag, "_") {
				fmt.Printf("Config for %s missing\n", envTag)
			}
			value = recursiveExpandEnv(value)
			switch reflectedConf.Kind() {
			case reflect.Struct:
				// Recursively load inner struct fields
				for i := 0; i < reflectedConf.NumField(); i++ {
					if tag, ok := reflectedConf.Type().Field(i).Tag.Lookup("env"); ok {
						def, _ := reflectedConf.Type().Field(i).Tag.Lookup("envDefault")
						load(reflectedConf.Field(i).Addr(), envTag+tag, def)
					}
				}
			// Here for each type we should make a cast of the env variable and then set the value
			case reflect.String:
				reflectedConf.SetString(value)
			case reflect.Int:
				if value, err := strconv.Atoi(value); err == nil {
					reflectedConf.Set(reflect.ValueOf(value))
				}
			case reflect.Bool:
				if value, err := strconv.ParseBool(value); err == nil {
					reflectedConf.Set(reflect.ValueOf(value))
				}
			}
		}
	}
}
