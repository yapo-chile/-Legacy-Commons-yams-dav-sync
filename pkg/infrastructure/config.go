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
	LoggerConf       LoggerConf       `env:"LOGGER_"`
	LocalStorageConf LocalStorage     `env:"IMAGES_"`
	YamsConf         YamsConf         `env:"YAMS_"`
	Database         DatabaseConfig   `env:"DATABASE_"`
	ErrorControl     ErrorControlConf `env:"ERRORS_"`
	LastSync         LastSyncConf     `env:"LAST_SYNC_"`
}

// LocalStorage hols all configuration for local storage
type LocalStorage struct {
	Path string `env:"PATH"`
}

type YamsConf struct {
	MgmtURL            string `env:"MGMT_URL" envDefault:"https://mgmt-us-east-1-yams.schibsted.com/api/v1"`
	AccessKeyID        string `env:"ACCESS_KEY_ID"`
	TenantID           string `env:"TENTAND_ID"`
	DomainID           string `env:"DOMAIN_ID"`
	BucketID           string `env:"BUCKET_ID"`
	PrivateKeyFile     string `env:"PRIVATE_KEY" envDefault:"writer-key.rsa"`
	TimeOut            int    `env:"TiMEOUT" envDefault:30`
	MaxConcurrentConns int    `env:"MAX_CONCURRENT_CONN" envDefault:"100"`
}

type ErrorControlConf struct {
	MaxRetriesPerError int `env:"MAX_RETRIES_PER_ERROR" envDefault:"3"`
	MaxResultsPerPage  int `env:"MAX_RESULTS_PER_PAGE" envDefault:"10"`
}

type LastSyncConf struct {
	DefaultDate string `env:"DEFAULT_DATE" envDefault:"31-12-2015"`
}

type DatabaseConfig struct {
	Host        string `env:"HOST" envDefault:"db"`
	Port        int    `env:"PORT" envDefault:"5432"`
	Dbname      string `env:"NAME" envDefault:"pgdb"`
	DbUser      string `env:"USER" envDefault:"postgres"`
	DbPasswd    string `env:"PASSWORD" envDefault:"postgres"`
	Sslmode     string `env:"SSL_MODE" envDefault:"disable"`
	MaxIdle     int    `env:"MAX_IDLE" envDefault:"10"`
	MaxOpen     int    `env:"MAX_OPEN" envDefault:"100"`
	MgFolder    string `env:"MIGRATIONS_FOLDER" envDefault:"migrations"`
	MgDriver    string `env:"MIGRATIONS_DRIVER" envDefault:"postgres"`
	ConnRetries int    `env:"CONN_RETRIES" envDefault:"60"`
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
