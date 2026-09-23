## 2026-09-23 Scanner neu installiert (Stand 329b871), Katalog auf 0.2.0 umgeschaltet

Installation ausserhalb von Git, 2026-09-23 12:47:59 -03 (MacBook):

```sh
# Backup des alten Scanners (v0.1.2-0.20260602133442-156df7a272c9,
# sha256 03183e328e8596dac326c8188b8a3b34c2287f8ef64b712519f33ace1c17085f)
mkdir -p ~/.claude/backups/bumblebee-binary-vor-upstream-sync-2026-09-23
cp -p ~/go/bin/bumblebee ~/.claude/backups/bumblebee-binary-vor-upstream-sync-2026-09-23/bumblebee
# Neubau aus dem gemergten Hauptstand (PR #12)
cd ~/bumblebee && go install ./cmd/bumblebee
```

- Ziel: `~/go/bin/bumblebee` (von `automation/bumblebee-scan` als `$BB`
  benutzt; der Autostart `claude.macbook.bumblebee.daily` startet das Skript
  direkt aus `~/bumblebee/automation/`, an ihm wurde nichts geaendert).
- Neu: `v0.1.2-0.20260923154611-329b871c385e`, sha256
  `f69f23b7ac16de6248f349f4191cea400eef5218eaabf5f4f9bf6fc6df1fba01`,
  `bumblebee selftest` OK.
- Zurueck: Backup-Datei nach `~/go/bin/bumblebee` kopieren.

Probelauf 2026-09-23 12:48 -03: `BUMBLEBEE_WA=<Ersatzprogramm, das nur
mitschreibt> ~/bumblebee/automation/bumblebee-scan` (kein launchd, keine
echte Nachricht; das Ersatzprogramm wurde nicht aufgerufen, weil nichts zu
melden war). Ergebnis in `last-run.log`:
`findings=0 files=757659 duration_ms=10755 catalog=updated:8ef7fbc catalog_date=2026-09-23`.
Der private Katalog-Klon steht damit auf Upstream `8ef7fbc`: 13 Kataloge,
alle Schema 0.2.0, 1101 Eintraege (vorher 9 Kataloge, Schema 0.1.0, 686
Eintraege). Das Aufraeumen aus PR #11 hat gegriffen: 30 statt 121
findings-Dateien.

## 2026-09-23 Originalprojekt zusammengefuehrt: Scanner liest Katalog-Schema 0.2.0

Der installierte Scanner (Stand 2026-06-02, Upstream-Commit 156df7a) kennt nur
Katalog-Schema 0.1.0. Upstream hat alle Kataloge auf 0.2.0 gehoben, deshalb
meldete der Tages-Scan seit dem Katalog-Fix `catalog=NOT-UPDATED(scanner-too-old
...)` und die Bedrohungsliste blieb beim Stand 2026-06-02.

- Zusammengefuehrt: `perplexityai/bumblebee` main bis `d76e369` (8 Commits:
  Schema 0.2.0 mit any-version-Eintraegen `"*"`, Kataloge GlassWorm, Mastra,
  Mini Shai-Hulud LeoPlatform, schnelleres Verzeichnis-Durchlaufen, neuer
  Release-Ablauf in `release.yml`).
- Konflikte: keine. Upstream beruehrt `internal/exposure`, `internal/osv`,
  `internal/walk`, `internal/scanner`, `threat_intel/`, `docs/schema/v0.2.0`,
  README, CONTRIBUTING, `release.yml`; die eigenen Anpassungen liegen in
  `automation/`, `internal/fsread`, `internal/npmproject`, den 13
  Ecosystem-Paketen (Delegation an fsread), `tests.yml`, CHANGELOG und
  INSTALL-NOTES. Git hat README automatisch zusammengefuehrt; der Abschnitt
  "Automatische Tests" ist erhalten.
- Eigene Logik geprueft und unveraendert: WhatsApp-Meldung (`melden()`),
  Katalog-Probe vor dem Umschalten, `SCAN-FEHLGESCHLAGEN`-Meldung.
- Probe wie im Tages-Skript (leerer Ordner gegen die Upstream-Kataloge): neuer
  Scanner `status":"complete"`, Exit 0; alter Scanner `unsupported exposure
  catalog schema_version "0.2.0"`. Die Katalog-Probe wird mit dem neuen
  Scanner also umschalten.
- `go build ./...`, `go vet ./...`, `go test ./...` gruen, `gofmt -l .` leer,
  `bumblebee selftest` OK.
- `release.yml` von Upstream laeuft nur bei Tags `v*` oder manuell, nicht bei
  Pull Requests oder Pushes auf main.

## 2026-09-23 Automatische Tests fuer Vorschlaege (tests.yml) und Aufraeum-Fehler im Tages-Scan

- Neu: `.github/workflows/tests.yml`. Laeuft nur bei Pull Requests, nicht bei
  reinen Doku-Aenderungen (`**.md`, `docs/**`, `archive/**`). Ein neuer Push
  bricht den laufenden Lauf ab (concurrency), Zeitlimit 15 Minuten. Go-Version
  aus `go.mod`, mit Zwischenspeicher. Schritte: `go vet ./...`, `go test ./...`
  und `shellcheck --shell=bash` fuer die Shell-Skripte unter `automation/`
  (shellcheck kennt kein zsh, deshalb bash-Modus).
- `ci.yml` des Originalprojekts bleibt unveraendert und laeuft bei PRs
  ebenfalls (Linux und macOS, gofmt, Race-Test, Selbsttest, govulncheck). Es
  deckt go vet/test schon ab, hat aber keine Pfad-Ausnahmen, keinen
  Abbruch alter Laeufe und kein shellcheck; darum die zusaetzliche Datei statt
  einer Aenderung an `ci.yml` (haelt Upstream-Zusammenfuehrungen konfliktfrei).
- Kein Test musste uebersprungen werden: `go test ./...` braucht weder Netz
  noch macOS.
- Fund von shellcheck, behoben: Schritt 5 von `automation/bumblebee-scan`
  sollte nur die 30 neuesten Ergebnis- und Diagnose-Dateien behalten, benutzte
  aber `ls -1t "$RES"/$pat`. zsh expandiert ein Muster in einer Variablen
  nicht, `ls` bekam das woertliche Muster, scheiterte still, und es wurde nie
  etwas geloescht (121 findings-Dateien lagen in
  `~/bumblebee-scan-results/daily/`). Jetzt `find ... -name "$pat" | sort -r`,
  die Dateinamen tragen einen Zeitstempel. Beim naechsten Lauf werden die
  aelteren Dateien ueber 30 hinaus entfernt, wie urspruenglich vorgesehen.
  Getestet in einer Sandbox mit je 40 Dateien: es bleiben die 30 neuesten.
- README: Abschnitt "Automatische Tests (moritzdagee-Fork)" am Ende.

## 2026-09-23 Treffer-Meldung ueber WhatsApp statt Mac-Mitteilung

Befund der Fehler-Durchsicht vom 2026-09-23. `automation/bumblebee-scan`
meldete Treffer per `osascript display notification`, also ueber die
Mac-Mitteilungszentrale. Das verstoesst gegen Suite-Regel 0e (seit 2026-08-03:
keine Mac-Mitteilungen, einziger Push-Kanal ist die WhatsApp-Gruppe
"Moritz und Ted"); die Umstellung des Mac-Checkups galt als letzte, dieser
Sender war uebersehen.

- Neu: Funktion `melden()` ruft `~/assistant_dataflow/scripts/wa send
  "Moritz und Ted" <text> --ja` mit `/usr/bin/python3` auf. Gemeldet wird bei
  Treffern und (neu) bei abgebrochenem Scan. Kein Ausweichen auf andere
  Kanaele; scheitert die Zustellung, steht `WHATSAPP-ZUSTELLUNG-FEHLGESCHLAGEN`
  mit dem Text in `last-run.log`.
- Fuer Tests laesst sich das Sende-Programm ueber `BUMBLEBEE_WA` ersetzen.
- INSTALL-NOTES.md angepasst.

Test: Sandbox-Kopie mit abgebrochenem Scan und Ersatz-Sendeprogramm: Aufruf
`send "Moritz und Ted" "<text>" --ja` kommt an; mit scheiterndem
Ersatz-Sendeprogramm steht die Fehlzustellung in `last-run.log`. Es wurde
dabei keine echte Nachricht gesendet.

## 2026-09-23 Taeglicher Scan: Katalog-Aktualisierung und Fehler nicht mehr still

Befund der Fehler-Durchsicht vom 2026-09-23.

- `automation/bumblebee-scan` aktualisierte den privaten Bedrohungs-Katalog
  (`~/.local/share/bumblebee-catalogs`) mit `git pull --depth 1 ... || true`.
  Bei einem flachen Klon hat jeder neue Upstream-Stand keinen gemeinsamen
  Vorfahren mit dem alten, der Pull brach deshalb bei JEDEM Lauf mit
  "divergent branches" ab, und `|| true` verschluckte das. Der Katalog stand
  seit 2026-06-02 still (Upstream ist bei 2026-08-07, drei neue Kataloge).
- Neu: `git fetch --depth 1 origin main`, dann Probe-Scan des neuen Katalogs
  mit dem installierten Scanner, erst danach `reset --hard`. Grund: Upstream
  hat die Kataloge auf Schema 0.2.0 gehoben, der installierte Scanner
  (Stand 2026-06-02) kennt nur 0.1.0 und bricht dann den ganzen Scan ab.
  Ohne die Probe haette die reparierte Aktualisierung jeden Scan zerstoert.
- Der Scan-Aufruf endete ebenfalls auf `|| true`. Ein abgebrochener Scan
  schreibt keine `scan_summary`, wurde aber als "findings=0" gezaehlt und sah
  aus wie ein sauberer Tag. Jetzt: fehlt die `scan_summary`, steht
  `SCAN-FEHLGESCHLAGEN` mit Fehlertext in `last-run.log` und das Skript endet
  mit Status 1 (sichtbar in `launchctl list`).
- `last-run.log` zeigt je Lauf den Katalog-Status (`catalog=updated|unchanged|
  NOT-UPDATED(...)|fetch-failed`) und das Katalog-Datum.

Test: Sandbox-Kopie des Skripts (eigene Ergebnis- und Katalog-Ordner, leerer
Scan-Ordner) in drei Faellen: alter Scanner + neuer Katalog -> Katalog bleibt,
Log `NOT-UPDATED(scanner-too-old ...)`; kaputter Katalog -> `SCAN-FEHLGESCHLAGEN
exit=2`, Status 1; aus upstream/main gebauter Scanner -> `catalog=updated:d76e369`.

Offen (Vorschlag, nicht Teil dieser Aenderung): Scanner auf Upstream-Stand
bringen, damit er Schema 0.2.0 liest; bis dahin meldet das Log den Rueckstand
taeglich.

## 2026-09-03 Ein Dateileser und ein package.json-Leser statt 15 Kopien

Befund des Code-Reviews der ganzen Suite vom 2026-09-02 (Funde 21 und 26,
Wiederverwendung).

- `readBounded()` (oeffnen, Stat, nur regulaere Dateien, Groessengrenze,
  ReadAll) lag byte-identisch in 13 Ecosystem-Paketen: mcp, editorext,
  composer, pypi, gomod, npm, yarn, homebrew, bun, rubygems, skills, pnpm,
  browserext. Eine Aenderung an der Groessen- oder Symlink-Regel haette an 13
  Stellen synchron erfolgen muessen; ein vergessener Ort haette Scanner
  uneinig gemacht, ohne dass ein Test es merkt. Jetzt `internal/fsread`
  (`ReadBounded`, `Diag`); die 13 Methoden delegieren nur noch.
- `loadDirectDeps()` (package.json, vier Dependency-Abschnitte) lag
  wortgleich in yarn und bun. Jetzt `internal/npmproject` (`LoadDirectDeps`).

Tests: neu `internal/fsread/fsread_test.go` (3) und
`internal/npmproject/npmproject_test.go` (3), erst rot, dann gruen.
`go build`, `go vet` und `go test ./...` (25 Pakete) gruen. Unbenutzte
Importe per goimports entfernt.

## 2026-08-31 Autostart-Name auf Suite-Namensschema umgestellt

`love.bios.bumblebee.daily` heisst jetzt `claude.macbook.bumblebee.daily`
(Suite-Regel 0i vom 2026-08-31: Name zeigt Autor, Ausfuehrungsort, Zweck).
Nur das Etikett und der Plist-Dateiname wurden geaendert; das ausgefuehrte
Programm (`automation/bumblebee-scan`) blieb identisch, erteilte
macOS-Freigaben bleiben deshalb gueltig. Angepasst: INSTALL-NOTES.md,
automation/bumblebee-scan, Plist umbenannt. Ausserhalb Git: Live-Plist in
~/Library/LaunchAgents getauscht (bootout alt / bootstrap neu); Backup der
alten Datei unter ~/.claude/backups/launchagents-vor-namensschema-2026-08-31/.

## 2026-08-11 Zweig-Aufraeumen nach Bestandsaufnahme (Protokoll)

Anlass: Bestandsaufnahme vom 11.08. fand suite-weit rund 120 liegengebliebene
Zweige. Entscheidungsbaum: inhaltlich leer gegen den Hauptzweig (git diff/cherry) -> geloescht;
reine Doku-Zweige -> per PR uebernommen; alles andere -> erst als Tag archiv/<zweig>-2026-08-11
gesichert, dann geloescht (Wiederherstellung: git checkout -b <zweig> <tag>).
Zweige mit Commits juenger als 24h blieben unangetastet. Arbeitskopien nur entfernt,
wenn sauber und ohne aktive Sitzung.

- ZWEIG-REMOTE fix/sprechender-name-anmeldeobjekte | ARCHIVIERT+GELOESCHT | Tag archiv/fix/sprechender-name-anmeldeobjekte-2026-08-11

