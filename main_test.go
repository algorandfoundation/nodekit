package main

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_Main(t *testing.T) {
	viper.Set("algod-token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("algod-endpoint", "http://localhost:8080")
	main()
}
