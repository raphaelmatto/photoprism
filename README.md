# PhotoPrism Fork — raphaelmatto

**Last Updated:** October 28, 2026

Personal fork of [PhotoPrism](https://github.com/photoprism/photoprism) with display quality enhancements for retina/HiDPI screens.

## Features Added

### Display Settings (Settings > Content > Display)

Three new settings that work independently of each other:

- **Original Images** — Serve original image files in the lightbox viewer instead of generated thumbnails. Preserves per-area sharpening and full resolution from tools like Photoshop.
- **Retina Lightbox** — Scale lightbox images by `1/devicePixelRatio` for pixel-perfect quality on HiDPI displays (e.g. Apple Retina). Works with both originals and thumbnails.
- **Retina Thumbnails** — Use higher-resolution tiles in grid views. Mosaic and list views use `tile_500` instead of `tile_224`, cards view uses `tile_1080` instead of `tile_500`.

All settings take effect immediately without a server restart. They are stored in `settings.yml` under the `display` key.

### `photoprism dump-edits` (CLI)

New subcommand that lists every photo whose title, caption, taken-at, location, keywords, subject, notes, artist, or copyright has been manually edited in the web UI but not yet round-tripped back to the source file. Emits markdown by default or `--json` for tooling. See [Round-Tripping Edits to Lightroom](#round-tripping-edits-to-lightroom).

### ARM64 Docker Build

Production images build from `docker/photoprism/resolute/Dockerfile` on a native ARM64 GitHub Actions runner. Development and production use dated PhotoPrism base images for Ubuntu 26.04 LTS (Resolute Raccoon), replacing the unsupported Ubuntu 25.10 (Questing Quokka) base without updating the fork's application code.

An earlier custom ARM64 Dockerfile worked around broken TensorFlow headers in the deprecated `photoprism/develop:bookworm` base image. See [upstream issue #5444](https://github.com/photoprism/photoprism/issues/5444).

## Branch Structure

| Branch | Purpose |
|--------|---------|
| `production` | Integration and deploy branch. It should always contain everything running on the server, including code from feature branches with open upstream PRs. |
| `feature/*` | Short-lived feature branches created from `production`. These branches may contain both upstream-safe commits and fork-only commits while the feature is being built. |
| `pr/*` | Optional clean branches or worktrees used only to prepare a scoped upstream PR from cherry-picked upstream-safe commits. |
| `develop` | Local sync branch for `upstream/develop` when needed. Do not treat this as the fork's main integration branch. |
| `release` | Upstream release branch (untouched). |

## Remotes

| Remote | URL |
|--------|-----|
| `origin` | `git@github.com:raphaelmatto/photoprism.git` (this fork) |
| `upstream` | `https://github.com/photoprism/photoprism.git` (official repo) |

## CI/CD — Automated Docker Builds

A GitHub Actions workflow (`.github/workflows/build-production.yml`) builds and pushes an ARM64 Docker image on every push to the `production` branch.

- **Runner**: `ubuntu-24.04-arm` (native ARM64, no emulation)
- **Registry**: `ghcr.io/raphaelmatto/photoprism:latest`
- **Immutable tag**: `ghcr.io/raphaelmatto/photoprism:sha-<git-sha>`
- **Dockerfile**: `docker/photoprism/resolute/Dockerfile`
- **Build time**: ~5 minutes

The workflow uses `GITHUB_TOKEN` for registry auth, publishes both `latest` and commit-specific `sha-...` tags, and can also be triggered manually with `workflow_dispatch`.

### Checks Before Publishing

Pushes to `production` and `feature/*`, and pull requests targeting `production`, run three ARM64 checks in the same dated development base used locally:

- **Frontend:** Install locked dependencies and run the complete Vitest suite, with two workers to limit memory use.
- **Backend:** Compile PhotoPrism and run `make test-ci-go`, covering `pkg/`, photo search, FFmpeg helpers, and thumbnails. Tests use an isolated SQLite database and a temporary `/photoprism` volume.
- **Lint:** Block new Go lint findings relative to the previous push or pull request base, and ESLint errors in changed frontend source files. Full Go and frontend lint reports are uploaded as the `lint-reports` artifact, including existing findings outside the gate.

The image publishing job requires all three checks to pass and only publishes from `production`. Feature branches and pull requests run checks without publishing images. CodeQL separately scans Go and JavaScript on `production` and its pull requests.

The focused backend check is not the full backend suite. The initial full-suite run found stale expectations for fork settings and metadata, AI score differences, and tests requiring additional integration services. Those remain maintenance work; the publication gate does not claim they pass. The normal `make lint-go` and `make lint-js` commands now fail on lint errors, including existing ones.

## Branch Policy

Use this workflow to keep upstream contribution work cheap while preserving a healthy deploy branch:

1. Start every new feature from `production`.
2. Develop on a short-lived `feature/<name>` branch, not directly on `develop`.
3. Keep commits intentionally separated:
   - upstream-safe commits that could be submitted to PhotoPrism.
   - fork-only commits for deploy workflow, documentation, theme tweaks, or anything else that should remain fork-specific.
4. Merge or fast-forward the full feature branch back into `production` so `production` always remains a superset of open PR branches.
5. Create the upstream PR branch by cherry-picking only the upstream-safe commits onto `upstream/develop` or another appropriate upstream base.

The main rule is: do not rely on untangling a large mixed working tree later. Small, scoped commits are much cheaper to reuse than file-by-file extraction after the fact.

For upstream PR branches, treat these files as fork-only unless the PR is explicitly about fork operations:

- `README.md`
- `AGENTS.md`
- `.github/workflows/build-production.yml`

Before opening or updating an upstream PR, run:

```bash
./scripts/check-upstream-pr-safety.sh develop origin/feature/<name>
```

The script automatically flags:

- fork-only files in the diff
- fork registry and fork GitHub URLs derived from the `origin` remote
- SSH fork remotes
- absolute host paths such as `/Users/...`, `/Volumes/...`, and `/private/tmp/...`

You can add a repo-local privacy regex for names, places, or other sensitive strings without committing those terms:

```bash
UPSTREAM_PR_EXTRA_PATTERN='name-one|name-two|town-name' \
  ./scripts/check-upstream-pr-safety.sh develop origin/feature/<name>
```

If the script fails, either move the fork-only change back to `production` or replace the fixture/text with a generic placeholder before pushing the PR branch.

## Dev Workflow

Run these commands from the repository root. On macOS, start Docker Desktop with `open -a Docker` if the Docker daemon is stopped.

Start the containers and the PhotoPrism service without an interactive shell:

```bash
docker compose up -d photoprism mariadb traefik dummy-webdav dummy-oidc
docker compose exec -T -u photoprism photoprism make start
```

The default development user is `photoprism` (UID 1000). If you override the container user, adjust the `-u` option accordingly. Open http://localhost:2342/ and sign in with `admin` / `photoprism`.

After backend changes, stop the service, rebuild, and start it again:

```bash
docker compose exec -T -u photoprism photoprism make stop
docker compose exec -T -u photoprism photoprism make build-go
docker compose exec -T -u photoprism photoprism make start
```

After frontend changes, rebuild the browser assets and refresh the page:

```bash
docker compose exec -T -u photoprism photoprism make build-js
```

For frontend watch mode, run `docker compose exec -u photoprism photoprism make watch-js`. To open an interactive shell, run `make terminal`.

Photo folders belong under `storage/originals/`. Start a scan from **Library → Index** after adding or updating images.

## How to Deploy

### Server Setup (one-time)

1. Create a GitHub Personal Access Token with `read:packages` scope at https://github.com/settings/tokens/new

2. Authenticate Docker on the server:
   ```bash
   echo "<token>" | docker login ghcr.io -u raphaelmatto --password-stdin
   ```

3. Update the compose file — change one line:
   ```yaml
   # Was:
   image: photoprism/photoprism:latest
   # Now:
   image: ghcr.io/raphaelmatto/photoprism:latest
   ```

### Deploy Workflow

After testing changes locally:

```bash
# 1. Merge the finished feature branch into production
git checkout production
git merge feature/<name>
git push origin production

# 2. Wait ~5 minutes for GitHub Actions to build

# 3. On the server
docker compose pull
docker compose up -d
```

If the server keeps reusing an older local copy of `ghcr.io/raphaelmatto/photoprism:latest`, replace just that tag before restarting:

```bash
docker compose -f prod.yml down
docker rmi -f ghcr.io/raphaelmatto/photoprism:latest
docker pull ghcr.io/raphaelmatto/photoprism:latest
docker compose -f prod.yml up -d
```

This is useful when disk pressure or local tag reuse prevents Docker from fetching the newest `latest` image cleanly.

For rollback or pinned deploys, use an immutable image tag instead of `latest`:

```yaml
image: ghcr.io/raphaelmatto/photoprism:sha-<git-sha>
```

### Reverting to Official Image

Change the compose file back:
```yaml
image: photoprism/photoprism:latest
```
Then `docker compose pull && docker compose up -d`.

## How to Sync with Upstream

```bash
git checkout develop
git fetch upstream
git merge upstream/develop
# Resolve any conflicts, then:
git push origin develop
```

When a new upstream change should also land in your deploy branch, merge or cherry-pick it from `develop` into `production` deliberately. Do not treat `develop` as the long-lived place where fork feature work accumulates.

## Round-Tripping Edits to Lightroom

PhotoPrism does not write XMP. When users edit captions/keywords/etc. in the web UI those edits live only in the PhotoPrism database — Lightroom (which reads XMP/IPTC) has no idea they exist. The `photoprism dump-edits` command bridges this gap: it lists every photo whose `*_src` column is `manual` or `batch`, so you can replay those edits in Lightroom and re-index from the resulting XMP.

### Quickstart

```bash
# Run on the server, inside the photoprism container.
docker exec photoprism-photoprism-1 photoprism dump-edits        > edits.md
docker exec photoprism-photoprism-1 photoprism dump-edits --json > edits.json
```

Each photo block in `edits.md` lists the indexed path, the photo UID, the timestamp of the last edit, and every field that has been manually changed. Hierarchical keywords are formatted as `Parent > Child` so they map directly onto Lightroom Classic's keyword hierarchy. Lat/lng pairs render as `lat=…, lng=…`. The `--json` variant emits the same data structured for an LR-side script.

Fields that don't map cleanly to XMP — favorite, private, archive, etc. — are intentionally **not** included; the manifest only surfaces edits that have a Lightroom-side representation.

### Full Workflow

docker exec -it production-photoprism /bin/bash

1. **Dump.** `photoprism dump-edits > edits.md` — read the manifest, apply each block in Lightroom.
2. **Lightroom.** Edit each photo, then **Metadata → Save Metadata to File** (⌘S). For RAW the edits land in a `.xmp` sidecar; for JPEG they're embedded.
3. **Identify touched paths.** In Lightroom Classic, filter by **Metadata Status: Has been changed externally** to list the files whose XMP you just wrote.
4. **Re-index.** Run `photoprism index --overwrite-meta <path>` scoped to the touched paths (or to the whole library if it's small). The flag tells PhotoPrism to overwrite previously-edited values from the file's metadata, which clears the `*_src='manual'` markers as a side effect.
5. **Verify.** Re-run `photoprism dump-edits`. `No manually-edited photos found.` means everything round-tripped cleanly. Anything that still appears is an edit that didn't survive the round-trip — either Lightroom doesn't expose that field, or the value didn't actually make it into the XMP. Useful diagnostic.

### Race-Condition Notes

Between step 1 (dump) and step 4 (re-index) users may make more edits. Those new edits will also have `*_src='manual'` and the manifest from step 1 won't have captured them. Two options:

- **Time-bound the cycle.** Stamp the manifest with the run time; only re-index files whose XMP `ModifyDate` is newer than that stamp. Step 3's "changed externally" filter already does this implicitly.
- **Do it during low-traffic hours.** The site stays up during indexing; it's just slower.

The visible data never changes across the cycle — re-indexing replays the same values, just now sourced from XMP rather than the DB.

## Recommended Server Settings

For retina display support with minimal disk usage:

| Setting | Value | Reason |
|---------|-------|--------|
| Static Size Limit | 1084+ | Pre-generates `tile_1080` for retina cards view |
| Dynamic Size Limit | 720 | No need for large dynamic thumbnails when serving originals |
| Dynamic Previews | Off | All needed sizes are pre-generated |

After changing the Static Size Limit, regenerate thumbnails:
```bash
# Delete existing thumbnail cache
rm -rf /path/to/photoprism/storage/cache/thumbnails

# Rescan to regenerate
# (resize EC2 instance temporarily if needed for CPU)
```

## Local Development

```bash
docker compose up --build
# In another terminal:
make terminal
make fix-permissions   # first time only
make dep
make build-js
make build-go
./photoprism start
```

Frontend hot reload: `make watch-js` (in a separate terminal inside the container).

Go changes require `make build-go` and restarting PhotoPrism.

### TODO
- [ ] Validate retina thumbnails in: Labels, photo preview, edit details, edit labels, edit files, batch edit, account avatar
