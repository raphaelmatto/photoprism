# PhotoPrism Fork — raphaelmatto

Personal fork of [PhotoPrism](https://github.com/photoprism/photoprism) with display quality enhancements for retina/HiDPI screens.

## Features Added

### Display Settings (Settings > Content > Display)

Three new settings that work independently of each other:

- **Original Images** — Serve original image files in the lightbox viewer instead of generated thumbnails. Preserves per-area sharpening and full resolution from tools like Photoshop.
- **Retina Lightbox** — Scale lightbox images by `1/devicePixelRatio` for pixel-perfect quality on HiDPI displays (e.g. Apple Retina). Works with both originals and thumbnails.
- **Retina Thumbnails** — Use higher-resolution tiles in grid views. Mosaic and list views use `tile_500` instead of `tile_224`, cards view uses `tile_1080` instead of `tile_500`.

All settings take effect immediately without a server restart. They are stored in `settings.yml` under the `display` key.

### ARM64 Docker Build

Production images now build from the standard `docker/photoprism/questing/Dockerfile` on a native ARM64 GitHub Actions runner. An earlier custom ARM64 Dockerfile existed to work around broken TensorFlow headers in the deprecated `photoprism/develop:bookworm` base image, but that workaround was retired after switching to the supported `questing` base. See [upstream issue #5444](https://github.com/photoprism/photoprism/issues/5444).

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
- **Dockerfile**: `docker/photoprism/questing/Dockerfile`
- **Build time**: ~5 minutes

The workflow uses `GITHUB_TOKEN` for registry auth, publishes both `latest` and commit-specific `sha-...` tags, and can also be triggered manually with `workflow_dispatch`.

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

## Dev workflow

To start the containers:
docker compose up -d photoprism mariadb traefik dummy-webdav dummy-oidc

To log into the photoprism container:
docker exec -it photoprism-photoprism-1 /bin/bash

To start photoprism, once logged in:
./photoprism start

To view:
http://localhost:2342/

After changes run:
make build-js

... or try this for a hot-reload:
make watch-js

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
