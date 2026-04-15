# PhotoPrism Local Preview Notes

**Last Updated:** April 15, 2026

## Purpose

This note captures the local startup and image-loading workflow for this fork so it does not need to be rediscovered later.

## Start the Development Stack

From the repo root:

```bash
docker compose up -d mariadb dummy-webdav dummy-oidc photoprism
```

Important:

- The development `photoprism` container starts idle by default.
- `docker compose up` alone does **not** start the PhotoPrism web server.

## Start the PhotoPrism Web Server

After the container is up, start the app inside the running container:

```bash
docker compose exec photoprism ./photoprism start
```

Then open:

```text
http://localhost:2342/library/login
```

Default local login:

- Username: `admin`
- Password: `photoprism`

## Check Whether the Server Is Running

If the page does not load, check:

```bash
docker compose exec photoprism ./photoprism status
docker compose ps
docker compose logs --tail=200 photoprism
```

## Rebuild After Code Changes

For normal iteration:

```bash
docker compose exec photoprism make build-js
docker compose exec photoprism make build-go
docker compose exec photoprism ./photoprism stop
docker compose exec photoprism ./photoprism start
```

For frontend-only iteration, a watcher is faster:

```bash
docker compose exec photoprism make watch-js
```

## Stop the Server or Stack

Stop only the app process:

```bash
docker compose exec photoprism ./photoprism stop
```

Stop the whole stack:

```bash
docker compose down
```

## Where Images Come From in This Dev Setup

In the current `compose.yaml`, the repo is bind-mounted into the container and PhotoPrism uses:

- Originals: `./storage/originals`
- Import: `./storage/import`

If those folders are empty, PhotoPrism will show no images.

## Fastest Way to Load Sample Images

Copy a representative set of files into:

```bash
storage/originals/
```

Recommended sample mix:

- Portrait photos
- Landscape photos
- Panoramas
- A few videos or live photos if relevant

Then index them:

```bash
docker compose exec photoprism ./photoprism index
```

If you changed or replaced files and want a full rescan:

```bash
docker compose exec photoprism ./photoprism index --force
```

## Import Workflow

If you want to use the import path instead of copying directly into originals:

1. Put files under:

```bash
storage/import/
```

2. Move them into originals through PhotoPrism:

```bash
docker compose exec photoprism ./photoprism import
```

3. Re-index if needed:

```bash
docker compose exec photoprism ./photoprism index
```

## Using Images From Another Path on Disk

If you already have image folders elsewhere on the host, there are two reasonable options:

### Option 1: Copy a Small Working Set

Best for feature work:

- Copy a small subset into `storage/originals/`
- Run `photoprism index`

This keeps the dev environment simple and reproducible.

### Option 2: Mount an External Folder

Best for a larger existing library:

- Temporarily change the `photoprism` service volumes in `compose.yaml`
- Mount your host image directory into the container
- Point `PHOTOPRISM_ORIGINALS_PATH` at that mounted path

If you do this, document the local-only change or keep it out of committed config.

## Playwright Preview Workflow

Once the app is reachable at `http://localhost:2342/`, Playwright can be used to verify UI changes locally before building or publishing any Docker image.

Typical target URL:

```text
http://localhost:2342/library/login
```

## Related Docs

- [IMAGE_PACKING_SPEC.md](./IMAGE_PACKING_SPEC.md)
- [README.md](./README.md)
