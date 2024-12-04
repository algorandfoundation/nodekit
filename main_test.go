package main

import (
	"github.com/spf13/viper"
	"testing"
)

func Test_Main(t *testing.T) {
	viper.Set("algod-token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("algod-endpoint", "http://localhost:8080")
	main()
}
