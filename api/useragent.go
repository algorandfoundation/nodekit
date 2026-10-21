package api

import (
	"context"
	"net/http"
)

// version is the nodekit build version reported in the User-Agent. It stays
// "dev" for anything not built from a release: the linker stamps the real
// value into main.version, and main hands it here at startup.
var version = "dev"

// SetVersion records the build version that outgoing requests report. Call it
// once, before any request is made; it is not safe to call concurrently with
// UserAgent.
func SetVersion(v string) {
	if v == "" {
		return
	}
	version = v
}

// UserAgent is the value nodekit identifies itself with on every outbound
// request, so that the services it talks to (the GitHub release API, the
// catchpoint and short-link services, a node's own algod API) can tell nodekit
// apart from the other clients calling them, and tell releases apart from each
// other.
func UserAgent() string {
	return "Nodekit v" + version
}

// SetUserAgent stamps the nodekit User-Agent onto a request. Without it the
// request goes out as Go's default "Go-http-client/1.1", which says nothing
// about which program made it.
func SetUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", UserAgent())
}

// UserAgentEditor is SetUserAgent as a RequestEditorFn, for the generated
// Algod client. That client is regenerated from the algod OpenAPI spec, so the
// header has to be added at construction time rather than in api/lf.go, where
// `make generate` would overwrite it.
func UserAgentEditor(ctx context.Context, req *http.Request) error {
	SetUserAgent(req)
	return nil
}
