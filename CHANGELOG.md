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

