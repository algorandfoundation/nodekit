package linux

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Listings of the Algorand repository key before and after the signing
// subkey rotated on 2026-10-31.
const expiredKeyColons = `pub:-:3072:1:EDACD29DA10A4EA6:1560257538:::-:::scSC::::::23::0:
fpr:::::::::611DE94A396F01359C72AF56EDACD29DA10A4EA6:
uid:-::::1560257538::F2553E1583119219B91ACA894E93708FE0FB855A::Algorand developers <dev@algorand.com>::::::::::0:
sub:e:4096:1:007CB592A5D131FE:1667312442:1793456442:::::s::::::23:
fpr:::::::::F321C257B7CC77F631EE89C3007CB592A5D131FE:
`

const rotatedKeyColons = `pub:-:3072:1:EDACD29DA10A4EA6:1560257538:::-:::scSC::::::23::0:
fpr:::::::::611DE94A396F01359C72AF56EDACD29DA10A4EA6:
uid:-::::1560257538::F2553E1583119219B91ACA894E93708FE0FB855A::Algorand developers <dev@algorand.com>::::::::::0:
sub:-:4096:1:F1C9063D88DB5791:1793469233:1919613233:::::s::::::23:
fpr:::::::::A2696A69562A3F158F5DD446F1C9063D88DB5791:
`

// Two days after the rotation.
var afterRotation = time.Unix(1793623957, 0)

func TestParseGPGColons(t *testing.T) {
	certs := parseGPGColons("gpg: keybox created\n" + rotatedKeyColons)
	if len(certs) != 1 {
		t.Fatalf("got %d certs, want 1", len(certs))
	}
	cert := certs[0]
	if want := []string{"Algorand developers <dev@algorand.com>"}; !reflect.DeepEqual(cert.UIDs, want) {
		t.Errorf("UIDs = %q, want %q", cert.UIDs, want)
	}
	if len(cert.Keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(cert.Keys))
	}
	pub, sub := cert.Keys[0], cert.Keys[1]
	if !pub.Primary || pub.Fingerprint != "611DE94A396F01359C72AF56EDACD29DA10A4EA6" || pub.Caps != "sc" || !pub.Expires.IsZero() {
		t.Errorf("unexpected primary key %+v", pub)
	}
	if sub.Primary || sub.Fingerprint != "A2696A69562A3F158F5DD446F1C9063D88DB5791" || sub.Caps != "s" || sub.Algo != "rsa" || sub.Bits != 4096 {
		t.Errorf("unexpected subkey %+v", sub)
	}
	if got := sub.Expires.Format(time.DateOnly); got != "2030-10-30" {
		t.Errorf("subkey expires %s, want 2030-10-30", got)
	}
}

func TestNeedsKeyUpdate(t *testing.T) {
	expired := parseGPGColons(expiredKeyColons)
	rotated := parseGPGColons(rotatedKeyColons)
	extended := parseGPGColons(strings.Replace(expiredKeyColons, "1793456442", "1919613233", 1))

	tests := []struct {
		name          string
		local, remote []pgpCert
		want          bool
	}{
		{"rotated subkey", expired, rotated, true},
		{"same key", rotated, rotated, false},
		{"remote older than installed", rotated, expired, false},
		{"expiry extended", expired, extended, true},
		{"expiry brought forward", extended, expired, true},
		{"subkey revoked", rotated, parseGPGColons(strings.Replace(rotatedKeyColons, "sub:-:", "sub:r:", 1)), true},
		{"revocation already installed", parseGPGColons(strings.Replace(rotatedKeyColons, "sub:-:", "sub:r:", 1)), rotated, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsKeyUpdate(tt.local, tt.remote, afterRotation); got != tt.want {
				t.Errorf("needsKeyUpdate = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestDescribeCerts(t *testing.T) {
	got := describeCerts(parseGPGColons(expiredKeyColons), afterRotation)
	want := `  pub  rsa3072  611D E94A 396F 0135 9C72  AF56 EDAC D29D A10A 4EA6
       Algorand developers <dev@algorand.com>
       signing/certify, created 2019-06-11, does not expire
  sub  rsa4096  F321 C257 B7CC 77F6 31EE  89C3 007C B592 A5D1 31FE
       signing, created 2022-11-01, EXPIRED 2026-10-31
`
	if got != want {
		t.Errorf("describeCerts =\n%s\nwant\n%s", got, want)
	}
}

func TestParseOneLineSources(t *testing.T) {
	content := `# deb [signed-by=/old.gpg] https://releases.algorand.com/deb/ stable main
deb [arch=amd64 signed-by=/usr/share/keyrings/algorand-keyring.gpg] https://releases.algorand.com/deb/ stable main
deb http://archive.ubuntu.com/ubuntu noble main
deb https://releases.algorand.com/deb/ stable main
deb https://releases.algorand.com.evil.example/deb/ stable main
deb http://mirror.example/ubuntu noble main # releases.algorand.com
deb-src https://releases.algorand.com/deb/ stable main
`
	want := []aptSource{
		{File: "a.list", Keyrings: []string{"/usr/share/keyrings/algorand-keyring.gpg"}},
		{File: "a.list"},
	}
	if got := parseOneLineSources("a.list", content); !reflect.DeepEqual(got, want) {
		t.Errorf("parseOneLineSources = %+v, want %+v", got, want)
	}
}

func TestParseDeb822Sources(t *testing.T) {
	content := `Types: deb
URIs: http://archive.ubuntu.com/ubuntu
Suites: noble
Components: main

Types: deb
URIs: https://releases.algorand.com/deb/
Suites: stable
Components: main
Signed-By: /usr/share/keyrings/algorand-keyring.gpg

Enabled: no
Types: deb
URIs: https://releases.algorand.com/deb/
Suites: stable
Components: main
`
	want := []aptSource{{File: "a.sources", Keyrings: []string{"/usr/share/keyrings/algorand-keyring.gpg"}}}
	if got := parseDeb822Sources("a.sources", content); !reflect.DeepEqual(got, want) {
		t.Errorf("parseDeb822Sources = %+v, want %+v", got, want)
	}
}

func TestParseDeb822SourcesWhitespaceSeparator(t *testing.T) {
	content := "Types: deb\nURIs: https://releases.algorand.com/deb/\nSuites: stable\nComponents: main\nSigned-By: /usr/share/keyrings/algorand-keyring.gpg\n \t\nTypes: deb\nURIs: http://archive.ubuntu.com/ubuntu\nSuites: noble\nComponents: main\n"
	want := []aptSource{{File: "a.sources", Keyrings: []string{"/usr/share/keyrings/algorand-keyring.gpg"}}}
	if got := parseDeb822Sources("a.sources", content); !reflect.DeepEqual(got, want) {
		t.Errorf("parseDeb822Sources = %+v, want %+v", got, want)
	}
}

func TestParseDeb822SourcesInlineKey(t *testing.T) {
	content := `Types: deb
URIs: https://releases.algorand.com/deb/
Suites: stable
Components: main
Signed-By:
 -----BEGIN PGP PUBLIC KEY BLOCK-----
 .
 mQENBF...
 /x2Fq9base64LineStartingWithSlash
 -----END PGP PUBLIC KEY BLOCK-----
`
	want := []aptSource{{File: "a.sources", InlineKey: true}}
	if got := parseDeb822Sources("a.sources", content); !reflect.DeepEqual(got, want) {
		t.Errorf("parseDeb822Sources = %+v, want %+v", got, want)
	}
}

func TestFindAlgorandAptSources(t *testing.T) {
	dir := t.TempDir()
	old := aptEtcDir
	aptEtcDir = dir
	t.Cleanup(func() { aptEtcDir = old })

	if err := os.MkdirAll(filepath.Join(dir, "sources.list.d"), 0755); err != nil {
		t.Fatal(err)
	}
	list := filepath.Join(dir, "sources.list.d", "algorand.list")
	if err := os.WriteFile(list, []byte("deb [signed-by=/k.gpg] https://releases.algorand.com/deb/ stable main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := findAlgorandAptSources()
	want := []aptSource{{File: list, Keyrings: []string{"/k.gpg"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findAlgorandAptSources = %+v, want %+v", got, want)
	}
}

// apt-get update output with a stale key and a cached index: apt exits 0.
const staleKeyUpdateOutput = `Hit:1 http://archive.ubuntu.com/ubuntu noble InRelease
Hit:2 https://releases.algorand.com/deb stable InRelease
Err:2 https://releases.algorand.com/deb stable InRelease
  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
Reading package lists...
W: An error occurred during the signature verification. The repository is not updated and the previous index files will be used. GPG error: https://releases.algorand.com/deb stable InRelease: The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
W: Failed to fetch https://releases.algorand.com/deb/dists/stable/InRelease  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
W: Some index files failed to download. They have been ignored, or old ones used instead.
`

func TestAptUpdateIssues(t *testing.T) {
	failures, warnings, _ := aptUpdateIssues(staleKeyUpdateOutput, nil)
	if len(failures) != 2 {
		t.Errorf("got %d failures, want 2: %q", len(failures), failures)
	}
	if !containsAny(strings.Join(failures, "\n"), signatureProblems) {
		t.Error("NO_PUBKEY not recognised as a signature problem")
	}
	if want := []string{"W: Some index files failed to download. They have been ignored, or old ones used instead."}; !reflect.DeepEqual(warnings, want) {
		t.Errorf("warnings = %q, want %q", warnings, want)
	}

	failures, warnings, _ = aptUpdateIssues("Hit:1 https://releases.algorand.com/deb stable InRelease\nReading package lists...\n", nil)
	if len(failures) != 0 || len(warnings) != 0 {
		t.Errorf("clean update reported issues: %q %q", failures, warnings)
	}
}

func TestAptUpdateIssuesBenignRepoWarning(t *testing.T) {
	out := `Hit:1 https://releases.algorand.com/deb stable InRelease
Reading package lists...
W: https://releases.algorand.com/deb/dists/stable/InRelease: Key is stored in legacy trusted.gpg keyring (/etc/apt/trusted.gpg), see the DEPRECATION section in apt-key(8) for details.
`
	failures, warnings, _ := aptUpdateIssues(out, nil)
	if len(failures) != 0 {
		t.Errorf("legacy keyring warning treated as a failure: %q", failures)
	}
	if len(warnings) != 1 {
		t.Errorf("got %d warnings, want 1: %q", len(warnings), warnings)
	}
}

func TestAptUpdateIssuesErrOnly(t *testing.T) {
	out := `Err:1 https://releases.algorand.com/deb stable InRelease
  Could not resolve 'releases.algorand.com'
Hit:2 http://archive.ubuntu.com/ubuntu noble InRelease
Reading package lists...
`
	failures, _, _ := aptUpdateIssues(out, nil)
	want := []string{"Err:1 https://releases.algorand.com/deb stable InRelease", "Could not resolve 'releases.algorand.com'"}
	if !reflect.DeepEqual(failures, want) {
		t.Errorf("failures = %q, want %q", failures, want)
	}
}

func TestInstalledAlgorandKeyringsSkipsGlobalScanForInlineKeys(t *testing.T) {
	dir := t.TempDir()
	old := aptEtcDir
	aptEtcDir = dir
	t.Cleanup(func() { aptEtcDir = old })

	got, found := installedAlgorandKeyrings([]aptSource{{File: "a.sources", InlineKey: true}})
	if len(got) != 0 || !found {
		t.Errorf("inline key source selected keyrings %q (found %t)", got, found)
	}
	if usesGlobalKeys([]aptSource{{File: "a.sources", InlineKey: true}}) {
		t.Error("inline key source reported as using global keys")
	}
	if !usesGlobalKeys([]aptSource{{File: "a.list"}}) {
		t.Error("source without signed-by not reported as using global keys")
	}
}

func TestFormatFingerprint(t *testing.T) {
	if got, want := formatFingerprint("611DE94A396F01359C72AF56EDACD29DA10A4EA6"), "611D E94A 396F 0135 9C72  AF56 EDAC D29D A10A 4EA6"; got != want {
		t.Errorf("formatFingerprint = %q, want %q", got, want)
	}
}

func TestAptUpdateIssuesLock(t *testing.T) {
	out := "Reading package lists...\nE: Could not get lock /var/lib/apt/lists/lock. It is held by process 1234 (apt-get)\nE: Unable to lock directory /var/lib/apt/lists/\n"
	failures, warnings, fatal := aptUpdateIssues(out, nil)
	if len(failures) != 0 || len(warnings) != 0 || len(fatal) != 2 {
		t.Errorf("lock errors sorted as failures %q, warnings %q, fatal %q", failures, warnings, fatal)
	}
}

func TestAptUpdateIssuesNonFatalErrors(t *testing.T) {
	out := `Hit:1 https://releases.algorand.com/deb stable InRelease
E: The repository 'file:/opt/localrepo ./ Release' does not have a Release file.
E: Problem executing scripts APT::Update::Post-Invoke-Success 'if /usr/bin/test -w /var/lib/command-not-found/ -a -e /usr/lib/cnf-update-db; then /usr/lib/cnf-update-db > /dev/null; fi'
E: Sub-process returned an error code
`
	failures, warnings, fatal := aptUpdateIssues(out, nil)
	if len(failures) != 0 || len(fatal) != 0 || len(warnings) != 3 {
		t.Errorf("non-fatal errors sorted as failures %q, warnings %q, fatal %q", failures, warnings, fatal)
	}
}

func TestPolicySources(t *testing.T) {
	installedOnly := `algorand:
  Installed: 5.0.2
  Candidate: 5.0.2
  Version table:
 *** 5.0.2 100
        100 /var/lib/dpkg/status
`
	proxied := `algorand:
  Installed: 5.0.2
  Candidate: 5.0.2
  Version table:
 *** 5.0.2 500
        500 http://apt-cache:3142/HTTPS///releases.algorand.com/deb stable/main amd64 Packages
        100 /var/lib/dpkg/status
     4.4.1 500
        500 http://apt-cache:3142/HTTPS///releases.algorand.com/deb stable/main amd64 Packages
`
	if got := policySources(installedOnly); len(got) != 0 {
		t.Errorf("installed copy reported as a repository: %q", got)
	}
	if got, want := policySources(proxied), []string{"http://apt-cache:3142/HTTPS///releases.algorand.com/deb"}; !reflect.DeepEqual(got, want) {
		t.Errorf("policySources() = %q, want %q", got, want)
	}
}

func TestStripComment(t *testing.T) {
	tests := map[string]string{
		"deb https://releases.algorand.com/deb/ stable main # note":               "deb https://releases.algorand.com/deb/ stable main ",
		"deb [signed-by=/k#1.gpg] https://releases.algorand.com/deb/ stable main": "deb [signed-by=/k#1.gpg] https://releases.algorand.com/deb/ stable main",
		"# deb https://releases.algorand.com/deb/ stable main":                    "",
	}
	for line, want := range tests {
		if got := stripComment(line); got != want {
			t.Errorf("stripComment(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestAptUpdateIssuesOtherRepository(t *testing.T) {
	out := `Err:3 https://example.com/apt stable InRelease
  Could not resolve 'example.com'
W: Failed to fetch https://example.com/apt/dists/stable/InRelease  Could not resolve 'example.com'
E: The repository 'https://example.com/apt stable InRelease' is not signed.
W: Some index files failed to download. They have been ignored, or old ones used instead.
`
	failures, warnings, fatal := aptUpdateIssues(out, nil)
	if len(failures) != 0 || len(fatal) != 0 || len(warnings) != 3 {
		t.Errorf("other repository errors sorted as failures %q, warnings %q, fatal %q", failures, warnings, fatal)
	}
}

func stubAptGetUpdate(t *testing.T, out string, err error) {
	t.Helper()
	old := runAptGetUpdate
	runAptGetUpdate = func() (string, error) { return out, err }
	t.Cleanup(func() { runAptGetUpdate = old })
}

func stubAptCachePolicy(t *testing.T, out string) {
	t.Helper()
	old := runAptCachePolicy
	runAptCachePolicy = func() (string, error) { return out, nil }
	t.Cleanup(func() { runAptCachePolicy = old })
}

const mirrorPolicy = `algorand:
  Installed: 4.4.1
  Candidate: 4.4.1
  Version table:
 *** 4.4.1 500
        500 https://mirror.example.com/algorand/deb stable/main amd64 Packages
        100 /var/lib/dpkg/status
`

const mirrorStaleKeyOutput = `Hit:1 http://archive.ubuntu.com/ubuntu noble InRelease
Err:2 https://mirror.example.com/algorand/deb stable InRelease
  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
Reading package lists...
W: GPG error: https://mirror.example.com/algorand/deb stable InRelease: The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
W: Failed to fetch https://mirror.example.com/algorand/deb/dists/stable/InRelease  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY F1C9063D88DB5791
W: Some index files failed to download. They have been ignored, or old ones used instead.
`

func TestAptUpdateIssuesMirror(t *testing.T) {
	repos := policySources(mirrorPolicy)
	failures, _, _ := aptUpdateIssues(mirrorStaleKeyOutput, repos)
	if len(failures) != 2 {
		t.Errorf("got %d mirror failures, want 2: %q", len(failures), failures)
	}

	// A repository whose URI only starts with the mirror's is someone else's.
	other := strings.ReplaceAll(mirrorStaleKeyOutput, "/algorand/deb", "/algorand/deb-testing")
	failures, warnings, fatal := aptUpdateIssues(other, repos)
	if len(failures) != 0 || len(fatal) != 0 || len(warnings) != 3 {
		t.Errorf("look-alike mirror sorted as failures %q, warnings %q, fatal %q", failures, warnings, fatal)
	}
}

func TestAptUpdateIssuesUnknownErrorsAreFatal(t *testing.T) {
	tests := map[string]string{
		"pre-invoke hook": "E: Problem executing scripts APT::Update::Pre-Invoke 'false'\nE: Sub-process returned an error code\n",
		"lists directory": "E: List directory /var/lib/apt/lists/partial is missing. - Acquire (13: Permission denied)\n",
		"unknown":         "E: Something apt has not said before\n",
	}
	for name, out := range tests {
		t.Run(name, func(t *testing.T) {
			failures, warnings, fatal := aptUpdateIssues(out, nil)
			want := strings.Count(out, "E:")
			if len(failures) != 0 || len(warnings) != 0 || len(fatal) != want {
				t.Errorf("sorted as failures %q, warnings %q, fatal %q; want %d fatal", failures, warnings, fatal, want)
			}
		})
	}
}

func TestAptUpdate(t *testing.T) {
	exit100 := errors.New("exit status 100")
	tests := []struct {
		name    string
		out     string
		err     error
		keys    keyOutcome
		wantErr string // empty means no error
	}{
		{"clean", "Hit:1 https://releases.algorand.com/deb stable InRelease\n", nil, keyCurrent, ""},
		{"stale key, declined", staleKeyUpdateOutput, nil, keyDeclined, "--accept-new-key"},
		{"stale key, unchecked", staleKeyUpdateOutput, nil, keyUnchecked, "could not check every keyring"},
		{"lock held", "E: Could not get lock /var/lib/apt/lists/lock. It is held by process 1234 (apt-get)\n", exit100, keyCurrent, "Could not get lock"},
		{"other repository broken", "E: The repository 'https://example.com/apt stable InRelease' is not signed.\n", exit100, keyCurrent, ""},
		{"failed without output", "sudo: a password is required\n", errors.New("exit status 1"), keyCurrent, "password is required"},
		{"pre-invoke hook failed", "E: Problem executing scripts APT::Update::Pre-Invoke 'false'\nE: Sub-process returned an error code\n", exit100, keyCurrent, "Pre-Invoke"},
		{"mirror stale key", mirrorStaleKeyOutput, nil, keyUnchecked, "mirror.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubAptGetUpdate(t, tt.out, tt.err)
			stubAptCachePolicy(t, mirrorPolicy)
			err := aptUpdate(tt.keys)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("aptUpdate() = %v, want no error", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("aptUpdate() = %v, want an error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestIsManagedKeyringPath(t *testing.T) {
	dir := t.TempDir()
	old := aptEtcDir
	aptEtcDir = dir
	t.Cleanup(func() { aptEtcDir = old })

	tests := map[string]bool{
		"/usr/share/keyrings/algorand-keyring.gpg":          true,
		filepath.Join(dir, "keyrings", "algorand.asc"):      true,
		filepath.Join(dir, "trusted.gpg.d", "algorand.gpg"): true,
		"/usr/share/keyrings/../../../etc/shadow":           false,
		"/usr/share/keyrings/sub/algorand.gpg":              false,
		"/etc/passwd":                                       false,
		"usr/share/keyrings/algorand.gpg":                   false,
	}
	for path, want := range tests {
		if got := isManagedKeyringPath(path); got != want {
			t.Errorf("isManagedKeyringPath(%q) = %t, want %t", path, got, want)
		}
	}
}

// writtenKeyring is one keyring write captured by the keyFixture stub.
type writtenKeyring struct {
	Path string
	Data []byte
}

// keyFixture sets up an apt directory whose Algorand source is signed by a
// keyring holding the given fixture key, and serves the published key from
// the 2026 fixture. It returns the keyring path and the paths written.
func keyFixture(t *testing.T, installedKey string) (string, *[]writtenKeyring) {
	t.Helper()
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not installed")
	}
	dir := t.TempDir()
	oldEtc, oldFetch, oldWrite, oldNow := aptEtcDir, fetchPublishedKey, writeKeyring, timeNow
	t.Cleanup(func() { aptEtcDir, fetchPublishedKey, writeKeyring, timeNow = oldEtc, oldFetch, oldWrite, oldNow })
	aptEtcDir = dir
	timeNow = func() time.Time { return afterRotation }

	keyring := filepath.Join(dir, "keyrings", "algorand.asc")
	if err := os.MkdirAll(filepath.Dir(keyring), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("testdata", installedKey))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyring, data, 0644); err != nil {
		t.Fatal(err)
	}
	published, err := os.ReadFile(filepath.Join("testdata", "algorand-2026.asc"))
	if err != nil {
		t.Fatal(err)
	}
	fetchPublishedKey = func() ([]byte, error) { return published, nil }
	var written []writtenKeyring
	writeKeyring = func(path string, armored []byte) error {
		written = append(written, writtenKeyring{path, armored})
		return nil
	}
	return keyring, &written
}

func writtenPaths(written []writtenKeyring) []string {
	var paths []string
	for _, w := range written {
		paths = append(paths, w.Path)
	}
	return paths
}

func fingerprints(certs []pgpCert) map[string]bool {
	fprs := map[string]bool{}
	for _, cert := range certs {
		for _, key := range cert.Keys {
			fprs[key.Fingerprint] = true
		}
	}
	return fprs
}

func TestRefreshAptKey(t *testing.T) {
	tests := []struct {
		name        string
		installed   string
		answer      bool
		wantAsked   bool
		wantOutcome keyOutcome
	}{
		{"expired key, accepted", "algorand-expired.asc", true, true, keyInstalled},
		{"expired key, declined", "algorand-expired.asc", false, true, keyDeclined},
		{"current key", "algorand-2026.asc", true, false, keyCurrent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring, written := keyFixture(t, tt.installed)
			asked := false
			outcome, err := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{keyring}}}, func(string) bool {
				asked = true
				return tt.answer
			})
			if err != nil {
				t.Fatal(err)
			}
			if asked != tt.wantAsked || outcome != tt.wantOutcome {
				t.Errorf("asked %t, outcome %v; want asked %t, outcome %v", asked, outcome, tt.wantAsked, tt.wantOutcome)
			}
			wantWritten := []string(nil)
			if tt.wantAsked && tt.answer {
				wantWritten = []string{keyring}
			}
			if got := writtenPaths(*written); !reflect.DeepEqual(got, wantWritten) {
				t.Errorf("wrote %q, want %q", got, wantWritten)
			}
		})
	}
}

func TestRefreshAptKeyRefusesUnmanagedPath(t *testing.T) {
	_, written := keyFixture(t, "algorand-expired.asc")
	outsider := filepath.Join(t.TempDir(), "algorand.asc")
	data, _ := os.ReadFile(filepath.Join("testdata", "algorand-expired.asc"))
	if err := os.WriteFile(outsider, data, 0644); err != nil {
		t.Fatal(err)
	}
	outcome, err := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{outsider}}}, func(string) bool {
		t.Error("asked to install into an unmanaged path")
		return true
	})
	if err != nil || outcome != keyUnchecked || len(*written) != 0 {
		t.Errorf("outcome %v, err %v, wrote %q; want keyUnchecked and nothing written", outcome, err, writtenPaths(*written))
	}
}

func TestRefreshAptKeyRefusesOtherPrimary(t *testing.T) {
	keyring, written := keyFixture(t, "algorand-expired.asc")
	// Algorand's RPM key: a real key, but not the apt repository primary.
	other, err := os.ReadFile(filepath.Join("testdata", "other-primary.asc"))
	if err != nil {
		t.Fatal(err)
	}
	fetchPublishedKey = func() ([]byte, error) { return other, nil }
	outcome, _ := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{keyring}}}, func(string) bool {
		t.Error("asked to install a key with another primary")
		return true
	})
	if outcome != keyUnchecked || len(*written) != 0 {
		t.Errorf("outcome %v, wrote %q; want keyUnchecked and nothing written", outcome, writtenPaths(*written))
	}
}

func TestIsAlgorandKey(t *testing.T) {
	if !isAlgorandKey(parseGPGColons(rotatedKeyColons)) {
		t.Error("Algorand key not recognised")
	}
	other := parseGPGColons(strings.ReplaceAll(rotatedKeyColons, algorandPrimaryFingerprint, "0000000000000000000000000000000000000000"))
	if isAlgorandKey(other) {
		t.Error("key with another primary accepted as the Algorand key")
	}
	if isAlgorandKey(append(parseGPGColons(rotatedKeyColons), other...)) {
		t.Error("key file with an extra primary accepted as the Algorand key")
	}
}

func TestRefreshAptKeyMergesKeys(t *testing.T) {
	keyring, written := keyFixture(t, "algorand-expired.asc")
	outcome, err := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{keyring}}}, func(string) bool { return true })
	if err != nil || outcome != keyInstalled || len(*written) != 1 {
		t.Fatalf("outcome %v, err %v, wrote %d keyrings; want one keyring installed", outcome, err, len(*written))
	}
	certs, err := readCerts((*written)[0].Data)
	if err != nil {
		t.Fatal(err)
	}
	fprs := fingerprints(certs)
	for _, want := range []string{algorandPrimaryFingerprint, "F321C257B7CC77F631EE89C3007CB592A5D131FE", "A2696A69562A3F158F5DD446F1C9063D88DB5791"} {
		if !fprs[want] {
			t.Errorf("merged keyring lacks %s", want)
		}
	}
}

func TestRefreshAptKeyPublishedOlderThanInstalled(t *testing.T) {
	keyring, written := keyFixture(t, "algorand-2026.asc")
	old, err := os.ReadFile(filepath.Join("testdata", "algorand-expired.asc"))
	if err != nil {
		t.Fatal(err)
	}
	fetchPublishedKey = func() ([]byte, error) { return old, nil }
	outcome, _ := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{keyring}}}, func(string) bool {
		t.Error("asked to install an older key")
		return true
	})
	if outcome != keyCurrent || len(*written) != 0 {
		t.Errorf("outcome %v, wrote %q; want keyCurrent and nothing written", outcome, writtenPaths(*written))
	}
}

func TestRefreshAptKeyRefusesForeignKeyring(t *testing.T) {
	tests := map[string][]string{
		"another vendor's key": {"other-primary.asc"},
		"several keys":         {"algorand-expired.asc", "other-primary.asc"},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			keyring, written := keyFixture(t, "algorand-expired.asc")
			var data []byte
			for _, file := range files {
				part, err := os.ReadFile(filepath.Join("testdata", file))
				if err != nil {
					t.Fatal(err)
				}
				data = append(data, part...)
			}
			if err := os.WriteFile(keyring, data, 0644); err != nil {
				t.Fatal(err)
			}
			outcome, _ := refreshAptKey([]aptSource{{File: "a.list", Keyrings: []string{keyring}}}, func(string) bool {
				t.Error("asked to replace a keyring that is not just the Algorand key")
				return true
			})
			if outcome != keyUnchecked || len(*written) != 0 {
				t.Errorf("outcome %v, wrote %q; want keyUnchecked and nothing written", outcome, writtenPaths(*written))
			}
		})
	}
}

func TestRefreshAptKeyTrustedGPGD(t *testing.T) {
	keyring, written := keyFixture(t, "algorand-expired.asc")
	trusted := filepath.Join(aptEtcDir, "trusted.gpg.d", "algorand.asc")
	if err := os.MkdirAll(filepath.Dir(trusted), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(keyring, trusted); err != nil {
		t.Fatal(err)
	}
	asked := false
	outcome, err := refreshAptKey([]aptSource{{File: "a.list"}}, func(string) bool {
		asked = true
		return true
	})
	if err != nil || !asked || outcome != keyInstalled {
		t.Errorf("asked %t, outcome %v, err %v; want asked and keyInstalled", asked, outcome, err)
	}
	if got := writtenPaths(*written); !reflect.DeepEqual(got, []string{trusted}) {
		t.Errorf("wrote %q, want %q", got, trusted)
	}
}

func TestRefreshAptKeyGlobalKeyNotFound(t *testing.T) {
	_, written := keyFixture(t, "algorand-expired.asc")
	outcome, _ := refreshAptKey([]aptSource{{File: "a.list"}}, func(string) bool {
		t.Error("asked with no keyring to update")
		return true
	})
	if outcome != keyUnchecked || len(*written) != 0 {
		t.Errorf("outcome %v, wrote %q; want keyUnchecked and nothing written", outcome, writtenPaths(*written))
	}
}
