# TODO

Fork-specific work queue. Each entry is self-contained so an AI agent can
pick it up cold. Keep entries small and actionable; delete when done.

---

## Fix `fit_*` fallback to `tile_*` in thumbnail handler

**Context.** When the client requests a fit_* thumbnail size that exceeds
the server's configured `ThumbSize` (Static Size Limit) and
`ThumbUncached` (Dynamic Previews) is disabled, the API currently serves
the nearest cached thumbnail regardless of type. In practice that means
`GET /api/v1/t/<hash>/<token>/fit_1280` can return the pre-generated
`tile_1080` (1080×1080 center-crop square) instead of the largest
cached `fit_*`. Square crops replace correct-aspect previews — heads get
cut off in packed/cards views whenever a user keeps Size Limits low.

**Reproduce.**
1. Set Advanced → Static Size Limit = 1084, Dynamic Previews = off.
2. Run `photoprism thumbs --force` so only `fit_720` and `tile_1080`
   get pre-generated for each photo.
3. Request `fit_1280` directly:
   `curl -s /api/v1/t/<hash>/<token>/fit_1280 | exiftool -`. Actual
   dimensions come back 1080×1080 (square), not the expected 1280×≈.

**Desired behavior.** If the requested `fit_N` isn't available, fall
back to the **largest cached fit_\* that is ≤ N**, never to a `tile_*`.
A `fit_*` request must never return a square.

**Where to look.**
- `internal/api/` — handler registered for `/t/:hash/:token/:size`.
  Grep for `thumb.ParseSize` or the `Thumbnail` handler.
- `internal/thumb/sizes.go` — `Size.Fit` bool distinguishes fit vs tile.
- `internal/thumb/filenames.go` / `internal/thumb/file.go` — where the
  file path for a cached thumb is resolved; likely the fallback decision
  lives here or just above it in the handler.

**Constraints.**
- Don't regenerate on-demand when `ThumbUncached` is false — that would
  defeat the user's explicit setting.
- Keep the fix minimal so it can be sent upstream as a PR.
- Add a test under `internal/thumb/` that covers "request fit_N with
  only fit_M (M<N) and tile_K cached → returns fit_M".

**Related frontend mitigation.** The frontend already caps requested
packed sizes to `ThumbSize` so this bug stops biting day-to-day users;
this backend fix is the proper upstream-worthy solution.
