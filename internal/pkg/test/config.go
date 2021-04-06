package test

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

//NewConfig creates new config from yaml
func NewConfig(t *testing.T, yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	err := res.ReadConfig(strings.NewReader(yaml))
	if !assert.Nil(t, err) {
		assert.Nil(t, err, err.Error())
	}
	return res
}
