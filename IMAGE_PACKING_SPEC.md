# PhotoPrism Image Packing Spec

**Last Updated:** April 15, 2026

## Goal

Add a global display setting that enables a justified, Flickr-style packed image layout across PhotoPrism preview surfaces so users can see the full image area without preview-time cropping.

The primary user need is search and recognition during browsing. Users often scan preview pages to find people or details in photos, and cropped previews can hide the content they are looking for.

## Product Summary

When Image Packing is enabled, PhotoPrism should render image previews in justified rows that preserve each image's full visible area while maintaining a dense, visually stable layout.

This feature should:

- Preserve the full image area in preview grids.
- Apply across all PhotoPrism preview surfaces, not just one page.
- Remain enabled on mobile.
- Reflow on viewport resize.
- Avoid visible layout churn while scrolling.
- Maintain retina-quality previews.
- Treat other media types the same way as images for layout purposes.

## User Decisions Captured

- Layout style: justified, Flickr-style packing.
- Scope: all views.
- Control model: global display setting.
- Cropping: disallowed for the packed layout.
- Mobile behavior: enabled on mobile as well as desktop.
- Media handling: no special treatment for videos, live photos, RAW, vectors, or documents.
- Primary goal: prevent users from missing important image content in previews.
- Resize behavior: reflow is acceptable on resize, but not during scrolling.
- Preview quality: packed layout must retain retina-quality previews.

## Non-Goals

- Replacing the lightbox viewer.
- Changing the existing metadata sidebars, selection model, or edit flows.
- Introducing per-page or per-session packing toggles.
- Designing media-type-specific packing rules in the first version.
- Solving every thumbnail-optimization problem in the first version.

## Proposed User Experience

### Display Setting

Add a new global display setting under `Settings > Content > Display`:

- `Image Packing`
  - Label intent: enable packed previews that preserve the full image area.
  - Suggested hint: `Arrange previews in justified rows so the full image stays visible without cropping.`

This setting should persist in `settings.yml` under the existing `display` group.

### Layout Behavior

When `Image Packing` is off:

- Existing square-tile behavior remains unchanged.

When `Image Packing` is on:

- Preview rows become justified rows with variable item widths.
- Images preserve aspect ratio.
- Preview rows share a common target row height within a given container width.
- The row is expanded or contracted so items fill the available width with minimal gaps.
- No preview should crop the underlying image content.
- Portrait images should remain fully visible rather than being center-cropped into squares.

### Stability Expectations

- Layout may recompute when the viewport width changes.
- Layout must not reshuffle already rendered content merely because the user scrolls.
- Infinite scrolling or pagination must append new packed rows deterministically.
- The same input order must produce the same layout for the same viewport width and spacing settings.

### Mobile Expectations

- Packed layout stays enabled on mobile.
- Row heights may be smaller on narrow screens, but the layout should still preserve full image visibility.
- Touch targets, selection affordances, and action buttons must remain usable.

## Surfaces in Scope

The first implementation target is all preview surfaces that currently rely on cropped square or near-square thumbnails, including at least:

- Main photo mosaic view.
- Cards view.
- List view previews.
- Albums page previews.
- Labels page previews.
- Library browse previews.
- Photo preview and edit surfaces where thumbnail previews are shown.
- Other existing pages already updated for `Retina Thumbnails`.

If a surface cannot adopt packed rows without disproportionate complexity, it should be documented explicitly as a deferred exception rather than silently omitted.

## Functional Requirements

### Settings

- Add `display.imagePacking` to the frontend settings model.
- Add `ImagePacking` to backend display settings.
- Expose the setting through the existing settings API and persistence flow.
- Default to `false` for backward compatibility.

### Layout Engine

- Use a deterministic justified-layout algorithm based on each asset's intrinsic width and height.
- Preserve source order.
- Compute row breaks from known dimensions and container width.
- Use consistent gutter spacing between items.
- Avoid runtime measurement loops that cause visible jumping while scrolling.
- Support incremental appends when more results are loaded.

### Asset Dimensions

- Use known media dimensions from the model whenever available.
- If dimensions are missing, fall back to a conservative placeholder aspect ratio.
- Once dimensions are known, layout should stabilize and remain deterministic.

### Preview Rendering

- Render previews using non-cropping image presentation.
- Prefer CSS/object-fit behavior or background rendering that preserves the full frame.
- Selection, favorite, open, and context-menu actions must continue to work.
- Live photo and video indicators remain visible and functional.

### Retina Quality

- Packed previews must remain sharp on high-DPI displays.
- `Retina Thumbnails` and `Image Packing` should work together rather than competing.
- The implementation should request thumbnails large enough for the rendered box size multiplied by device pixel ratio when practical.

## Technical Direction

### Frontend Architecture

The current fork already centralizes related display behavior through the `display` settings group. Image Packing should follow the same pattern:

- Add a new setting in the backend display settings struct.
- Add a frontend default in the settings model.
- Add a checkbox in the Display settings page.
- Introduce a shared packed-layout utility rather than duplicating layout math per view.

The preferred implementation is a reusable packed-row layout helper used by all affected preview components.

### Layout Computation Model

For a sequence of assets with aspect ratios:

1. Choose a target row height for the active viewport and surface.
2. Accumulate items in order until their scaled widths approximately fill the available row width including gutters.
3. Solve the row height that exactly fits the container width.
4. Render each item with its computed width and height.
5. Repeat for remaining items.

This should be done in JavaScript from model metadata, not by waiting for browser image decode events.

### Scroll Stability

To avoid reflow during scrolling:

- Compute row layout from the loaded result set and current container width.
- Recompute only when:
  - container width changes,
  - display settings change,
  - result set changes,
  - more results are appended.
- Do not base row membership on intersection-observer visibility state.

Virtualization may still be used, but it must not alter row composition.

### Thumbnail Strategy

The current code switches among fixed thumbnail sizes such as `tile_224`, `tile_500`, and `tile_1080`. Packed layout breaks the assumption that all preview cells are square.

Version 1 should therefore:

- Keep using PhotoPrism thumbnail endpoints.
- Prefer fit-style, non-cropped thumbnails where available for packed previews.
- Request a size that is large enough for the rendered item dimensions at retina density.
- Fall back safely when an exact fit-size is unavailable.

An acceptable first implementation is a heuristic size selection strategy based on:

- rendered width,
- rendered height,
- device pixel ratio,
- current `Retina Thumbnails` setting.

This area should be isolated behind a helper so thumbnail selection can be refined later without rewriting view components.

## Data and API Considerations

- No new backend database schema is required.
- No new public API endpoints are required.
- Existing thumbnail and viewer endpoints should remain compatible.
- If packed previews require additional fit-size support or better non-cropping thumb selection, those changes should be additive.

## Performance Requirements

- Packing calculations must be fast enough for large result sets.
- Layout computation should avoid forced synchronous reflow.
- Re-rendering should be bounded to meaningful changes.
- The packed layout should not trigger expensive image churn while scrolling.

If needed, the first version may:

- compute layout per page of loaded results,
- debounce resize recomputation,
- use simple heuristics before introducing more complex optimization.

## Accessibility and Interaction Requirements

- Keyboard navigation and focus behavior must remain intact.
- Existing open/select/favorite actions must remain reachable.
- Preview overlays and badges must remain legible against variable aspect ratios.
- Hover- and touch-based affordances must not depend on square cells.

## Acceptance Criteria

Image Packing is complete when all of the following are true:

- A global `Image Packing` display setting exists and persists.
- Enabling it changes all in-scope preview surfaces to a justified packed layout.
- Packed previews preserve the full visible image area without cropping.
- The same result ordering stays visually stable while scrolling.
- Resizing the browser recomputes layout cleanly.
- Mobile layouts remain usable with packing enabled.
- Retina displays show visibly sharp previews when `Retina Thumbnails` is enabled.
- Existing interactions still work across packed previews.

## Testing Strategy

### Manual Verification

- Desktop preview at `localhost:2342`.
- Mobile preview with a narrow viewport.
- Compare `Image Packing` on versus off.
- Compare `Retina Thumbnails` on versus off while packing is enabled.
- Verify mixed galleries containing landscape, portrait, panorama, square, video, and live-photo assets.
- Verify albums, labels, browse, list, cards, and mosaic surfaces.

### Automated Coverage

- Unit-test the packed-layout helper with deterministic input ratios and container widths.
- Add frontend tests covering:
  - row generation,
  - stable ordering,
  - resize recomputation,
  - thumbnail-size selection heuristics.
- Use Playwright or existing browser-based acceptance coverage for representative pages once the local app preview is stable.

## Risks

- The current preview system is heavily oriented around fixed square tile sizes.
- Some surfaces may rely on CSS assumptions that break once widths and heights become variable.
- Packed layouts can look poor if thumbnail resolution is under-requested.
- Virtualized views may need refactoring to preserve row stability.

## Recommended Rollout

Implement in this order:

1. Add the setting and shared layout helper.
2. Convert the main mosaic view.
3. Extend to cards and list preview surfaces.
4. Extend the remaining browse, albums, labels, and edit/preview surfaces.
5. Refine retina thumbnail selection after the layout is visually correct.

## Open Implementation Questions

These are engineering questions, not product blockers:

- Which existing `fit_*` sizes are sufficient for packed previews, and where do we need smarter selection?
- Do any in-scope surfaces require a minimum or maximum row height different from the others?
- Can existing virtualization be retained, or should packed rows own their own rendering window?
- Should list view adopt full justified mini-previews or a more constrained packed-preview column while retaining table semantics?
