# Upstream Feature and Integration Assessment

**Last Updated:** October 29, 2026

Fork maintenance record. Keep this document out of upstream contribution branches.

## Decision

On October 29, 2026, the user reviewed the findings and confirmed that none of the investigated upstream additions would be useful for their workflow. Retain the current fork. No upstream feature imports or broad merge are planned. The candidate fixes and estimates below preserve the investigation for future reference; they are not an implementation queue.

The workflow remains external metadata editing, image replacement, and folder scans. The PhotoPrism database can be rebuilt from the image folders, so database recovery-point work is not a prerequisite for development. Existing customizations have been used successfully in production; verification should focus on changed behavior rather than a broad regression pass as routine preparation.

## Assessment Context

The local sample contains 218 JPEGs and no XMP sidecars, PDFs, HEIC files, or videos. This describes the local sample only, not the production library. Sidecar, document, video, and specialist-format features may have little immediate value for that sample.

The Ubuntu 26.04 LTS base update already refreshes the container tools. That benefit should not be counted again as a reason to merge upstream application code.

## Scope and Snapshot

Reviewed on October 28, 2026, with upstream refs fetched during the review. Commits dated October 29 in UTC can still fall on October 28 in America/New_York.

| Target | Commit | Difference From This Fork |
| --- | --- | --- |
| Local fork, including the base and CI work | `0bcbfbafd` | Baseline for this assessment |
| Latest published release, October 7 | [`261007-65faaae5d`](https://github.com/photoprism/photoprism/releases/tag/261007-65faaae5d) | 2,596 upstream commits absent locally |
| Current upstream `release` branch | `7a347f125` | 3,176 upstream commits absent locally; includes work after the published release |
| Current upstream `develop` branch | `7f4f529f4` | 3,181 upstream commits absent locally |

Commit counts include tests, translations, documentation, dependencies, and merge history; they are not counts of missing features. A branch named `release` is not equivalent to the last published release tag.

## Measured Merge Difficulty

Trial merges used `git merge-tree --write-tree`, without applying the result to the working tree:

- Latest published release: **43 conflicting files**—31 application files, four tests, and eight build/documentation files.
- Current development: **47 conflicting files**.
- Of the fork's 160 changed files, upstream has also changed 101 by the published release, and 102 by the current release branch.

The principal conflict areas are the lightbox, photo views and toolbar, navigation, settings, metadata parsing, search, thumbnails, and migration registration. These are central to this fork's custom behavior.

Conflict counts understate the work. Some changes merge as text while requiring an application decision or an API adaptation. One concrete example is the upstream relocation of `component/sidebar/info.vue` into a redesigned lightbox sidebar: preserving the old file does not integrate the fork's configurable metadata rows into the new sidebar.

Your ability to rebuild the database from images simplifies a future upgrade. It does not remove the need to preserve the fork's code and UI behavior.

## Features Missing From the Published Release

Effort estimates below are rough engineering estimates for a targeted port plus focused verification. They are not implementation results or fixed commitments.

| Upstream Addition | Value for This Workflow | Interaction With the Fork | Targeted Integration Effort |
| --- | --- | --- | --- |
| Broader XMP metadata support, including GPS, camera/lens fields, and face regions | Useful if external edits are supplied in sidecars; less important when replacing JPEGs with embedded metadata | Touches the same parser and index paths as the fork's capture-date priority, hierarchical keywords, and metadata overwrite behavior | Medium to large: about 1–3 days for a chosen subset; avoid importing the entire parser refactor blindly |
| Re-reading modified XMP sidecars during an incremental scan | Useful if only a sidecar changes while the image file stays unchanged | The selected commit's code patch checks cleanly; its README hunk conflicts. It also refers to format/config helpers introduced upstream, so dependencies must be checked | Medium: roughly half a day to two days, depending on prerequisites |
| Editable full-screen info sidebar, album/label controls, and manual face markers | Limited value if all metadata edits remain external | Direct overlap with configurable metadata rows, searchable metadata fields, lightbox theme, originals/Retina behavior, and toolbar layout | Large: several days; this is a UI integration project |
| Interactive 360° photos and videos, including later Insta360 support; higher preview limits | Useful for a 360° library, not ordinary JPEG browsing | Substantial lightbox changes, viewer dependencies, media classification, preview sizes, and interaction handling overlap the fork's viewer | Large: several days; preserve original-image and Retina paths explicitly |
| Inline multi-page PDF viewer | Useful if PhotoPrism should also serve documents | A new viewer component exists, but it hooks into the heavily modified lightbox, styles, controls, keyboard handling, and download behavior | Medium to large: roughly 1–3 days for a scoped port |
| Improved layered TIFF/PSD previews; native HEIC/AVIF and JPEG XL conversion | Useful if these formats are added to the library | Layered TIFF/PSD patch checks cleanly. Native conversion intersects thumbnail code and needs upstream helper files absent from this fork | Small to medium for layered TIFF/PSD; medium for native conversion, roughly half a day to two days per selected change |
| Drag-and-drop web uploads | A convenience if uploads happen through the browser | The initial implementation patch checks cleanly against this fork | Small: a few hours to a day, including the upload tests |
| Accessibility settings, including reduced motion and disabling hover-open menus; Collections settings tab | Optional UI improvements | Fork settings and navigation are customized; individual accessibility options can be ported without replacing the settings layout | Small to medium: roughly half a day to a day for a specific option |
| New face models, improved clustering, private/hidden people, ONNX classification and NSFW models, optional NVIDIA inference | Useful if these AI features are wanted | The AI pipeline and settings changed substantially. The fork's Add AI Keywords control is absent upstream and must remain connected to any replacement pipeline | Large: several days; no obvious benefit if these features are unused |
| Label-filter AND/NOT operators | Useful for built-in label filtering | The fork already has its own broader free-text phrase/Boolean parser. These are overlapping search improvements, not an interchangeable implementation | Medium for a selected addition; preserve the fork's existing query semantics |

Sources: [May release](https://github.com/photoprism/photoprism/releases/tag/260523-0544f71c1), [July release](https://github.com/photoprism/photoprism/releases/tag/260728-bbde8f452), [September release](https://github.com/photoprism/photoprism/releases/tag/260919-28c46a116), [October release](https://github.com/photoprism/photoprism/releases/tag/261007-65faaae5d). Conflict and effort assessments come from local code comparison and patch checks.

## Changes After the Published Release

### Library Access Through MCP and Expanded OAuth

The published release has MCP reference tools for configuration and search-filter documentation. Later October work adds authenticated library tools such as `search_media`, `get_media`, collection listing, and people/label listing, together with expanded OAuth app connections.

That could be useful if an AI client should search and inspect this library. It is not part of the October 7 published release. Integration would reach authentication, sessions, ACL, search results, and dependencies, so treat it as a larger feature port rather than a small endpoint addition. Current source also distinguishes authenticated library sessions from public/reference sessions.

Sources: [MCP package documentation](https://github.com/photoprism/photoprism/blob/7a347f125bb93a44cfeff7cfd29dbad5337ea665/internal/mcp/README.md), [library search tool](https://github.com/photoprism/photoprism/commit/5092e1bac), [OAuth app scopes](https://github.com/photoprism/photoprism/commit/db844d807).

### GORM v2 and Experimental PostgreSQL Support

The GORM migration is after the October 7 release and is present in the current upstream branches. Its principal migration commit changes 296 files. This is an additional reason to distinguish a published-release update from following current development.

The existing MariaDB setup has no immediate need for PostgreSQL. A move to current development would still require checking the fork's custom search SQL, migrations, and database helpers against the new ORM and SQL behavior.

Sources: [GORM migration](https://github.com/photoprism/photoprism/commit/64b820cca), [experimental PostgreSQL documentation change](https://github.com/photoprism/photoprism/commit/cd522cfa9).

## Specific Backport Checks

These are `git apply --check` results against local `production`. A clean patch check confirms mechanical applicability, not compilation, complete dependencies, or correct behavior. No candidate was applied.

| Candidate | Check Result | Assessment |
| --- | --- | --- |
| [`e39f9f20a`: Recover after a primary file is deleted](https://github.com/photoprism/photoprism/commit/e39f9f20a) | Clean | A strong small candidate for a replace/re-scan workflow; two files including tests |
| [`d8959859a`: Match folder albums by exact path](https://github.com/photoprism/photoprism/commit/d8959859a) | Clean | Small candidate if folder albums show unintended pictures. Exact matching can change descendant-folder behavior, so confirm the desired semantics |
| [`ba95e3a72`: Drag-and-drop upload](https://github.com/photoprism/photoprism/commit/ba95e3a72) | Clean | Small optional feature; includes an extensive upload test file |
| [`20fd49f2a`: Layered TIFF/PSD previews](https://github.com/photoprism/photoprism/commit/20fd49f2a) | Clean | Useful format-specific addition if these files are used |
| [`290d1a13e`: Profile endpoint authorization checks](https://github.com/photoprism/photoprism/commit/290d1a13e) | Test hunk conflicts | Production-code hunks check cleanly; port and run the focused authorization tests rather than discarding them |
| [`ec6b83412`: Incremental XMP sidecar re-read](https://github.com/photoprism/photoprism/commit/ec6b83412) | README hunk conflicts | Code hunks check cleanly, but format/config prerequisites and follow-up fixes still need evaluation |
| [`568432df8`: Native HEIC/AVIF conversion](https://github.com/photoprism/photoprism/commit/568432df8) | Conflicts and missing helper file | Bring the conversion prerequisites and reconcile tests; not a single-commit drop-in |
| [`15885d8c9`: Retry PNG export without a rejected ICC profile](https://github.com/photoprism/photoprism/commit/15885d8c9) | Thumbnail hunk conflicts | Reconcile with the fork's atomic thumbnail writes; relevant only to affected inputs |
| [`e2aef7cdc`: Decode ExifTool keyword arrays](https://github.com/photoprism/photoprism/commit/e2aef7cdc) | Metadata hunk conflicts | The fork already decodes keyword-array elements and preserves multi-word keywords. Review any remaining string-slice case separately; do not replace the custom Keywords branch with generic word splitting |

## What Exact Folder Matching Means

This change affects which pictures a folder album returns. It makes the saved folder path a literal match rather than interpreting the path as a general search pattern.

The current search code uses SQL `LIKE` for a path without a trailing slash. It also supports search operators: `*` is converted to `%`, and `|` separates alternatives. Consequently, characters that are legitimate in a filesystem folder name can have a different meaning in its saved search filter.

For example:

```text
Trips_2026/photo-a.jpg
Trips-2026/photo-b.jpg
```

In SQL `LIKE`, `_` matches any one character. A folder-album filter for `Trips_2026` can therefore also match `Trips-2026`, returning photo-b when only photo-a belongs to the requested folder.

Upstream commit `d8959859a` takes the actual folder album's stored path and selects the existing exact-path search mode by appending a trailing slash. The search code removes that slash and compares the path with `=`. Underscores, percent signs, asterisks, and pipes are then literal folder-name characters. The upstream tests cover those characters, case differences, and a folder literally named `null`.

For ordinary names without special characters, the existing `LIKE` filter does not automatically add wildcards, so the result will often be unchanged. This is chiefly a correctness fix for folder-name interpretation, not a new recursive-browsing feature. The exact mode selects pictures directly in that folder; it does not include descendants. The existing non-wildcard path mode can already exclude descendants, so this patch is not a general recursion toggle. It does not change folder sorting or move any files.

The patch checked cleanly against the fork but has not been applied. The earlier assessment's note about descendant behavior was a caution to inspect the saved filter, not evidence that ordinary folder names currently perform a recursive search.

Sources: `internal/entity/search/photos.go` (path filter), `internal/entity/search/conditions.go` (`OrLike`), and [the upstream implementation and tests](https://github.com/photoprism/photoprism/commit/d8959859a0e37e89089396bbf680ec6607a1ae3d).

## Custom Behavior That Must Survive an Update

- **Capture dates:** The fork explicitly prefers a still-image capture timestamp over a different creation timestamp and has a Lightroom CreateDate fallback. Upstream extracted `ResolveTimeZone`, whose initial CreatedAt handling still differs. Port these rules and their regression tests into the new resolver if adopting that refactor.
- **Keyword preservation:** Keep hierarchical and multi-word keywords, case-preserving handling, the keyword page/tree, and the Add AI Keywords setting. Upstream's superficially similar fixes do not replace all of this behavior.
- **Viewer and browsing:** Preserve originals, Retina lightbox/thumbnails, packing, Scroll view, Light Table theme, border controls, and configurable/searchable metadata. The corresponding Display settings are fork additions, not upstream equivalents.
- **Navigation and sorting:** Keep the fork's folder tree, navigation visibility and expansion behavior, album ordering, capture-date folder defaults, and reverse-sort state. Upstream's Collections/settings changes and folder fixes need to be reconciled with them.
- **Thumbnail integrity:** Keep atomic writes while evaluating newer conversion, sizing, orientation, and ICC handling.
- **Fork operations:** Retain the dated LTS bases, production-only publishing, CI gates, and private/fork PR-safety rules when resolving build/documentation conflicts.

## Practical Integration Choices

These alternatives were evaluated during the investigation. The user selected staying on the current fork without importing the reviewed additions.

1. **Stay on the current fork:** Current decision. Small fixes can be considered if a concrete issue is encountered; no candidate is scheduled by this document.
2. **Port one desired feature:** Reasonable if PDF viewing, 360° media, sidecar-only updates, or AI-client library search becomes useful. Estimate and implement that feature independently.
3. **Update to the published October release:** High effort, plausibly several days to a week or more, because of the viewer, settings, metadata, and thumbnail conflicts. Requires focused checks of the customizations affected by the update.
4. **Follow current upstream development:** Higher uncertainty and additional database/authentication adaptation beyond the published release. There is no current workflow requirement that justifies it by itself.

## CI Work Completed Before This Review

Local production includes `d0208b763` (test/lint helpers) and `0bcbfbafd` (workflow gating). Validation passed for 649 frontend tests, 49 focused backend packages, actionlint/shellcheck, and positive/negative lint-gate checks. The broader backend suite and full existing lint backlog are not represented as passing. A hosted Actions run remains pending a push.

The base update and two CI commits remain local. This assessment did not merge upstream code or deploy a production image.
