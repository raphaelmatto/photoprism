const PackedThumbSizes = [
  { size: "fit_720", max: 720 },
  { size: "fit_1280", max: 1280 },
  { size: "fit_1920", max: 1920 },
  { size: "fit_2560", max: 2560 },
  { size: "fit_4096", max: 4096 },
  { size: "fit_5120", max: 5120 },
  { size: "fit_7680", max: 7680 },
];

// layoutPackedRows computes deterministic justified rows for the provided items.
export function layoutPackedRows(items, containerWidth, options = {}) {
  const targetRowHeight = options.targetRowHeight ?? 220;
  const gutter = options.gutter ?? 6;
  const minAspectRatio = options.minAspectRatio ?? 0.2;
  const maxAspectRatio = options.maxAspectRatio ?? 5;

  if (!Array.isArray(items) || items.length === 0 || containerWidth <= 0) {
    return [];
  }

  const rows = [];
  let row = [];
  let aspectRatioSum = 0;

  for (const [index, item] of items.entries()) {
    const aspectRatio = clampAspectRatio(item);
    row.push({ item, index, aspectRatio });
    aspectRatioSum += aspectRatio;

    const gapWidth = Math.max(row.length - 1, 0) * gutter;
    const projectedWidth = aspectRatioSum * targetRowHeight + gapWidth;

    if (projectedWidth >= containerWidth) {
      rows.push(buildRow(row, aspectRatioSum, containerWidth, gutter));
      row = [];
      aspectRatioSum = 0;
    }
  }

  if (row.length > 0) {
    rows.push(buildRow(row, aspectRatioSum, containerWidth, gutter, targetRowHeight));
  }

  return rows;

  function clampAspectRatio(item) {
    const width = Number.parseInt(item?.Width, 10) || 1;
    const height = Number.parseInt(item?.Height, 10) || 1;
    return Math.min(maxAspectRatio, Math.max(minAspectRatio, width / height));
  }
}

// choosePackedThumbSize returns the smallest fit thumbnail that should remain sharp at the rendered size.
// maxServerSize caps the request to what the server will actually serve as a fit_* thumbnail;
// larger sizes fall back to the previous candidate so the server never substitutes a tile_* crop.
export function choosePackedThumbSize(width, height, retinaThumbnails, maxServerSize = 0) {
  const devicePixelRatio = retinaThumbnails ? Math.max(window.devicePixelRatio || 1, 1) : 1;
  const targetSize = Math.ceil(Math.max(width, height) * devicePixelRatio);

  let fallback = PackedThumbSizes[0].size;
  for (const candidate of PackedThumbSizes) {
    if (maxServerSize > 0 && candidate.max > maxServerSize) {
      return fallback;
    }
    if (candidate.max >= targetSize) {
      return candidate.size;
    }
    fallback = candidate.size;
  }

  return "fit_7680";
}

// serverMaxFitSize returns the largest fit_* pixel size the server advertises
// via the config.thumbs list. Returns 0 when the list is empty or missing,
// which disables the cap in choosePackedThumbSize.
export function serverMaxFitSize(thumbs) {
  if (!Array.isArray(thumbs) || thumbs.length === 0) {
    return 0;
  }
  let max = 0;
  for (const t of thumbs) {
    if (typeof t?.size !== "string" || !t.size.startsWith("fit_")) {
      continue;
    }
    const w = Number.parseInt(t?.w, 10) || 0;
    if (w > max) {
      max = w;
    }
  }
  return max;
}

function buildRow(row, aspectRatioSum, containerWidth, gutter, maxHeight = 0) {
  const gapWidth = Math.max(row.length - 1, 0) * gutter;
  const availableWidth = Math.max(containerWidth - gapWidth, row.length);
  let rowHeight = Math.max(80, Math.floor(availableWidth / aspectRatioSum));

  if (maxHeight > 0) {
    rowHeight = Math.min(rowHeight, maxHeight);
  }

  const items = row.map(({ item, index, aspectRatio }) => {
    return {
      item,
      index,
      width: Math.max(1, Math.round(rowHeight * aspectRatio)),
      height: rowHeight,
    };
  });

  const totalWidth = items.reduce((sum, rowItem) => sum + rowItem.width, 0);
  const widthDelta = availableWidth - totalWidth;

  if (widthDelta > 0 && maxHeight === 0 && items.length > 0) {
    items[items.length - 1].width += widthDelta;
  } else if (widthDelta < 0) {
    let overflow = Math.abs(widthDelta);

    for (let i = items.length - 1; i >= 0 && overflow > 0; i -= 1) {
      const shrink = Math.min(overflow, items[i].width - 1);
      items[i].width -= shrink;
      overflow -= shrink;
    }
  }

  return {
    height: rowHeight,
    items,
  };
}
