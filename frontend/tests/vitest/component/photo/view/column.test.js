import { mount, config } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { nextTick, reactive } from "vue";
import Column from "component/photo/view/column.vue";
import Photo from "model/photo";

describe("component/photo/view/column", () => {
  let wrapper, settings;

  beforeEach(() => {
    vi.stubGlobal("devicePixelRatio", 2);
    settings = reactive({
      features: { places: false },
      display: { originals: true, retinaLightbox: true, metadata: { cards: [] } },
    });
    wrapper = mount(Column, {
      props: {
        photos: [
          new Photo({
            UID: "ptew0y5i0re4pjte",
            Hash: "5523095c3b3a69d41085e38d85d016aec4ad6c46",
            Width: 953,
            Height: 1301,
            Title: "Odd dimensions",
          }),
        ],
      },
      global: {
        mocks: {
          $config: { ...config.global.mocks.$config, getSettings: () => settings },
        },
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
    vi.unstubAllGlobals();
  });

  it("keeps intrinsic dimensions and fractional Retina CSS width", () => {
    const image = wrapper.get(".column-image__img");
    expect(image.attributes("width")).toBe("953");
    expect(image.attributes("height")).toBe("1301");
    expect(image.element.style.transform).toBe("scale(0.5)");
    expect(image.attributes("src")).toMatch(/\/original$/);
    expect(wrapper.get(".column-photo").element.style.maxWidth).toBe("476.5px");
  });

  it("keeps normal image sizing when Retina mode is disabled", async () => {
    settings.display.retinaLightbox = false;
    await nextTick();
    expect(wrapper.get(".column-image__img").element.style.transform).toBe("scale(1)");
  });

  it("updates image dimensions when moving to a display with a different density", async () => {
    vi.stubGlobal("devicePixelRatio", 1);
    window.dispatchEvent(new Event("resize"));
    await nextTick();
    expect(wrapper.get(".column-image__img").element.style.transform).toBe("scale(1)");
  });

  it("continues to use generated previews when originals are disabled", async () => {
    settings.display.originals = false;
    await nextTick();
    expect(wrapper.get(".column-image__img").attributes("src")).toBe("/t.jpg");
    expect(wrapper.get(".column-photo").element.style.maxWidth).toBe("50px");
  });
});
