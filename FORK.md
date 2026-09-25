# FORK.md — islamsc1/3x-ui "with-allow-insecure" fork

Must-read for any agent (Claude Code, OpenCode, Codex, …) asked to **update this
fork to a new upstream version**. Follow it step by step. Do not improvise a
different workflow — every past update used this one, and the install/update
scripts on real servers depend on its exact conventions (tag names, branch
name, URLs).

---

## 1. What this fork is

| Thing | Value |
|---|---|
| Upstream | `https://github.com/MHSanaei/3x-ui` (git remote `upstream`) |
| Fork | `https://github.com/islamsc1/3x-ui` (git remote `origin`) |
| Release branch | `with-allow-insecure` (on `origin`) |
| Release tags | `vX.Y.Z-allowinsecure` (e.g. `v3.8.5-allowinsecure`) |
| Work branch per update | `sdd/vXYZ-upgrade` (e.g. `sdd/v385-upgrade`) |
| Shape of the branch | **upstream release tag + exactly ONE fork commit** |

The fork is **one commit** on top of an upstream release tag. That commit does two
things:

1. **Feature**: restores the per-TLS `allowInsecure` option that upstream deleted.
2. **Plumbing**: points the install/update scripts and the in-panel updater at the
   fork's releases, because upstream binaries do not have the feature.

It is never merged with upstream. Each update re-creates the branch as
`<new upstream tag>` + `<cherry-pick of the previous fork commit>`, then
force-updates `origin/with-allow-insecure` and pushes a new tag.

### Version history

| Fork tag | Upstream base | Fork commit | Notes |
|---|---|---|---|
| `v3.3.1-allowinsecure(2)` | v3.3.1 | — | early |
| `v3.5.0-allowinsecure` | v3.5.0 | `8638085a` | older approach; also carried unrelated `x-ui.sh`/`IndexPage.tsx` edits — do not copy from it |
| `v3.7.0-allowinsecure` | v3.7.0 | `c13a63c6` | clean port, added tests + 13 locales |
| `v3.8.5-allowinsecure` | v3.8.5 | see `git log -1 v3.8.5-allowinsecure` | this doc added; outbound side removed (see §3.3); version-compare fix |

**Source of the next cherry-pick = the newest `vX.Y.Z-allowinsecure` tag.**
List them: `git tag -l 'v*-allowinsecure' --sort=-v:refname`.

---

## 2. Where the feature came from — upstream commit `511adffc`

`https://github.com/MHSanaei/3x-ui/commit/511adffc5bb419a938e3b1eb0087a49f62d76db6`
("Remove allowInsecure", MHSanaei, 2026-02-11). Upstream removed the deprecated
TLS `allowInsecure` flag because Xray-core removed it and replaced it with
`pinnedPeerCertSha256` (`pcs`) and `verifyPeerCertByName` (`vcn`).

That commit predates the React rewrite, so its file paths no longer exist. What
it removed, and where each piece lives now:

| Removed in 511adffc (old path) | Today's equivalent (restored by the fork) |
|---|---|
| `web/html/form/tls_settings.html` — inbound toggle | `frontend/src/pages/inbounds/form/security/tls.tsx` (Switch on `streamSettings.tlsSettings.settings.allowInsecure`) |
| `web/assets/js/model/inbound.js` — model field + link emitters | `frontend/src/schemas/protocols/security/tls.ts` (`TlsClientSettingsSchema.allowInsecure`, default `false`) and `frontend/src/lib/xray/inbound-link.ts` (vmess obj, vless/trojan/ss query `allowInsecure=1`, external proxy) |
| `sub/subService.go` — `allowInsecure=1` in links | `internal/sub/service.go`: `applyShareTLSParams`, `applyVmessTLSParams`, `applyExternalProxyTLSObj`, `applyExternalProxyTLSParams` |
| `sub/subJsonService.go` — JSON sub `tlsSettings.allowInsecure` | `internal/sub/json_service.go`: `tlsData` |
| `web/assets/js/model/outbound.js`, `web/html/form/outbound.html` — outbound toggle + link import | **intentionally NOT restored** — see §3.3 |

To see it: `git show 511adffc` (fetch `upstream` first).

---

## 3. The fork commit, file by file

### 3.1 Feature touch-points (must survive every update)

| File | What the fork adds |
|---|---|
| `frontend/src/schemas/protocols/security/tls.ts` | `allowInsecure: z.boolean().default(false)` in `TlsClientSettingsSchema` **and** `allowInsecure: false` in the `settings` default literal of `TlsStreamSettingsSchema` |
| `frontend/src/pages/inbounds/form/security/tls.tsx` | `FormField` + `Switch` for `['streamSettings','tlsSettings','settings','allowInsecure']`, label `pages.inbounds.form.allowInsecure`, tooltip `…allowInsecureTip` |
| `frontend/src/lib/xray/inbound-link.ts` | 5 emit sites, all gated on `true`: `applyExternalProxyTLSObj`, `genVmessLink`, `applyExternalProxyTLSParams`, `genVlessLink` (inline TLS block), `writeTlsParams` (Trojan/SS). Put the `allowInsecure` line right before the `fp` line. |
| `frontend/src/lib/xray/stream-wire-normalize.ts` | `dropFalseFlags(out, ['allowInsecure'])` and `dropFalseFlags(settingsOut, ['allowInsecure'])` in `normalizeTlsForWire` so `false` is never stored |
| `internal/sub/service.go` | 4 emitters listed in §2, gated on `true` |
| `internal/sub/json_service.go` | `tlsData` copies `allowInsecure` only when `true` |
| `internal/web/translation/*.json` (all 13) | `pages.inbounds.form.allowInsecure` + `allowInsecureTip` |
| Tests | `internal/sub/allowinsecure_test.go`; `frontend/src/test/inbound-link.test.ts` (allowInsecure cases); `frontend/src/test/golden/fixtures/security/tls-allowinsecure.json`; snapshot lines in `frontend/src/test/__snapshots__/{security.test.ts.snap,inbound-full.test.ts.snap,inbound-form-blocks.test.tsx.snap}` |

Rules of the feature:
- Flag off/absent → **nothing** emitted anywhere (links stay byte-identical to upstream).
- Flag on → `allowInsecure=1` (query), `"allowInsecure": true` (vmess JSON / JSON sub).
- Clash output gets `skip-cert-verify` for free (upstream `clash_service.go` reads `allowInsecure`).
- Hysteria links deliberately do not emit `insecure` from the inbound flag (byte-stable with Go output).

### 3.2 Things upstream already provides (do NOT duplicate)

Since v3.8.x upstream itself supports **host / external-proxy level**
`allowInsecure` (`ExternalProxyEntrySchema.allowInsecure`,
`internal/sub/host_sub.go`, tuic `allow_insecure`, hysteria `insecure`,
`cloneStreamForExternalProxy`). When a cherry-pick adds a key or emitter that
already exists upstream, **keep upstream's and drop the fork's duplicate**. Example
from v3.8.5: `external-proxy.ts` got `allowInsecure` twice (duplicate object key), so
the fork line was removed.

### 3.3 Outbound side: deliberately absent

The bundled Xray-core refuses TLS configs that contain `"allowInsecure": true`:

```
infra/conf/transport_security.go:
  if c.AllowInsecure { return nil, errors.PrintRemovedFeatureError(`"allowInsecure"`, …) }
```

The panel's own outbounds run inside that Xray, so an outbound toggle (or
importing a share link with `allowInsecure=1` into Outbounds) writes
`tlsSettings.allowInsecure: true` and **Xray fails to start**. Fork tags up to
`v3.7.0-allowinsecure` had this bug. From `v3.8.5-allowinsecure` onward
`frontend/src/pages/xray/outbounds/**` and `frontend/src/lib/xray/outbound-link-parser.ts`
are **identical to upstream**. Keep them that way. Inbound `tlsSettings.settings.*`
is panel-only (Xray ignores it), which is why the inbound side is safe.

Verify this is still true on each update:

```bash
grep -n -A1 'if c.AllowInsecure' "$(go list -m -f '{{.Dir}}' github.com/xtls/xray-core)/infra/conf/transport_security.go"
```

If that check is gone (Xray accepts it again), the outbound side could come back — ask the owner first.

Note for owners: JSON subscriptions with the flag on send `allowInsecure: true`
to client apps. A client running Xray-core newer than 2026-06 rejects it; older
clients / non-Xray clients accept it. That is a known, accepted trade-off.

### 3.4 Fork plumbing (update path — must be exact)

| File | Upstream | Fork |
|---|---|---|
| `install.sh`, `update.sh` | every `MHSanaei/3x-ui` | `islamsc1/3x-ui` |
| `install.sh`, `update.sh` | `script_ref="main"` (dev build) and `[[ "${ref}" == "main" ]] && return 0` in `require_repo_files` | `with-allow-insecure` in both places |
| `x-ui.sh` | `MHSanaei/3x-ui/main/{install,update}.sh` | `islamsc1/3x-ui/with-allow-insecure/{install,update}.sh` |
| `x-ui.sh` `installed_script_url` | `MHSanaei/3x-ui/v${ver}/x-ui.sh`, fallback `…/main/x-ui.sh` | `islamsc1/3x-ui/v${ver}-allowinsecure/x-ui.sh`, fallback `…/with-allow-insecure/x-ui.sh` (message says `using with-allow-insecure`) |
| `x-ui.sh` `delete_script` reinstall hint | `mhsanaei/3x-ui/master/install.sh` | `islamsc1/3x-ui/with-allow-insecure/install.sh` |
| `internal/web/service/panel/panel.go` | `panelUpdaterURL` → `MHSanaei/3x-ui/main/update.sh`; `fetchPanelRelease` → `repos/MHSanaei/3x-ui/releases/…` | `islamsc1/3x-ui/with-allow-insecure/update.sh`; `repos/islamsc1/3x-ui/releases/…` |
| `internal/web/service/panel/panel.go` | `normalizeVersionTag` strips `v` | also strips `forkReleaseTagSuffix = "-allowinsecure"` (+ test cases in `panel_test.go`). Without it the panel reports "update available" forever, because `v3.8.5-allowinsecure` is not parseable semver and falls back to string inequality. |
| `.github/workflows/docker.yml` | runs everywhere | job gated `if: github.repository == 'MHSanaei/3x-ui'` (fork has no registry secrets) |

**Script approach (important):** for `install.sh`, `update.sh`, `x-ui.sh` always take
**upstream's new file verbatim** and re-apply only the URL/ref substitutions above.
Never merge old fork script bodies into new upstream scripts — upstream rewrites
these files often (e.g. v3.8 added per-tag file pinning + sha256 verification).
Older fork commits (`8638085a`, `c13a63c6`) also replaced upstream's
`install()`/`update()`/`update_dev()` menu bodies with an older variant — that was
dropped in v3.8.5 on purpose. Only URLs differ from upstream now.

Why the tag suffix matters end to end:
- `release.yml` triggers on tags `v*.*.*`, so `v3.8.5-allowinsecure` builds and
  uploads `x-ui-linux-<arch>.tar.gz` + `.sha256` to the fork's GitHub release.
- `install.sh`/`update.sh` resolve `releases/latest` on the fork → tag
  `vX.Y.Z-allowinsecure` → download that release, then fetch `x-ui.sh`/`x-ui.rc`/unit
  files from `raw.githubusercontent.com/islamsc1/3x-ui/<that tag>/…`. The tag must
  exist on `origin` or `require_repo_files` aborts the update (safely, old install
  untouched).
- `x-ui -v` prints the plain upstream version (`3.8.5`), so `x-ui.sh` maps it back to
  `v3.8.5-allowinsecure`.

Known limitations (accepted, not bugs to "fix" during an update):
- The fork publishes no `dev-latest` release, so the dev channel
  (`x-ui` menu "update dev", panel dev channel) fails cleanly with "not available".
- `x-ui.sh legacy_version` installs a plain **upstream** version (fork has no old tags).
- Docs/README/`AppSidebar.tsx` repo links still point to upstream — cosmetic, leave them.

---

## 4. Update procedure (do exactly this)

Replace `X.Y.Z` with the target upstream version, `PREV` with the newest fork tag.

### 4.0 Environment gotchas
- If the repo lives on a Windows mount (`/mnt/c`, `/mnt/e`, …) `git status` shows
  hundreds of CRLF-only modifications. They are noise — **do not commit, reset or
  stash them**. Work in a **git worktree on the Linux filesystem** instead (faster, clean).
- Needs Go (version in `go.mod`), a C compiler (CGo SQLite), Node (`.nvmrc`), `golangci-lint`.
- Never commit directly on `with-allow-insecure`; build on `sdd/vXYZ-upgrade`.

### 4.1 Prepare

```bash
git fetch upstream --tags          # a "dev-latest would clobber existing tag" warning is harmless
git fetch origin --tags
git tag -l 'v*-allowinsecure' --sort=-v:refname | head -1      # = PREV
git log --oneline -1 vX.Y.Z                                    # upstream tag exists?
git log --oneline PREV^..PREV                                  # the single fork commit
git worktree add ~/wt/3x-ui-vXYZ -b sdd/vXYZ-upgrade vX.Y.Z
cd ~/wt/3x-ui-vXYZ
```

Sanity: `git rev-list --count PREV^..PREV` must be `1` and `PREV^` must be an upstream
release tag (`git describe --tags PREV^`). If not, stop and ask.

### 4.2 Cherry-pick the previous fork commit

```bash
git cherry-pick PREV
```

Resolve conflicts with these rules:

| Conflict in | Resolution |
|---|---|
| `install.sh`, `update.sh`, `x-ui.sh` | `git checkout --ours -- <file>` (= new upstream), then re-apply §3.4 substitutions (commands below) |
| `internal/web/service/panel/panel.go` | new upstream code + the 3 URL swaps + `forkReleaseTagSuffix` in `normalizeVersionTag` |
| TS/Go feature files | keep **all** new upstream logic, add only the `allowInsecure` lines from §3.1 (next to the `fp`/fingerprint line). Keep upstream defaults (e.g. v3.8.5 changed TLS `fingerprint` default `'chrome'` → `''`; keep `''`). |
| Duplicate of something upstream now has | keep upstream, drop fork copy (§3.2) |
| Snapshots `*.snap` | take upstream (`--ours`), then regenerate only the allowInsecure-related entries — see 4.4 |
| Anything outbound (`pages/xray/outbounds/**`, `outbound-link-parser.ts`) | must end **identical to upstream** (`git checkout vX.Y.Z -- <file>`) |

Script substitutions (run after taking upstream's scripts):

```bash
sed -i 's#MHSanaei/3x-ui#islamsc1/3x-ui#g' install.sh update.sh x-ui.sh
sed -i 's#\[\[ "${ref}" == "main" \]\] && return 0#[[ "${ref}" == "with-allow-insecure" ]] \&\& return 0#; s#script_ref="main"#script_ref="with-allow-insecure"#' install.sh update.sh
sed -i 's#islamsc1/3x-ui/main/#islamsc1/3x-ui/with-allow-insecure/#g; s#islamsc1/3x-ui/v${ver}/x-ui.sh#islamsc1/3x-ui/v${ver}-allowinsecure/x-ui.sh#g; s#published for the installed version (${ver:-unknown}), using main#published for the installed version (${ver:-unknown}), using with-allow-insecure#; s#raw.githubusercontent.com/mhsanaei/3x-ui/master/install.sh#raw.githubusercontent.com/islamsc1/3x-ui/with-allow-insecure/install.sh#' x-ui.sh
```

Then **read** `git diff vX.Y.Z -- install.sh update.sh x-ui.sh`: every changed line
must be a URL/ref swap. If upstream renamed the `main` logic or added new download
URLs, adapt by hand following §3.4.

### 4.3 Invariant checks (all must pass)

```bash
# 1. no upstream download/update URL left in the update path
git grep -n -i 'mhsanaei/3x-ui' -- install.sh update.sh internal/web/service/panel/panel.go
git grep -n -i 'mhsanaei/3x-ui' -- x-ui.sh          # only legacy_version may remain
git grep -n '"main"\|/main/' -- install.sh update.sh x-ui.sh | grep -i 'ref\|raw.github'   # expect nothing
# 2. feature present
git grep -c 'allowInsecure' -- frontend/src/lib/xray/inbound-link.ts internal/sub/service.go internal/sub/json_service.go frontend/src/schemas/protocols/security/tls.ts frontend/src/pages/inbounds/form/security/tls.tsx
for f in internal/web/translation/*.json; do grep -q '"allowInsecureTip"' "$f" || echo "MISSING $f"; done
# 3. outbound side identical to upstream
git diff --stat vX.Y.Z -- frontend/src/pages/xray/outbounds frontend/src/lib/xray/outbound-link-parser.ts   # expect empty
# 4. xray-core still rejects allowInsecure (see 3.3)
# 5. no conflict markers, no duplicate keys
git grep -n '^<<<<<<<\|^>>>>>>>' ; git diff vX.Y.Z | grep '^+' | grep -c allowInsecure
# 6. fork-only files survived
test -f FORK.md && grep -q 'FORK.md' CLAUDE.md && grep -q forkReleaseTagSuffix internal/web/service/panel/panel.go
```

### 4.4 Tests and gate

```bash
go test ./internal/sub/ ./internal/web/service/panel/
cd frontend && npm ci && npm test && cd ..
make verify           # full CI mirror: gen-check lint format-check typecheck tests build
```

- `npm test` includes browser-mode files (Playwright Chromium). If it fails with
  `Executable doesn't exist` run `npx playwright install chromium-headless-shell`.
  If it then fails with `libnspr4.so: cannot open shared object file` and there is
  no sudo, fetch the libs without root and re-run `make verify` with them:
  ```bash
  mkdir -p ~/.local/pwlibs/debs && cd ~/.local/pwlibs/debs
  apt-get download libnspr4 libnss3 libasound2 && for d in *.deb; do dpkg-deb -x "$d" ~/.local/pwlibs/root; done
  export LD_LIBRARY_PATH=~/.local/pwlibs/root/usr/lib/x86_64-linux-gnu
  ```
  A red test file count (e.g. `137 passed (164)`) means tests did **not** all run —
  never report that as green.
- A Go test helper signature changed upstream (e.g. `NewSubJsonService` gained
  params in v3.8.5) → fix the call in `allowinsecure_test.go`, never the production code.
- Snapshot failures: only regenerate (`cd frontend && npx vitest run -u`) when the
  diff of the `.snap` file is **only** allowInsecure-related lines. Read
  `git diff -- '*.snap'` after regenerating. Any other snapshot change = real bug.
- `make gen` must leave `frontend/src/generated` + `frontend/public/openapi.json` clean.
- No `//` comments in new Go/TS code (repo hard rule).

### 4.5 Commit (one commit, same message shape every time)

```bash
git add -A && git status --short          # review: nothing unexpected
git -c core.editor=true cherry-pick --continue    # or git commit if the pick was already concluded
git commit --amend                         # rewrite message using the template below
```

Message template:

```
feat: restore allowInsecure TLS option on upstream vX.Y.Z (upstream 511adffc removal)

<why: fork-only feature, see FORK.md>
<touch-point map (copy §3.1 list, updated)>
<conflicts resolved in this port + any upstream overlap dropped>
<fork plumbing summary (§3.4)>
Verification: make verify green; invariant checks from FORK.md §4.3 clean.
```

### 4.6 Publish — needs explicit owner approval (force-push)

Ask the owner before running these. They rewrite the public release branch.

```bash
git tag vX.Y.Z-allowinsecure
git push origin vX.Y.Z-allowinsecure                       # triggers release.yml → release assets
git push --force-with-lease=with-allow-insecure:$(git rev-parse origin/with-allow-insecure) origin HEAD:with-allow-insecure
```

Push the **tag first**: until the release exists, `releases/latest` still points to
the previous fork release, and its tag still has its own scripts, so servers
updating mid-release stay consistent.

### 4.7 Post-release verification

```bash
gh release view vX.Y.Z-allowinsecure -R islamsc1/3x-ui    # assets: x-ui-linux-{amd64,arm64,...}.tar.gz + .sha256
gh release list -R islamsc1/3x-ui | head -3                # new one is "Latest"
curl -fsIL https://raw.githubusercontent.com/islamsc1/3x-ui/vX.Y.Z-allowinsecure/x-ui.sh | head -1
curl -fsSL https://raw.githubusercontent.com/islamsc1/3x-ui/with-allow-insecure/update.sh | grep -c islamsc1
```

If `gh release view` shows the release as a draft/pre-release, or missing
assets, check the Actions run of `release.yml` before announcing.

Finally clean up: `git worktree remove ~/wt/3x-ui-vXYZ` (after pushing), and
update the version-history table in §1 of this file in the **next** port.
