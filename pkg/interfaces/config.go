package interfaces

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	MgmtURL        string
	AccessKeyID    string
	TenantID       string
	DomainID       string
	PrivateKeyFile string
}

func NewConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	config := Config{
		MgmtURL:        viper.GetString("yams.mgmtURL"),
		AccessKeyID:    viper.GetString("yams.accessKeyID"),
		TenantID:       viper.GetString("yams.tenantID"),
		DomainID:       viper.GetString("yams.domainID"),
		PrivateKeyFile: viper.GetString("yams.privateKey"),
	}

	return config, nil
}
