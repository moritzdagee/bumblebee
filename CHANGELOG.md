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

