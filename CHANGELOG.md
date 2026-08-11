## 2026-08-11 Zweig-Aufraeumen nach Bestandsaufnahme (Protokoll)

Anlass: Bestandsaufnahme vom 11.08. fand suite-weit rund 120 liegengebliebene
Zweige. Entscheidungsbaum: inhaltlich leer gegen den Hauptzweig (git diff/cherry) -> geloescht;
reine Doku-Zweige -> per PR uebernommen; alles andere -> erst als Tag archiv/<zweig>-2026-08-11
gesichert, dann geloescht (Wiederherstellung: git checkout -b <zweig> <tag>).
Zweige mit Commits juenger als 24h blieben unangetastet. Arbeitskopien nur entfernt,
wenn sauber und ohne aktive Sitzung.

- ZWEIG-REMOTE fix/sprechender-name-anmeldeobjekte | ARCHIVIERT+GELOESCHT | Tag archiv/fix/sprechender-name-anmeldeobjekte-2026-08-11

