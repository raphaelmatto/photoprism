import { describe, expect, it } from "vitest";
import { choosePackedThumbSize, layoutPackedRows, serverMaxFitSize } from "common/packed";

describe("common/packed", () => {
  it("returns empty rows when input is invalid", () => {
    expect(layoutPackedRows([], 1000)).toEqual([]);
    expect(layoutPackedRows(null, 1000)).toEqual([]);
    expect(layoutPackedRows([{ Width: 10, Height: 10 }], 0)).toEqual([]);
  });

  it("keeps justified rows aligned to the measured container width", () => {
    const rows = layoutPackedRows(
      [
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
      ],
      1000,
      { gutter: 10, targetRowHeight: 200 }
    );

    expect(rows).toHaveLength(1);
    expect(rows[0].items.map((item) => item.index)).toEqual([0, 1, 2]);
    expect(rows[0].items.reduce((sum, item) => sum + item.width, 0) + 20).toBe(1000);
  });

  it("does not stretch the last row to fill the container", () => {
    const rows = layoutPackedRows(
      [
        { Width: 400, Height: 200 },
        { Width: 400, Height: 200 },
      ],
      1000,
      { gutter: 10, targetRowHeight: 200 }
    );

    expect(rows).toHaveLength(1);
    expect(rows[0].items.reduce((sum, item) => sum + item.width, 0) + 10).toBeLessThan(1000);
    expect(rows[0].height).toBe(200);
  });

  it("chooses fit thumbnails based on rendered size", () => {
    expect(choosePackedThumbSize(700, 300, false)).toBe("fit_720");
    expect(choosePackedThumbSize(900, 300, false)).toBe("fit_1280");
  });

  it("caps the chosen size to the largest fit_* the server advertises", () => {
    // Server only serves fit_720; a 900px-wide render must fall back to fit_720
    // rather than request fit_1280, which the server would substitute with a
    // square tile_* crop.
    expect(choosePackedThumbSize(900, 300, false, 720)).toBe("fit_720");
    expect(choosePackedThumbSize(1500, 500, false, 1280)).toBe("fit_1280");
    // A zero cap disables capping (legacy callers without the argument).
    expect(choosePackedThumbSize(900, 300, false, 0)).toBe("fit_1280");
  });

  it("derives the server fit-size cap from the config.thumbs list", () => {
    expect(serverMaxFitSize(undefined)).toBe(0);
    expect(serverMaxFitSize([])).toBe(0);
    expect(
      serverMaxFitSize([
        { size: "fit_720", w: 720, h: 720 },
        { size: "fit_1280", w: 1280, h: 1280 },
        { size: "tile_1080", w: 1080, h: 1080 },
      ])
    ).toBe(1280);
    // Only fit_* entries count; tile_* is ignored even if larger.
    expect(
      serverMaxFitSize([
        { size: "fit_720", w: 720, h: 720 },
        { size: "tile_1080", w: 1080, h: 1080 },
      ])
    ).toBe(720);
  });
});
