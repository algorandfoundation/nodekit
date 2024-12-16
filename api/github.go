package api

import (
	"encoding/json"
	"errors"
	"strings"
)

const ChannelNotFoundMsg = "channel not found"

type GithubVersionResponse struct {
	ResponseCode   int
	ResponseStatus string
	JSON200        string
}

func (r GithubVersionResponse) StatusCode() int {
	return r.ResponseCode
}
func (r GithubVersionResponse) Status() string {
	return r.ResponseStatus
}

func GetGoAlgorandReleaseWithResponse(http HttpPkgInterface, channel string) (*GithubVersionResponse, error) {
	var versions GithubVersionResponse
	resp, err := http.Get("https://api.github.com/repos/algorand/go-algorand/releases")
	if resp == nil || err != nil {
		return nil, err
	}
	// Update Model
	versions.ResponseCode = resp.StatusCode
	versions.ResponseStatus = resp.Status

	// Exit if not 200
	if resp.StatusCode != 200 {
		return &versions, nil
	}

	defer resp.Body.Close()

	// Parse the versions to a map
	var versionsMap []map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&versionsMap); err != nil {
		return &versions, err
	}
	// Look for versions in the response
	var versionResponse *string
	for i := range versionsMap {
		tn := versionsMap[i]["tag_name"].(string)
		if strings.Contains(tn, channel) {
			versionResponse = &tn
			break
		}

	}

	// If the tag was not found, return an error
	if versionResponse == nil {
		return &versions, errors.New(ChannelNotFoundMsg)
	}

	// Update the JSON200 data and return
	versions.JSON200 = *versionResponse
	return &versions, nil
}
