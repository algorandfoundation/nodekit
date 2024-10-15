package main

import (
	"github.com/spf13/viper"
	"testing"
)

func Test_Main(t *testing.T) {
	viper.Set("server", "https://mainnet-api.4160.nodely.dev")
	main()
}
