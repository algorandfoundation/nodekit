package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

type testResponse struct {
	HttpPkgInterface
}

var jsonStr = `[{
    "tag_name": "v3.26.0-beta"
  }]`

func (testResponse) Get(url string) (resp *http.Response, err error) {

	responseBody := io.NopCloser(bytes.NewReader([]byte(jsonStr)))
	return &http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil
}

type testError struct {
	HttpPkgInterface
}

func (testError) Get(url string) (resp *http.Response, err error) {
	return &http.Response{
		StatusCode: 404,
	}, errors.New("not found")
}

func Test_Github(t *testing.T) {
	r, err := GetGoAlgorandReleaseWithResponse(new(testResponse), "beta")
	if err != nil {
		t.Error(err)
	}
	if r.StatusCode() != 200 {
		t.Error("should be 200 response")
	}
	if r.JSON200 != "v3.26.0-beta" {
		t.Error("should return v3.26.0-beta")
	}

	_, err = GetGoAlgorandReleaseWithResponse(new(testError), "beta")
	if err == nil {
		t.Error("should fail to get")
	}
}
