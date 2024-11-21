package internal

import (
	"encoding/json"
	"strings"
)

func GetGoAlgorandRelease(channel string, http HttpPkgInterface) (*string, error) {
	resp, err := http.Get("https://api.github.com/repos/algorand/go-algorand/releases")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var versions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
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
