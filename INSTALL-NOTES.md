# Bumblebee — Install & First-Scan Notes (moritzdagee fork)

> Audit trail for the local install of Perplexity's Bumblebee supply-chain
> scanner on Moritz's MacBook Air. Kept in this fork per the
> "document everything in GitHub" rule. **No machine inventory / scan output
> is committed here** — this fork is public, so raw scan results are stored
> only in a local folder outside the repo (see "Scan results" below).

## What this is

[Bumblebee](https://github.com/perplexityai/bumblebee) is a **read-only**
inventory + exposure scanner for developer laptops. It reads on-disk package,
extension, and dev-tool metadata and flags exact matches against known
software supply-chain compromises (e.g. the "shai-hulud" / mini-shai-hulud
npm/PyPI/RubyGems/Go/Composer campaigns). It never executes package managers,
never runs install scripts, never reads source files, and makes no network
calls.

- Upstream: `perplexityai/bumblebee` (Apache-2.0)
- This fork: `moritzdagee/bumblebee` (your own copy)
- Language: Go (built with go1.26.4; tool requires Go 1.25+)
- Installed binary version: `v0.1.2-0.20260602133442-156df7a272c9`

## Install steps performed — 2026-06-05

All commands run on macOS (Apple Silicon, zsh).

```sh
# 1. Install Go (was missing) via Homebrew
brew install go                        # -> go1.26.4 darwin/arm64

# 2. Fork upstream to the user's account
gh repo fork perplexityai/bumblebee --clone=false   # -> github.com/moritzdagee/bumblebee

# 3. Clone the fork into the working directory
#    origin = moritzdagee/bumblebee, upstream = perplexityai/bumblebee
gh repo clone moritzdagee/bumblebee "/Users/moritzcremer/local models/bumblebee"

# 4. Build + install from the fork checkout
cd "/Users/moritzcremer/local models/bumblebee"
go build -o bumblebee ./cmd/bumblebee  # local binary in the checkout
go install ./cmd/bumblebee             # -> /Users/moritzcremer/go/bin/bumblebee

# 5. Put the command on PATH (see "System change" below)

# 6. Verify
bumblebee selftest                     # -> selftest OK (5 findings in 5ms)
bumblebee version
```

## System change outside this repo (logged per house rules)

`~/.zshrc` was appended with the following, at 2026-06-05 ~20:38 local:

```sh
# Added 2026-06-05 for Bumblebee (Perplexity supply-chain scanner) — puts go-installed CLIs on PATH
export PATH="$HOME/go/bin:$PATH"
```

- Reason: `go install` places binaries in `~/go/bin`, which was not on PATH.
- Backup of the pre-edit file: `~/.zshrc.bak-bumblebee-20260605-203838`
- Restore: copy that backup back over `~/.zshrc`.

## First scan results — 2026-06-05 (CLEAN)

Run against all bundled threat-intel catalogs in `threat_intel/`
(shai-hulud / mini-shai-hulud, gemstuffer, trapdoor crypto-stealer,
node-ipc credential stealer, nx-console, laravel-lang, shopsprint typosquat).

| Scan | Scope | Items examined | Findings |
|---|---|---|---|
| baseline (inventory) | system/user package roots, browser ext, MCP configs | 4,908 | — |
| baseline (exposure) | same vs. all catalogs | 4,908 | **0** |
| deep (exposure) | entire `$HOME` vs. all catalogs (483,874 files) | 19,920 | **0** |

No exposure to any catalogued supply-chain compromise was found.

Raw scan output (NDJSON) is **not** in this repo. It lives locally at:
`/Users/moritzcremer/local models/bumblebee-scan-results/`

## Re-running a scan later

```sh
cd "/Users/moritzcremer/local models/bumblebee"
git pull upstream main                 # refresh threat catalogs from upstream
go install ./cmd/bumblebee             # rebuild if source changed
bumblebee scan --profile baseline --exposure-catalog ./threat_intel --findings-only
```

## Uninstall

```sh
rm /Users/moritzcremer/go/bin/bumblebee          # remove the command
# remove the PATH line from ~/.zshrc (or restore the backup above)
rm -rf "/Users/moritzcremer/local models/bumblebee"   # remove the source copy
```
