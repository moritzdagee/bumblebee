#!/bin/zsh
# Bumblebee daily background scan — installed 2026-06-05, revised same day.
# Read-only. Refreshes its own private threat-catalog copy first, then runs
# TWO fast, targeted scans and notifies via macOS Notification Center ONLY
# if a match is found:
#   1) baseline  -> global/user installed packages, browser extensions,
#                   AI-tool (MCP) configs, Homebrew, Go modules
#   2) project   -> dependency lockfiles inside known dev folders
# It deliberately does NOT crawl the entire home folder (iCloud/Photos/caches),
# which under launchd is huge and slow. For a full $HOME sweep, run manually:
#   bumblebee scan --profile deep --root "$HOME" \
#     --exposure-catalog "$HOME/.local/share/bumblebee-catalogs/threat_intel" \
#     --findings-only
# Triggered by the LaunchAgent love.bios.bumblebee.daily.

set -uo pipefail

BB="$HOME/go/bin/bumblebee"
CAT="$HOME/.local/share/bumblebee-catalogs"      # private auto-updating catalog clone
CATDIR="$CAT/threat_intel"
RES="/Users/moritzcremer/local models/bumblebee-scan-results/daily"
mkdir -p "$RES"
TS="$(date +%Y%m%d-%H%M%S)"
OUT_BASE="$RES/findings-baseline-$TS.ndjson"
OUT_PROJ="$RES/findings-project-$TS.ndjson"
DIAG="$RES/diag-$TS.ndjson"
LASTLOG="$RES/last-run.log"

# Dev folders to scan for project dependencies (only those that exist).
PROJECT_ROOTS=("$HOME/code" "$HOME/local models" "$HOME/Developer" "$HOME/src" "$HOME/work" "$HOME/Projects")
ROOT_ARGS=()
for d in "${PROJECT_ROOTS[@]}"; do
  [ -d "$d" ] && ROOT_ARGS+=(--root "$d")
done

# 1. Refresh the private threat catalogs (best-effort; offline is fine).
git -C "$CAT" pull --quiet --depth 1 2>/dev/null || true

# 2a. Baseline exposure scan (global/user surface).
"$BB" scan --profile baseline --exposure-catalog "$CATDIR" --findings-only \
  --max-duration 3m > "$OUT_BASE" 2>> "$DIAG" || true

# 2b. Project exposure scan (dependency lockfiles in dev folders), if any exist.
if [ ${#ROOT_ARGS[@]} -gt 0 ]; then
  "$BB" scan --profile project "${ROOT_ARGS[@]}" --exposure-catalog "$CATDIR" \
    --findings-only --max-duration 3m > "$OUT_PROJ" 2>> "$DIAG" || true
else
  : > "$OUT_PROJ"
fi

# 3. Count real matches (finding records only) across both scans.
count_findings() { grep -c '"record_type":"finding"' "$1" 2>/dev/null | tr -dc '0-9'; }
NB="$(count_findings "$OUT_BASE")"; [ -z "$NB" ] && NB=0
NP="$(count_findings "$OUT_PROJ")"; [ -z "$NP" ] && NP=0
N=$((NB + NP))

echo "$(date '+%Y-%m-%d %H:%M:%S')  findings=$N (baseline=$NB project=$NP)" >> "$LASTLOG"

# 4. Alert only if something was found.
if [ "$N" -gt 0 ]; then
  /usr/bin/osascript -e "display notification \"$N supply-chain match(es) found. See $RES\" with title \"⚠️ Bumblebee found something\" sound name \"Basso\""
fi

# 5. Keep only the 30 most recent files of each kind.
for pat in 'findings-baseline-*.ndjson' 'findings-project-*.ndjson' 'diag-*.ndjson'; do
  ls -1t "$RES"/$pat 2>/dev/null | tail -n +31 | while read -r f; do rm -f "$f"; done
done
exit 0
