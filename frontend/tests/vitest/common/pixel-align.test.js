import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PixelAlign } from "common/pixel-align";

describe("common/pixel-align", () => {
  let container, element, position, frames, observer, alignment;

  beforeEach(() => {
    vi.stubGlobal("devicePixelRatio", 2);
    vi.stubGlobal("scrollX", 0);
    vi.stubGlobal("scrollY", 0);
    frames = new Map();
    let frameId = 0;
    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn((callback) => {
        frames.set(++frameId, callback);
        return frameId;
      })
    );
    vi.stubGlobal(
      "cancelAnimationFrame",
      vi.fn((id) => frames.delete(id))
    );
    vi.stubGlobal(
      "ResizeObserver",
      class {
        constructor(callback) {
          this.callback = callback;
          this.observe = vi.fn();
          this.unobserve = vi.fn();
          this.disconnect = vi.fn();
          observer = this;
        }
      }
    );
    container = document.createElement("div");
    element = document.createElement("div");
    element.className = "photo";
    container.append(element);
    position = { x: 100.25, y: 200.171875 };
    element.getBoundingClientRect = () => ({
      left: position.x + (parseFloat(element.style.left) || 0) - window.scrollX,
      top: position.y + (parseFloat(element.style.top) || 0) - window.scrollY,
      width: 476.5,
      height: 650.5,
    });
    alignment = new PixelAlign(container, ".photo");
  });

  afterEach(() => {
    alignment.destroy();
    vi.unstubAllGlobals();
  });

  it("aligns fractional positions to physical pixels without changing image size", () => {
    alignment.align();
    const rect = element.getBoundingClientRect();
    expect(rect.left * 2).toBe(201);
    expect(rect.top * 2).toBe(400);
    expect(rect.width * 2).toBe(953);
    expect(rect.height * 2).toBe(1301);
  });

  it("does not accumulate corrections or change them when scrolling", () => {
    alignment.align();
    const initial = element.style.cssText;
    vi.stubGlobal("scrollY", 47.5);
    for (let i = 0; i < 10; i++) alignment.align();
    expect(element.style.cssText).toBe(initial);
    expect(element.getBoundingClientRect().top * 2).toBe(305);
  });

  it("scales integer source dimensions into the frame, including narrow layouts", () => {
    const image = document.createElement("img");
    image.width = 953;
    image.height = 1301;
    element.append(image);
    alignment.align();
    expect(image.style.transform).toBe("scale(0.5)");
    element.getBoundingClientRect = () => ({ left: 0, top: 0, width: 300, height: 410 });
    alignment.align();
    expect(image.style.transform).toBe(`scale(${300 / 953})`);
  });

  it("batches size changes and realigns after preceding metadata changes height", () => {
    alignment.align();
    position.y += 3.125;
    observer.callback();
    observer.callback();
    expect(frames.size).toBe(1);
    const callback = [...frames.values()][0];
    frames.clear();
    callback();
    expect(element.getBoundingClientRect().top * 2).toBe(407);
  });

  it("tracks appended photos and releases removed elements", () => {
    expect(observer.observe).toHaveBeenCalledWith(container);
    expect(observer.observe).toHaveBeenCalledWith(element);
    const appended = element.cloneNode();
    container.append(appended);
    element.remove();
    alignment.update();
    expect(observer.unobserve).toHaveBeenCalledWith(element);
    expect(observer.observe).toHaveBeenCalledWith(appended);
    expect(alignment.elements.has(element)).toBe(false);
  });

  it("uses the new pixel density after a display or browser zoom change", () => {
    alignment.align();
    vi.stubGlobal("devicePixelRatio", 1.25);
    alignment.align();
    expect(element.getBoundingClientRect().left * 1.25).toBeCloseTo(125);
    expect(element.getBoundingClientRect().top * 1.25).toBeCloseTo(250);
  });

  it("disconnects observers and cancels pending work on teardown", () => {
    alignment.destroy();
    expect(observer.disconnect).toHaveBeenCalled();
    expect(frames.size).toBe(0);
    window.dispatchEvent(new Event("resize"));
    expect(frames.size).toBe(0);
  });
});
