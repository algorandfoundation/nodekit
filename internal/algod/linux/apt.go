package linux

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/charmbracelet/log"
)

// Algorand signs its apt repository with a subkey of the Algorand primary key.
// When that subkey expires or rotates, `apt-get update` keeps the cached
// package index, prints a warning and still exits 0, so the upgrade quietly
// installs nothing. nodekit therefore compares the installed key with the
// published one before updating, and reads apt's output for repository errors.

const algorandRepoHost = "releases.algorand.com"

const algorandAptKeyURL = "https://releases.algorand.com/key.pub"

// algorandPrimaryFingerprint is the Algorand primary key that signs every
// repository subkey. nodekit only installs published keys under this primary.
const algorandPrimaryFingerprint = "611DE94A396F01359C72AF56EDACD29DA10A4EA6"

// aptEtcDir is where apt keeps its sources and trusted keys. Tests replace it.
var aptEtcDir = "/etc/apt"

// Swapped out in tests.
var (
	fetchPublishedKey = func() ([]byte, error) { return fetchKey(algorandAptKeyURL) }
	writeKeyring      = installKeyring
	// LC_ALL=C keeps apt's messages in English so they can be recognised.
	// It is passed through sudo rather than via env, so sudoers rules that
	// only allow apt-get still match.
	runAptGetUpdate = func() (string, error) {
		return system.Run([]string{"sudo", "LC_ALL=C", "apt-get", "update"})
	}
	runAptCachePolicy = func() (string, error) {
		return system.Run([]string{"env", "LC_ALL=C", "apt-cache", "policy", "algorand"})
	}
	timeNow = time.Now
)

// aptSource is an enabled apt source that points at the Algorand repository.
type aptSource struct {
	File string
	// Keyrings lists the signed-by keyring files. Empty, without InlineKey,
	// means apt checks the repository against its global trusted keys.
	Keyrings []string
	// InlineKey is set when the key is written into the source itself.
	InlineKey bool
}

// findAlgorandAptSources returns the enabled apt sources for the Algorand repository.
func findAlgorandAptSources() []aptSource {
	var sources []aptSource
	files := []string{filepath.Join(aptEtcDir, "sources.list")}
	for _, pattern := range []string{"*.list", "*.sources"} {
		matches, _ := filepath.Glob(filepath.Join(aptEtcDir, "sources.list.d", pattern))
		files = append(files, matches...)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Warn(fmt.Sprintf("Could not read %s: %v", file, err))
			}
			continue
		}
		if strings.HasSuffix(file, ".sources") {
			sources = append(sources, parseDeb822Sources(file, string(content))...)
		} else {
			sources = append(sources, parseOneLineSources(file, string(content))...)
		}
	}
	return sources
}

// isAlgorandURI reports whether an apt source URI points at the Algorand repository.
func isAlgorandURI(uri string) bool {
	u, err := url.Parse(uri)
	return err == nil && strings.EqualFold(u.Hostname(), algorandRepoHost)
}

// parseOneLineSources reads the one-line format:
// deb [arch=amd64 signed-by=/path/key.gpg] https://releases.algorand.com/deb/ stable main
func parseOneLineSources(file, content string) []aptSource {
	var sources []aptSource
	for _, line := range strings.Split(content, "\n") {
		line = stripComment(line)
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "deb")
		if !ok || (!strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\t")) {
			continue
		}
		rest = strings.TrimSpace(rest)
		var options []string
		if strings.HasPrefix(rest, "[") {
			end := strings.Index(rest, "]")
			if end < 0 {
				continue
			}
			options = strings.Fields(rest[1:end])
			rest = rest[end+1:]
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 || !isAlgorandURI(fields[0]) {
			continue
		}
		source := aptSource{File: file}
		for _, option := range options {
			if value, ok := strings.CutPrefix(option, "signed-by="); ok {
				source.Keyrings = append(source.Keyrings, keyringPaths(value)...)
			}
		}
		sources = append(sources, source)
	}
	return sources
}

// stripComment drops a trailing # comment. A # inside [options] is part of
// the options, as apt reads it.
func stripComment(line string) string {
	depth := 0
	for i, r := range line {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '#':
			if depth == 0 {
				return line[:i]
			}
		}
	}
	return line
}

// parseDeb822Sources reads the deb822 format used by *.sources files.
// Stanzas are separated by lines that are empty or hold only whitespace.
func parseDeb822Sources(file, content string) []aptSource {
	var sources []aptSource
	var stanza []string
	flush := func() {
		if source, ok := deb822Source(file, stanza); ok {
			sources = append(sources, source)
		}
		stanza = nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		stanza = append(stanza, line)
	}
	flush()
	return sources
}

func deb822Source(file string, stanza []string) (aptSource, bool) {
	fields := map[string]string{}
	last := ""
	for _, line := range stanza {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && last != "" {
			fields[last] += "\n" + strings.TrimSpace(line)
			continue
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		last = strings.ToLower(strings.TrimSpace(name))
		fields[last] = strings.TrimSpace(value)
	}
	algorand := false
	for _, uri := range strings.Fields(fields["uris"]) {
		if isAlgorandURI(uri) {
			algorand = true
		}
	}
	if !algorand || !containsField(fields["types"], "deb") {
		return aptSource{}, false
	}
	if enabled := strings.ToLower(fields["enabled"]); enabled == "no" || enabled == "false" {
		return aptSource{}, false
	}
	source := aptSource{File: file}
	if signedBy := fields["signed-by"]; strings.Contains(signedBy, "BEGIN PGP") {
		source.InlineKey = true
	} else {
		source.Keyrings = keyringPaths(signedBy)
	}
	return source, true
}

func containsField(value, want string) bool {
	for _, field := range strings.Fields(value) {
		if field == want {
			return true
		}
	}
	return false
}

// keyringPaths returns the file paths in a signed-by value. Fingerprints are
// skipped: nodekit can only refresh keyring files.
func keyringPaths(value string) []string {
	var paths []string
	for _, entry := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if strings.HasPrefix(entry, "/") {
			paths = append(paths, entry)
		}
	}
	return paths
}

// pgpKey is one primary key or subkey, as listed by gpg --with-colons.
type pgpKey struct {
	Primary     bool
	Fingerprint string
	Algo        string
	Bits        int
	Caps        string
	Created     time.Time
	Expires     time.Time // zero when the key does not expire
	Revoked     bool
}

func (k pgpKey) expired(now time.Time) bool {
	return !k.Expires.IsZero() && !now.Before(k.Expires)
}

// pgpCert is a primary key with its subkeys (Keys[0] is the primary) and user IDs.
type pgpCert struct {
	Keys []pgpKey
	UIDs []string
}

var pgpAlgos = map[string]string{
	"1": "rsa", "2": "rsa", "3": "rsa", "16": "elg", "17": "dsa", "18": "ecdh", "19": "ecdsa", "22": "eddsa",
}

// parseGPGColons parses `gpg --with-colons --fixed-list-mode` key listings.
func parseGPGColons(out string) []pgpCert {
	var certs []pgpCert
	var current *pgpKey
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, ":")
		if len(f) < 10 {
			continue
		}
		switch f[0] {
		case "pub", "sub":
			if f[0] == "pub" {
				certs = append(certs, pgpCert{})
			}
			if len(certs) == 0 {
				continue
			}
			bits, _ := strconv.Atoi(f[2])
			key := pgpKey{
				Primary: f[0] == "pub",
				Algo:    pgpAlgos[f[3]],
				Bits:    bits,
				Created: unixField(f[5]),
				Expires: unixField(f[6]),
				Revoked: f[1] == "r",
			}
			if len(f) > 11 {
				// Lowercase letters are this key's own capabilities; uppercase
				// ones on the primary summarise the whole certificate.
				key.Caps = strings.Map(func(r rune) rune {
					if r >= 'a' && r <= 'z' {
						return r
					}
					return -1
				}, f[11])
			}
			cert := &certs[len(certs)-1]
			cert.Keys = append(cert.Keys, key)
			current = &cert.Keys[len(cert.Keys)-1]
		case "fpr":
			if current != nil && current.Fingerprint == "" {
				current.Fingerprint = f[9]
			}
		case "uid":
			if len(certs) > 0 {
				certs[len(certs)-1].UIDs = append(certs[len(certs)-1].UIDs, f[9])
			}
		}
	}
	return certs
}

func unixField(value string) time.Time {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds == 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
}

// readCerts lists the keys in an armored or binary key file using a throwaway
// gpg home, so the user's own keyring is never touched.
func readCerts(data []byte) ([]pgpCert, error) {
	home, err := os.MkdirTemp("", "nodekit-gpg-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(home)
	keyFile := filepath.Join(home, "key")
	if err := os.WriteFile(keyFile, data, 0600); err != nil {
		return nil, err
	}
	out, err := system.RunOutput([]string{
		"gpg", "--batch", "--no-tty", "--homedir", home,
		"--with-colons", "--fixed-list-mode",
		"--import-options", "show-only", "--import", keyFile,
	})
	if err != nil {
		return nil, fmt.Errorf("could not read key with gpg: %w", err)
	}
	certs := parseGPGColons(out)
	if len(certs) == 0 {
		return nil, errors.New("no keys found")
	}
	return certs, nil
}

// needsKeyUpdate reports whether remote differs from local in a way apt
// should know about: a usable key that local lacks, a revocation local does
// not have yet, or a changed expiry date (extended or brought forward).
func needsKeyUpdate(local, remote []pgpCert, now time.Time) bool {
	installed := map[string]pgpKey{}
	for _, cert := range local {
		for _, key := range cert.Keys {
			installed[key.Fingerprint] = key
		}
	}
	for _, cert := range remote {
		for _, key := range cert.Keys {
			old, ok := installed[key.Fingerprint]
			switch {
			case key.Revoked:
				if ok && !old.Revoked {
					return true
				}
			case !ok:
				if !key.expired(now) {
					return true
				}
			case old.Revoked:
				// Never undo a revocation the installed key already has.
			case !key.Expires.Equal(old.Expires):
				return true
			}
		}
	}
	return false
}

func primaryFingerprints(certs []pgpCert) map[string]bool {
	fprs := map[string]bool{}
	for _, cert := range certs {
		if len(cert.Keys) > 0 {
			fprs[cert.Keys[0].Fingerprint] = true
		}
	}
	return fprs
}

// isAlgorandKey reports whether certs is exactly the Algorand primary key.
func isAlgorandKey(certs []pgpCert) bool {
	fprs := primaryFingerprints(certs)
	return len(certs) == 1 && fprs[algorandPrimaryFingerprint]
}

// formatFingerprint groups a fingerprint the way gpg prints it.
func formatFingerprint(fpr string) string {
	var groups []string
	for i := 0; i < len(fpr); i += 4 {
		end := min(i+4, len(fpr))
		groups = append(groups, fpr[i:end])
	}
	out := strings.Join(groups, " ")
	if len(groups) == 10 {
		out = strings.Join(groups[:5], " ") + "  " + strings.Join(groups[5:], " ")
	}
	return out
}

var capNames = map[rune]string{'s': "signing", 'c': "certify", 'e': "encryption", 'a': "authentication"}

// describeCerts renders keys for the user to compare before trusting them.
func describeCerts(certs []pgpCert, now time.Time) string {
	var b strings.Builder
	for _, cert := range certs {
		for _, key := range cert.Keys {
			kind := "sub"
			if key.Primary {
				kind = "pub"
			}
			fmt.Fprintf(&b, "  %s  %s%d  %s\n", kind, key.Algo, key.Bits, formatFingerprint(key.Fingerprint))
			if key.Primary {
				for _, uid := range cert.UIDs {
					fmt.Fprintf(&b, "       %s\n", uid)
				}
			}
			var caps []string
			for _, c := range key.Caps {
				if name, ok := capNames[c]; ok {
					caps = append(caps, name)
				}
			}
			status := "does not expire"
			switch {
			case key.Revoked:
				status = "REVOKED"
			case key.expired(now):
				status = "EXPIRED " + key.Expires.Format(time.DateOnly)
			case !key.Expires.IsZero():
				status = "expires " + key.Expires.Format(time.DateOnly)
			}
			fmt.Fprintf(&b, "       %s, created %s, %s\n", strings.Join(caps, "/"), key.Created.Format(time.DateOnly), status)
		}
	}
	return b.String()
}

// managedKeyringDirs are the directories nodekit will write keyrings into.
func managedKeyringDirs() []string {
	return []string{"/usr/share/keyrings", filepath.Join(aptEtcDir, "keyrings"), filepath.Join(aptEtcDir, "trusted.gpg.d")}
}

// isManagedKeyringPath reports whether path is a file directly inside one of
// the keyring directories, so a source cannot point nodekit at other files.
func isManagedKeyringPath(path string) bool {
	if !filepath.IsAbs(path) {
		return false
	}
	dir := filepath.Dir(filepath.Clean(path))
	for _, allowed := range managedKeyringDirs() {
		if dir == allowed {
			return true
		}
	}
	return false
}

// installedAlgorandKeyrings returns the keyring files that hold the Algorand
// repository key: the signed-by files, plus the files in trusted.gpg.d that
// contain the Algorand primary key when a source has no signed-by. found
// is false when a source has no signed-by and no such file exists.
func installedAlgorandKeyrings(sources []aptSource) (paths []string, found bool) {
	seen := map[string]bool{}
	add := func(path string) {
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	for _, source := range sources {
		for _, path := range source.Keyrings {
			add(path)
		}
	}
	if !usesGlobalKeys(sources) {
		return paths, true
	}
	for _, pattern := range []string{"*.gpg", "*.asc"} {
		matches, _ := filepath.Glob(filepath.Join(aptEtcDir, "trusted.gpg.d", pattern))
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			certs, err := readCerts(data)
			if err != nil {
				continue
			}
			if primaryFingerprints(certs)[algorandPrimaryFingerprint] {
				add(path)
				found = true
			}
		}
	}
	return paths, found
}

// keyFetchTimeout bounds the key download so an offline machine does not hang the upgrade.
const keyFetchTimeout = 30 * time.Second

func fetchKey(keyURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, keyURL, nil)
	if err != nil {
		return nil, err
	}
	api.SetUserAgent(req)
	client := &http.Client{
		Timeout: keyFetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" {
				return fmt.Errorf("refusing to follow a redirect to %s", req.URL)
			}
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", keyURL, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(data), "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
		return nil, fmt.Errorf("%s did not return an armored public key", keyURL)
	}
	return data, nil
}

// keyOutcome records what the key check did, so an apt failure afterwards can
// tell the user what to do next.
type keyOutcome int

const (
	keyUnchecked keyOutcome = iota // some keyring apt uses could not be checked
	keyCurrent                     // every keyring apt uses already has the key
	keyDeclined                    // the user kept the installed key
	keyInstalled                   // the new key was installed
)

// staleKeyring is a keyring file that lacks part of the published key.
type staleKeyring struct {
	Path      string
	Installed []pgpCert
	Updated   []pgpCert // the installed key merged with the published one
	Data      []byte    // Updated, armored, as it will be written
}

// refreshAptKey installs Algorand's published repository key, after asking,
// when it differs from the installed one. Problems reading or fetching keys
// are logged rather than returned: apt-get update reports anything that still
// breaks, and the outcome lets that error explain why.
func refreshAptKey(sources []aptSource, confirm func(string) bool) (keyOutcome, error) {
	if !system.CmdExists("gpg") {
		log.Warn("gpg is not installed, so the Algorand signing key cannot be checked")
		return keyUnchecked, nil
	}
	armored, err := fetchPublishedKey()
	if err != nil {
		log.Warn(fmt.Sprintf("Could not fetch the Algorand signing key: %v", err))
		return keyUnchecked, nil
	}
	remote, err := readCerts(armored)
	if err != nil {
		log.Warn(fmt.Sprintf("Could not read the published Algorand signing key: %v", err))
		return keyUnchecked, nil
	}
	if !isAlgorandKey(remote) {
		log.Warn(fmt.Sprintf("The key at %s is not the Algorand primary key nodekit expects (%s), so nodekit will not install it", algorandAptKeyURL, formatFingerprint(algorandPrimaryFingerprint)))
		return keyUnchecked, nil
	}

	now := timeNow()
	unchecked := false
	for _, source := range sources {
		if source.InlineKey {
			log.Warn(fmt.Sprintf("%s has the Algorand key written into it, which nodekit does not change. If apt reports a signature error, replace its Signed-By key with the one at %s", source.File, algorandAptKeyURL))
			unchecked = true
		}
	}
	paths, found := installedAlgorandKeyrings(sources)
	if !found {
		// The key is in the legacy keyring or nowhere apt can use it.
		warnLegacyKeyring(remote, now)
		unchecked = true
	}

	var stale []staleKeyring
	for _, path := range paths {
		if !isManagedKeyringPath(path) {
			log.Warn(fmt.Sprintf("nodekit only updates keyrings in %s, so it did not check %s", strings.Join(managedKeyringDirs(), ", "), path))
			unchecked = true
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			log.Warn(fmt.Sprintf("Could not read %s: %v", path, err))
			unchecked = true
			continue
		}
		local, err := readCerts(data)
		if err != nil {
			log.Warn(fmt.Sprintf("Could not read %s: %v", path, err))
			unchecked = true
			continue
		}
		if !isAlgorandKey(local) {
			// It may belong to another vendor too, so it is never replaced, even with --accept-new-key.
			log.Warn(fmt.Sprintf("%s does not hold just the Algorand key, so nodekit will not change it. Update the Algorand key in it by hand from %s", path, algorandAptKeyURL))
			unchecked = true
			continue
		}
		// Merge rather than replace, so keys the installed file has and the
		// published one lacks are kept.
		merged, err := mergeKeys(data, armored)
		if err != nil {
			log.Warn(fmt.Sprintf("Could not merge the published key into %s: %v", path, err))
			unchecked = true
			continue
		}
		updated, err := readCerts(merged)
		if err != nil || !isAlgorandKey(updated) {
			log.Warn(fmt.Sprintf("Could not read the merged key for %s: %v", path, err))
			unchecked = true
			continue
		}
		if !needsKeyUpdate(local, updated, now) {
			continue
		}
		stale = append(stale, staleKeyring{Path: path, Installed: local, Updated: updated, Data: merged})
	}
	if len(stale) == 0 {
		if unchecked {
			return keyUnchecked, nil
		}
		log.Debug("Algorand signing key is up to date")
		return keyCurrent, nil
	}

	// The prompt goes to stderr with the logs, so it shows even when stdout is redirected.
	out := os.Stderr
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Algorand has updated its package signing key (%s).\n", algorandAptKeyURL)
	for _, keyring := range stale {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Installed key (%s):\n", keyring.Path)
		fmt.Fprint(out, describeCerts(keyring.Installed, now))
		fmt.Fprintln(out)
		fmt.Fprintln(out, "After the update:")
		fmt.Fprint(out, describeCerts(keyring.Updated, now))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "The primary key is unchanged.")
	fmt.Fprintln(out)

	if !confirm("Install the new Algorand signing key?") {
		log.Warn("Keeping the installed Algorand signing key")
		return keyDeclined, nil
	}
	for _, keyring := range stale {
		if err := writeKeyring(keyring.Path, keyring.Data); err != nil {
			return keyUnchecked, err
		}
		log.Info(fmt.Sprintf("Installed the new Algorand signing key to %s", keyring.Path))
	}
	if unchecked {
		return keyUnchecked, nil
	}
	return keyInstalled, nil
}

// mergeKeys combines an installed key file with the published key in a
// throwaway gpg home and returns the result armored. gpg keeps every subkey,
// signature and revocation from both.
func mergeKeys(installed, published []byte) ([]byte, error) {
	home, err := os.MkdirTemp("", "nodekit-gpg-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(home)
	files := []string{filepath.Join(home, "installed"), filepath.Join(home, "published")}
	for i, data := range [][]byte{installed, published} {
		if err := os.WriteFile(files[i], data, 0600); err != nil {
			return nil, err
		}
	}
	gpg := []string{"gpg", "--batch", "--no-tty", "--homedir", home}
	if _, err := system.RunOutput(append(append(gpg, "--import"), files...)); err != nil {
		return nil, fmt.Errorf("could not import keys with gpg: %w", err)
	}
	out, err := system.RunOutput(append(gpg, "--armor", "--export", algorandPrimaryFingerprint))
	if err != nil {
		return nil, fmt.Errorf("could not export the merged key with gpg: %w", err)
	}
	return []byte(out), nil
}

// usesGlobalKeys reports whether any source relies on apt's global trusted keys.
func usesGlobalKeys(sources []aptSource) bool {
	for _, source := range sources {
		if len(source.Keyrings) == 0 && !source.InlineKey {
			return true
		}
	}
	return false
}

// warnLegacyKeyring explains how to update an out-of-date Algorand key in the
// legacy trusted.gpg keyring, which holds other vendors' keys too, so nodekit
// leaves it alone.
func warnLegacyKeyring(remote []pgpCert, now time.Time) {
	path := filepath.Join(aptEtcDir, "trusted.gpg")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	certs, err := readCerts(data)
	if err != nil {
		return
	}
	for _, cert := range certs {
		if len(cert.Keys) > 0 && cert.Keys[0].Fingerprint == algorandPrimaryFingerprint && needsKeyUpdate([]pgpCert{cert}, remote, now) {
			log.Warn(fmt.Sprintf("The Algorand signing key in %s is out of date. nodekit does not change that shared keyring; update it with:\n  curl -fsSL %s | sudo gpg --no-default-keyring --keyring %s --import", path, algorandAptKeyURL, path))
			return
		}
	}
}

// installKeyring writes the key to path, armored for .asc files and binary
// otherwise, which is what apt expects from each extension.
func installKeyring(path string, armored []byte) error {
	dir, err := os.MkdirTemp("", "nodekit-key-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	src := filepath.Join(dir, "key.asc")
	if err := os.WriteFile(src, armored, 0644); err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Ext(path), ".asc") {
		binary := filepath.Join(dir, "key.gpg")
		if err := system.RunAll(system.CmdsList{{"gpg", "--batch", "--homedir", dir, "--dearmor", "-o", binary, src}}); err != nil {
			return err
		}
		src = binary
	}
	return system.RunAll(system.CmdsList{{"sudo", "install", "-m", "0644", src, path}})
}

// repoFailures are apt messages meaning a repository was not refreshed.
// Other warnings about the repository, such as a key kept in the legacy
// keyring, leave the package lists usable.
var repoFailures = []string{
	"Failed to fetch", "GPG error", "is not signed", "previous index files will be used",
	"signatures", "signature verification", "is not bound",
	"NO_PUBKEY", "EXPKEYSIG", "KEYEXPIRED", "REVKEYSIG", "BADSIG",
}

// aptFatalErrors are apt errors after which no repository was refreshed,
// even when they name a repository.
var aptFatalErrors = []string{
	"Could not get lock", "Unable to lock", "Could not open lock file", "Unable to acquire the dpkg frontend lock",
	"The list of sources could not be read", "Malformed entry", "Malformed line", "is not known on line",
	"Conflicting values set for option", "Unable to parse package file",
	"APT::Update::Pre-Invoke",
}

var signatureProblems = []string{
	"GPG error", "is not signed", "signatures", "signature verification", "is not bound",
	"NO_PUBKEY", "EXPKEYSIG", "KEYEXPIRED", "REVKEYSIG", "BADSIG",
}

// repoURIPattern finds a repository URI in an apt message: http(s)://,
// file:/, cdrom:[, mirror+https:// and so on.
var repoURIPattern = regexp.MustCompile(`[a-z][a-z0-9+.-]*:(/|\[)`)

func containsAny(line string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

// mentionsRepo reports whether line names one of the repositories: the
// Algorand host, or a mirror or proxy URI that apt says provides the package.
func mentionsRepo(line string, repos []string) bool {
	if strings.Contains(line, algorandRepoHost) {
		return true
	}
	for _, repo := range repos {
		for rest := line; ; {
			i := strings.Index(rest, repo)
			if i < 0 {
				break
			}
			rest = rest[i+len(repo):]
			// Stop https://host/deb from matching https://host/deb-testing.
			if rest == "" || strings.ContainsAny(rest[:1], "/ '\"") {
				return true
			}
		}
	}
	return false
}

// harmlessAptError reports whether an E: line from apt-get update leaves the
// Algorand lists usable. Anything not known to be harmless is fatal, since
// apt may have refreshed nothing and an old list would make the upgrade look
// up to date.
func harmlessAptError(line string, afterPostInvoke bool) bool {
	switch {
	case strings.Contains(line, "APT::Update::Post-Invoke"):
		// The lists were refreshed before the hook ran.
		return true
	case strings.Contains(line, "Sub-process returned an error code"):
		return afterPostInvoke
	case strings.Contains(line, "Some index files failed to download"):
		// A summary; failures of the Algorand repository are reported on
		// their own lines.
		return true
	default:
		// A problem with a repository other than Algorand's.
		return repoURIPattern.MatchString(line)
	}
}

// aptUpdateIssues sorts apt-get update output into:
//   - failures: the Algorand repository was not refreshed
//   - warnings: everything else, which does not stop the upgrade
//   - fatal: errors after which the lists may not have been refreshed, such as a held lock
//
// repos are the mirror or proxy URIs apt lists for the algorand package, on
// top of the Algorand host itself.
func aptUpdateIssues(out string, repos []string) (failures, warnings, fatal []string) {
	var fetchErrors []string
	inFetchError, afterPostInvoke := false, false
	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		if inFetchError && line != "" && (strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t")) {
			fetchErrors = append(fetchErrors, line)
			continue
		}
		inFetchError = false
		if strings.HasPrefix(line, "Err:") && mentionsRepo(line, repos) {
			fetchErrors = append(fetchErrors, line)
			inFetchError = true
			continue
		}
		isError := strings.HasPrefix(line, "E:")
		if !isError && !strings.HasPrefix(line, "W:") {
			continue
		}
		switch {
		case isError && containsAny(line, aptFatalErrors):
			fatal = append(fatal, line)
		case mentionsRepo(line, repos) && (isError || containsAny(line, repoFailures)):
			failures = append(failures, line)
		case !isError || harmlessAptError(line, afterPostInvoke):
			warnings = append(warnings, line)
		default:
			fatal = append(fatal, line)
		}
		if isError {
			afterPostInvoke = strings.Contains(line, "APT::Update::Post-Invoke")
		}
	}
	// apt normally repeats a failed fetch as a W: line; fall back to the
	// Err: lines in case it does not.
	if len(failures) == 0 {
		failures = fetchErrors
	}
	return failures, warnings, fatal
}

// aptUpdate runs apt-get update and fails when the Algorand repository could
// not be refreshed, which apt itself reports only as a warning.
func aptUpdate(keys keyOutcome) error {
	// Ask before updating: a failed update can leave apt without the lists
	// that name the repository.
	repos := algorandRepoURIs()
	log.Info("Refreshing package lists")
	out, err := runAptGetUpdate()
	failures, warnings, fatal := aptUpdateIssues(out, repos)
	if len(failures) > 0 {
		msg := "apt could not refresh the Algorand repository, so no new version can be installed:\n  " + strings.Join(failures, "\n  ")
		if containsAny(strings.Join(failures, "\n"), signatureProblems) {
			switch keys {
			case keyDeclined:
				msg += "\nRun `nodekit upgrade` again and accept the new Algorand signing key, or pass --accept-new-key when there is no terminal to answer on."
			case keyUnchecked:
				msg += "\nnodekit could not check every keyring apt uses for this repository (see the warnings above). Install the key from " + algorandAptKeyURL + " into the keyring named by the Algorand apt source."
			default:
				msg += "\nThe Algorand keyrings already have the published key; the repository signature itself may be at fault."
			}
		}
		return errors.New(msg)
	}
	if len(fatal) > 0 {
		return fmt.Errorf("apt-get update failed, so the package lists may not have been refreshed:\n  %s", strings.Join(fatal, "\n  "))
	}
	if err != nil && len(warnings) == 0 {
		return fmt.Errorf("apt-get update failed: %w\n%s", err, strings.TrimSpace(out))
	}
	for _, line := range warnings {
		log.Warn(line)
	}
	if err != nil {
		log.Warn(fmt.Sprintf("apt-get update failed (%v) because of the problems above, which do not affect the Algorand repository; continuing with the upgrade", err))
	}
	return nil
}

// algorandRepoURIs returns the repositories apt lists for the algorand
// package, which include mirrors and caching proxies the sources parser does
// not recognise. It returns nil when apt cannot be asked.
func algorandRepoURIs() []string {
	out, err := runAptCachePolicy()
	if err != nil {
		return nil
	}
	return policySources(out)
}

// aptHasAlgorandSource asks apt whether any repository offers the algorand
// package. Unlike the sources parser, this also sees mirrors and caching
// proxies. It answers true when apt cannot be asked, so it never blocks an
// upgrade on its own.
func aptHasAlgorandSource() bool {
	out, err := runAptCachePolicy()
	if err != nil {
		return true
	}
	return len(policySources(out)) > 0
}

// policySources returns the repository URIs `apt-cache policy` lists for the
// package, leaving out the installed copy in /var/lib/dpkg/status.
func policySources(out string) []string {
	var uris []string
	seen := map[string]bool{}
	inTable := false
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "Version table:" {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		// Source lines are "<priority> <source> ..."; version lines are
		// "[***] <version> <priority>".
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if _, err := strconv.Atoi(fields[0]); err != nil {
			continue
		}
		if _, err := strconv.Atoi(fields[1]); err == nil {
			continue
		}
		uri := strings.TrimRight(fields[1], "/")
		if fields[1] == "/var/lib/dpkg/status" || uri == "" || seen[uri] {
			continue
		}
		seen[uri] = true
		uris = append(uris, uri)
	}
	return uris
}

// InstalledVersion returns the version of the installed algorand package.
func InstalledVersion() (string, error) {
	if system.CmdExists("dpkg-query") {
		if out, err := system.RunOutput([]string{"dpkg-query", "-W", "-f=${Version}", "algorand"}); err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), nil
		}
	}
	if system.CmdExists("rpm") {
		if out, err := system.RunOutput([]string{"rpm", "-q", "--qf", "%{VERSION}-%{RELEASE}", "algorand"}); err == nil {
			return strings.TrimSpace(out), nil
		}
	}
	return "", errors.New("algorand package version not found")
}
