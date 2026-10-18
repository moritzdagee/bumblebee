// Tests for the fork's own daily scan script automation/bumblebee-scan.
//
// The script runs with a fake scanner (a small shell script that prints
// canned NDJSON), a fake WhatsApp sender that only records its arguments,
// and temporary folders for HOME, results and the catalog clone. Nothing
// here scans the real home folder, sends a message or touches the network:
// the catalog "origin" is a local git repository reached via file://.
package automation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const summary = `{"record_type":"scan_summary","files_considered":123,"duration_ms":45}`
const finding = `{"record_type":"finding","package":"evil"}`

type env struct {
	t       *testing.T
	home    string
	results string
	catalog string
	origin  string
	waLog   string
	bbLog   string
	vars    []string
}

// fakeScanner prints FAKE_BB_OUT on a normal scan. A catalog probe (any
// --root other than $HOME) fails when FAKE_PROBE_FAIL=1.
const fakeScanner = `#!/bin/sh
printf '%s\n' "$*" >> "$FAKE_BB_LOG"
root=""
prev=""
for a in "$@"; do
  [ "$prev" = "--root" ] && root="$a"
  prev="$a"
done
if [ "$root" != "$HOME" ]; then
  if [ "${FAKE_PROBE_FAIL:-0}" = "1" ]; then
    echo "unsupported catalog schema 9.9" >&2
    exit 2
  fi
  exit 0
fi
printf '%b' "$FAKE_BB_OUT"
echo "diag line" >&2
exit "${FAKE_BB_RC:-0}"
`

// fakeWA stands in for assistant_dataflow/scripts/wa (run via /usr/bin/python3).
const fakeWA = `import os, sys
with open(os.environ["FAKE_WA_LOG"], "a") as f:
    f.write("\t".join(sys.argv[1:]) + "\n")
sys.exit(int(os.environ.get("FAKE_WA_RC", "0")))
`

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// Isolate from the user's global git config and hooks (the global
	// pre-commit hook refuses commits on main).
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func newEnv(t *testing.T) *env {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("bumblebee-scan is a macOS script (zsh, /usr/bin/python3)")
	}
	for _, p := range []string{"/bin/zsh", "/usr/bin/python3"} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("%s not available", p)
		}
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	e := &env{
		t:       t,
		home:    filepath.Join(base, "home"),
		results: filepath.Join(base, "results"),
		catalog: filepath.Join(base, "catalog"),
		origin:  filepath.Join(base, "origin"),
		waLog:   filepath.Join(base, "wa.log"),
		bbLog:   filepath.Join(base, "bb.log"),
	}
	bb := filepath.Join(base, ".local/bin", "bumblebee")
	wa := filepath.Join(base, ".local/bin", "wa")
	write(t, bb, fakeScanner, 0o755)
	write(t, wa, fakeWA, 0o644)
	if err := os.MkdirAll(e.home, 0o755); err != nil {
		t.Fatal(err)
	}
	e.vars = []string{
		"HOME=" + e.home,
		"PATH=" + os.Getenv("PATH"),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"BUMBLEBEE_BIN=" + bb,
		"BUMBLEBEE_WA=" + wa,
		"BUMBLEBEE_RESULTS=" + e.results,
		"BUMBLEBEE_CATALOGS=" + e.catalog,
		"FAKE_WA_LOG=" + e.waLog,
		"FAKE_BB_LOG=" + e.bbLog,
	}
	return e
}

// withCatalog sets up a local "upstream" with one catalog commit and a
// shallow clone of it as the private catalog copy.
func (e *env) withCatalog() {
	t := e.t
	write(t, filepath.Join(e.origin, "threat_intel", "a.json"), `{"v":1}`, 0o644)
	gitCmd(t, e.origin, "init", "-q", "-b", "main")
	gitCmd(t, e.origin, "add", ".")
	gitCmd(t, e.origin, "commit", "-q", "-m", "c1")
	cmd := exec.Command("git", "clone", "-q", "--depth", "1", "file://"+e.origin, e.catalog)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
}

func (e *env) newUpstreamCommit() string {
	write(e.t, filepath.Join(e.origin, "threat_intel", "b.json"), `{"v":2}`, 0o644)
	gitCmd(e.t, e.origin, "add", ".")
	gitCmd(e.t, e.origin, "commit", "-q", "-m", "c2")
	return gitCmd(e.t, e.origin, "rev-parse", "HEAD")
}

func (e *env) run(extra ...string) (int, string) {
	e.t.Helper()
	cmd := exec.Command("/bin/zsh", "bumblebee-scan")
	cmd.Env = append(append([]string{}, e.vars...), extra...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode(), string(out)
		}
		e.t.Fatalf("run: %v", err)
	}
	return 0, string(out)
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

func TestCleanScanLogsSummaryAndStaysSilent(t *testing.T) {
	e := newEnv(t)
	code, out := e.run("FAKE_BB_OUT=" + summary + `\n`)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	log := read(t, filepath.Join(e.results, "last-run.log"))
	for _, want := range []string{"findings=0", "files=123", "duration_ms=45", "catalog=fetch-failed(offline?)"} {
		if !strings.Contains(log, want) {
			t.Errorf("last-run.log missing %q:\n%s", want, log)
		}
	}
	if wa := read(t, e.waLog); wa != "" {
		t.Errorf("no message expected on a clean run, got %q", wa)
	}
}

func TestScanArgumentsRootAndDenyList(t *testing.T) {
	e := newEnv(t)
	e.run("FAKE_BB_OUT=" + summary + `\n`)
	args := read(t, e.bbLog)
	for _, want := range []string{
		"scan --profile deep --root " + e.home,
		"--findings-only",
		"--exclude Photos Library.photoslibrary",
		"--exclude Caches",
		"--exclude .Trash",
		"--exclude Group Containers",
		"--max-duration 15m",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("scanner args missing %q:\n%s", want, args)
		}
	}
}

func TestFindingsAreCountedAndReported(t *testing.T) {
	e := newEnv(t)
	code, out := e.run("FAKE_BB_OUT=" + finding + `\n` + finding + `\n` + summary + `\n`)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if log := read(t, filepath.Join(e.results, "last-run.log")); !strings.Contains(log, "findings=2") {
		t.Errorf("expected findings=2 in log:\n%s", log)
	}
	wa := read(t, e.waLog)
	if !strings.Contains(wa, "send\tMoritz und Ted\t") || !strings.Contains(wa, "2 Treffer") {
		t.Errorf("expected a report with 2 Treffer, got %q", wa)
	}
}

func TestAbortedScanIsAFailureNotACleanWeek(t *testing.T) {
	e := newEnv(t)
	// Scanner exits 3 and writes no scan_summary.
	code, _ := e.run("FAKE_BB_OUT="+finding+`\n`, "FAKE_BB_RC=3")
	if code != 1 {
		t.Fatalf("expected exit 1 for an unfinished scan, got %d", code)
	}
	log := read(t, filepath.Join(e.results, "last-run.log"))
	if !strings.Contains(log, "SCAN-FEHLGESCHLAGEN exit=3") || !strings.Contains(log, "diag line") {
		t.Errorf("failure not logged with exit code and diagnostics:\n%s", log)
	}
	if strings.Contains(log, "findings=") {
		t.Errorf("an aborted scan must not be logged as a findings count:\n%s", log)
	}
	if wa := read(t, e.waLog); !strings.Contains(wa, "abgebrochen") {
		t.Errorf("expected an abort report, got %q", wa)
	}
}

func TestFailedDeliveryIsRecordedNotRerouted(t *testing.T) {
	e := newEnv(t)
	e.run("FAKE_BB_OUT="+finding+`\n`+summary+`\n`, "FAKE_WA_RC=1")
	log := read(t, filepath.Join(e.results, "last-run.log"))
	if !strings.Contains(log, "WHATSAPP-ZUSTELLUNG-FEHLGESCHLAGEN") || !strings.Contains(log, "1 Treffer") {
		t.Errorf("failed delivery not recorded:\n%s", log)
	}
}

func TestOnlyThirtyNewestResultFilesAreKept(t *testing.T) {
	e := newEnv(t)
	for i := 1; i <= 35; i++ {
		name := fmt.Sprintf("20250101-%06d", i)
		write(t, filepath.Join(e.results, "findings-"+name+".ndjson"), "", 0o644)
		write(t, filepath.Join(e.results, "diag-"+name+".ndjson"), "", 0o644)
	}
	write(t, filepath.Join(e.results, "keep-me.txt"), "", 0o644)
	e.run("FAKE_BB_OUT=" + summary + `\n`)

	for _, kind := range []string{"findings-", "diag-"} {
		matches, _ := filepath.Glob(filepath.Join(e.results, kind+"*.ndjson"))
		sort.Strings(matches)
		if len(matches) != 30 {
			t.Fatalf("%s: expected 30 files, got %d", kind, len(matches))
		}
		// 35 old + 1 from this run = 36; the 6 oldest by name are gone.
		if !strings.HasSuffix(matches[0], kind+"20250101-000007.ndjson") {
			t.Errorf("%s: oldest kept file is %s", kind, filepath.Base(matches[0]))
		}
	}
	if _, err := os.Stat(filepath.Join(e.results, "keep-me.txt")); err != nil {
		t.Errorf("unrelated file was removed: %v", err)
	}
}

func TestCatalogUpdatedOnlyWhenScannerCanReadIt(t *testing.T) {
	e := newEnv(t)
	e.withCatalog()
	newRev := e.newUpstreamCommit()
	e.run("FAKE_BB_OUT=" + summary + `\n`)
	if got := gitCmd(t, e.catalog, "rev-parse", "HEAD"); got != newRev {
		t.Fatalf("catalog not switched to new upstream commit: %s != %s", got, newRev)
	}
	if log := read(t, filepath.Join(e.results, "last-run.log")); !strings.Contains(log, "catalog=updated:"+newRev[:7]) {
		t.Errorf("update not logged:\n%s", log)
	}
}

func TestCatalogKeptWhenScannerRejectsIt(t *testing.T) {
	e := newEnv(t)
	e.withCatalog()
	oldRev := gitCmd(t, e.catalog, "rev-parse", "HEAD")
	e.newUpstreamCommit()
	e.run("FAKE_BB_OUT="+summary+`\n`, "FAKE_PROBE_FAIL=1")
	if got := gitCmd(t, e.catalog, "rev-parse", "HEAD"); got != oldRev {
		t.Fatalf("catalog switched although the scanner rejected it")
	}
	log := read(t, filepath.Join(e.results, "last-run.log"))
	if !strings.Contains(log, "catalog=NOT-UPDATED(scanner-too-old:unsupported catalog schema 9.9") {
		t.Errorf("rejection not logged:\n%s", log)
	}
	// The daily scan still runs on the old catalog.
	if !strings.Contains(log, "findings=0") {
		t.Errorf("scan did not run after the rejected update:\n%s", log)
	}
}

func TestCatalogUnchangedWhenUpstreamHasNothingNew(t *testing.T) {
	e := newEnv(t)
	e.withCatalog()
	e.run("FAKE_BB_OUT=" + summary + `\n`)
	if log := read(t, filepath.Join(e.results, "last-run.log")); !strings.Contains(log, "catalog=unchanged") {
		t.Errorf("expected catalog=unchanged:\n%s", log)
	}
}
