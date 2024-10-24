package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func GetGoAlgorandRelease(channel string) (*string, error) {
	resp, err := http.Get("https://api.github.com/repos/algorand/go-algorand/releases")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var versions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		log.Fatal("ooopsss! an error occurred, please try again")
	}
	var versionResponse *string
	for i := range versions {
		tn := versions[i]["tag_name"].(string)
		if strings.Contains(tn, channel) {
			versionResponse = &tn
			break
		}

	}

	return versionResponse, nil
}
