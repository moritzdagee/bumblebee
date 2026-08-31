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

