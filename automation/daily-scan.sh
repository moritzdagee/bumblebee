#!/bin/zsh
# Bumblebee daily background scan — installed 2026-06-05, rev3.
# Read-only. Refreshes its own private threat-catalog copy first, then scans
# the WHOLE home folder EXCEPT a deny-list of heavy, dependency-free trees
# (Photos library, caches, app containers, Xcode build data, trash). Notifies
# via macOS Notification Center ONLY if a match is found.
# Triggered by the LaunchAgent love.bios.bumblebee.daily.

set -uo pipefail

BB="$HOME/go/bin/bumblebee"
CAT="$HOME/.local/share/bumblebee-catalogs"      # private auto-updating catalog clone
CATDIR="$CAT/threat_intel"
RES="/Users/moritzcremer/local models/bumblebee-scan-results/daily"
mkdir -p "$RES"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$RES/findings-$TS.ndjson"
DIAG="$RES/diag-$TS.ndjson"
LASTLOG="$RES/last-run.log"

# Heavy, dependency-free directory names to skip (matched by name or suffix).
EXCLUDES=(
  "Photos Library.photoslibrary"   # Photos library — 100k+ media files
  "Caches" ".cache"                # caches everywhere
  "Containers" "Group Containers"  # sandboxed app data
  "CoreSimulator" "DerivedData"    # Xcode simulators & build output
  ".Trash" "MobileSync"            # trash & iPhone backups
  "Photo Booth Library.photoslibrary"
)
EX_ARGS=()
for e in "${EXCLUDES[@]}"; do EX_ARGS+=(--exclude "$e"); done

# 1. Refresh the private threat catalogs (best-effort; offline is fine).
git -C "$CAT" pull --quiet --depth 1 2>/dev/null || true

# 2. Deep, read-only exposure scan of the whole home folder minus the deny-list.
"$BB" scan --profile deep --root "$HOME" --exposure-catalog "$CATDIR" \
  --findings-only "${EX_ARGS[@]}" --max-duration 15m > "$OUT" 2> "$DIAG" || true

# 3. Count real matches (finding records only).
N="$(grep -c '"record_type":"finding"' "$OUT" 2>/dev/null | tr -dc '0-9')"; [ -z "$N" ] && N=0

# Pull files_considered + duration from the summary for the run log.
SUMMARY="$(grep '"record_type":"scan_summary"' "$OUT" 2>/dev/null | tail -1)"
FILES="$(printf '%s' "$SUMMARY" | sed -n 's/.*"files_considered":\([0-9]*\).*/\1/p')"
DUR="$(printf '%s' "$SUMMARY" | sed -n 's/.*"duration_ms":\([0-9]*\).*/\1/p')"
echo "$(date '+%Y-%m-%d %H:%M:%S')  findings=$N  files=${FILES:-?}  duration_ms=${DUR:-?}" >> "$LASTLOG"

# 4. Alert only if something was found.
if [ "$N" -gt 0 ]; then
  /usr/bin/osascript -e "display notification \"$N supply-chain match(es) found. See $RES\" with title \"⚠️ Bumblebee found something\" sound name \"Basso\""
fi

# 5. Keep only the 30 most recent files of each kind.
for pat in 'findings-*.ndjson' 'diag-*.ndjson'; do
  ls -1t "$RES"/$pat 2>/dev/null | tail -n +31 | while read -r f; do rm -f "$f"; done
done
exit 0
