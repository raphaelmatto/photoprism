import { describe, expect, it } from "vitest";
import { fromWireOrder, resolveReverse, toWireOrder } from "common/sort";

describe("common/sort", () => {
  it("maps directionless date to newest (reverse flag carries direction)", () => {
    expect(toWireOrder("date")).toBe("newest");
  });

  it("passes non-date orders through unchanged", () => {
    expect(toWireOrder("name")).toBe("name");
    expect(toWireOrder("added")).toBe("added");
    expect(toWireOrder("relevance")).toBe("relevance");
  });

  it("migrates legacy newest/oldest into date + explicit reverse state", () => {
    expect(fromWireOrder("newest")).toEqual({ order: "date", reverse: false });
    expect(fromWireOrder("oldest")).toEqual({ order: "date", reverse: true });
  });

  it("returns unknown orders untouched and leaves reverse to the caller", () => {
    expect(fromWireOrder("name")).toEqual({ order: "name", reverse: null });
    expect(fromWireOrder("added")).toEqual({ order: "added", reverse: null });
  });

  it("resolves reverse direction using explicit values before defaults", () => {
    expect(resolveReverse({ queryReverse: "false", queryOrder: "oldest", storedReverse: "true", defaultOrder: "oldest" })).toBe(false);
    expect(resolveReverse({ queryOrder: "oldest", storedReverse: "false", defaultOrder: "newest" })).toBe(true);
    expect(resolveReverse({ storedReverse: "false", defaultOrder: "oldest" })).toBe(false);
  });

  it("uses the model date direction when no user preference exists", () => {
    expect(resolveReverse({ defaultOrder: "oldest" })).toBe(true);
    expect(resolveReverse({ defaultOrder: "newest" })).toBe(false);
    expect(resolveReverse({ defaultOrder: "name" })).toBe(false);
  });
});
