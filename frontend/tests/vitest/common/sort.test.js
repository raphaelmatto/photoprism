import { describe, expect, it } from "vitest";
import { fromWireOrder, toWireOrder } from "common/sort";

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
});
