import { mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PScroll from "component/scroll.vue";

describe("PScroll", () => {
  let wrapper;
  let loadMore;

  beforeEach(() => {
    vi.useFakeTimers();

    loadMore = vi.fn();

    Object.defineProperty(window, "innerHeight", {
      configurable: true,
      value: 800,
    });

    Object.defineProperty(window, "scrollY", {
      configurable: true,
      writable: true,
      value: 0,
    });

    Object.defineProperty(document.documentElement, "scrollHeight", {
      configurable: true,
      value: 2400,
    });

    wrapper = mount(PScroll, {
      props: {
        loadMore,
        loadDisabled: false,
        loadDistance: 800,
      },
    });
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }

    vi.useRealTimers();
  });

  it("loads more when the viewport reaches the configured bottom distance", async () => {
    window.scrollY = 800;

    window.dispatchEvent(new Event("scroll"));
    await wrapper.vm.$nextTick();

    expect(loadMore).toHaveBeenCalledTimes(1);
  });

  it("does not load more before the viewport reaches the configured bottom distance", async () => {
    window.scrollY = 799;

    window.dispatchEvent(new Event("scroll"));
    await wrapper.vm.$nextTick();

    expect(loadMore).not.toHaveBeenCalled();
  });
});
