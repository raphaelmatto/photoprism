// PixelAlign positions image frames on the display's pixel grid and fits their images.
export class PixelAlign {
  // Watch element sizes and window changes without doing layout work during scrolling.
  constructor(container, selector) {
    this.container = container;
    this.selector = selector;
    this.elements = new Map();
    this.frame = null;
    this.schedule = this.schedule.bind(this);
    this.observer = typeof ResizeObserver === "function" ? new ResizeObserver(this.schedule) : null;
    this.observer?.observe(container);
    window.addEventListener("resize", this.schedule);
    this.update();
  }

  // Refresh observed elements after photos or metadata change.
  update() {
    const elements = new Set(this.container.querySelectorAll(this.selector));

    for (const element of this.elements.keys()) {
      if (!elements.has(element)) {
        this.observer?.unobserve(element);
        this.elements.delete(element);
      }
    }

    for (const element of elements) {
      if (!this.elements.has(element)) {
        this.elements.set(element, { x: 0, y: 0 });
        this.observer?.observe(element);
      }
    }

    this.schedule();
  }

  // Batch layout reads before style writes and coalesce observer notifications.
  schedule() {
    if (this.frame !== null) {
      return;
    }

    this.frame = requestAnimationFrame(() => {
      this.frame = null;
      this.align();
    });
  }

  // Snap document positions so ordinary scrolling cannot change the correction.
  align() {
    const dpr = window.devicePixelRatio || 1;
    const updates = [];

    for (const [element, offset] of this.elements) {
      const rect = element.getBoundingClientRect();

      if (!rect.width || !rect.height) {
        continue;
      }

      const x = rect.left + window.scrollX - offset.x;
      const y = rect.top + window.scrollY - offset.y;
      const image = element.querySelector("img");
      const scale = image?.width ? rect.width / image.width : 1;
      updates.push({ element, image, scale, x: Math.round(x * dpr) / dpr - x, y: Math.round(y * dpr) / dpr - y });
    }

    for (const { element, image, scale, x, y } of updates) {
      element.style.left = `${x}px`;
      element.style.top = `${y}px`;
      if (image) {
        image.style.transform = `scale(${scale})`;
      }
      this.elements.set(element, { x, y });
    }
  }

  // Release observers and pending work when the view is removed.
  destroy() {
    this.observer?.disconnect();
    window.removeEventListener("resize", this.schedule);

    if (this.frame !== null) {
      cancelAnimationFrame(this.frame);
    }

    this.elements.clear();
  }
}
