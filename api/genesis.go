package api

import (
	"fmt"
	"net/http"
)

type GenesisFileKey string

const (
	MainnetGenesisKey GenesisFileKey = "mainnet"
	TestnetGenesisKey GenesisFileKey = "testnet"
	FnetGenesisKey    GenesisFileKey = "fnet"
)

type GenesisFileResponse struct {
	ResponseCode   int
	ResponseStatus string
	JSON200        string
}

func (r GenesisFileResponse) StatusCode() int {
	return r.ResponseCode
}
func (r GenesisFileResponse) Status() string {
	return r.ResponseStatus
}
func GetGenesis(key GenesisFileKey) {
	var url string
	if key == FnetGenesisKey {
		url = "http://relay-eu-no-1.algorand.green:8184/genesis"
	} else {
		url = fmt.Sprintf("https://raw.githubusercontent.com/algorand/go-algorand/master/installer/genesis/%s/genesis.json", key)
	}
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)

}
