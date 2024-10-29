package main

import (
	"github.com/spf13/viper"
	"testing"
)

func Test_Main(t *testing.T) {
	viper.Set("token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("server", "http://localhost:8080")
	main()
}
