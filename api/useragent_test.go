package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_UserAgent(t *testing.T) {
	if got := UserAgent(); got != "Nodekit vdev" {
		t.Fatalf("expected an unstamped build to report Nodekit vdev, got %s", got)
	}

	SetVersion("1.2.3")
	defer SetVersion("dev")
	if got := UserAgent(); got != "Nodekit v1.2.3" {
		t.Fatalf("expected Nodekit v1.2.3, got %s", got)
	}

	// An empty value means the linker never stamped one in, and must not
	// produce a bare "Nodekit v".
	SetVersion("")
	if got := UserAgent(); got != "Nodekit v1.2.3" {
		t.Fatalf("expected an empty version to be ignored, got %s", got)
	}
}

// Test_UserAgentRequests covers every outbound request going through the
// shared wrapper: the GitHub release checks, the catchpoint lookup, the short
// link posts and the upgrade download all call Get or Post.
func Test_UserAgentRequests(t *testing.T) {
	SetVersion("1.2.3")
	defer SetVersion("dev")

	var agents []string
	var contentTypes []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agents = append(agents, r.Header.Get("User-Agent"))
		contentTypes = append(contentTypes, r.Header.Get("Content-Type"))
	}))
	defer server.Close()

	resp, err := Http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	resp, err = Http.Post(server.URL, "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if len(agents) != 2 {
		t.Fatalf("expected two requests, got %d", len(agents))
	}
	for _, agent := range agents {
		if agent != "Nodekit v1.2.3" {
			t.Errorf("expected Nodekit v1.2.3, got %s", agent)
		}
	}
	// Post builds its own request, so it has to carry the content type that
	// http.Post would have set.
	if contentTypes[1] != "application/json" {
		t.Errorf("expected application/json, got %s", contentTypes[1])
	}
}

// Test_UserAgentEditor covers the generated Algod client, which sets headers
// through request editors rather than through the wrapper.
func Test_UserAgentEditor(t *testing.T) {
	SetVersion("1.2.3")
	defer SetVersion("dev")

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/v2/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = UserAgentEditor(req.Context(), req); err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("User-Agent"); got != "Nodekit v1.2.3" {
		t.Fatalf("expected Nodekit v1.2.3, got %s", got)
	}
}
