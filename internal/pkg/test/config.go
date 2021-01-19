package test

import (
	"strings"

	"github.com/spf13/viper"
)

//NewConfig creates new config from yaml
func NewConfig(yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	res.ReadConfig(strings.NewReader(yaml))
	return res
}
