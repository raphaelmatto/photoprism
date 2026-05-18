import { mount, config as VTUConfig } from "@vue/test-utils";
import { describe, it, expect, beforeEach, vi } from "vitest";
import * as contexts from "options/contexts";
import { nextTick } from "vue";
import PLightbox from "component/lightbox.vue";
import Album from "model/album";

const defaultStubs = {
  "v-dialog": true,
  "v-icon": true,
  "v-slider": true,
  "p-lightbox-menu": true,
  "p-sidebar-info": true,
};

const mountLightbox = (options = {}) =>
  mount(PLightbox, {
    ...options,
    global: {
      ...(options.global || {}),
      stubs: {
        ...defaultStubs,
        ...(options.global?.stubs || {}),
      },
    },
  });

describe("PLightbox (low-mock, jsdom-friendly)", () => {
  beforeEach(() => {
    localStorage.removeItem("lightbox.info");
    sessionStorage.removeItem("lightbox.muted");
  });

  it("toggleInfo updates info and localStorage when visible", async () => {
    const wrapper = mountLightbox();
    await wrapper.setData({ visible: true });

    // Use exposed onShortCut to trigger info toggle (KeyI)
    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem("lightbox.info")).toBe("true");

    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem("lightbox.info")).toBe("false");
  });

  it("toggleMute writes sessionStorage without requiring video or exposed state", async () => {
    const wrapper = mountLightbox();
    expect(sessionStorage.getItem("lightbox.muted")).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyM" });
    expect(sessionStorage.getItem("lightbox.muted")).toBe("true");
    await wrapper.vm.onShortCut({ code: "KeyM" });
    expect(sessionStorage.getItem("lightbox.muted")).toBe("false");
  });

  it("getPadding returns expected structure for large and small screens", async () => {
    const wrapper = mountLightbox();
    // Large viewport
    const large = wrapper.vm.$options.methods.getPadding.call(wrapper.vm, { x: 1200, y: 800 }, { width: 4000, height: 3000 });
    expect(large).toHaveProperty("top");
    expect(large).toHaveProperty("bottom");
    expect(large).toHaveProperty("left");
    expect(large).toHaveProperty("right");

    // Small viewport (<= mobileBreakpoint) should yield zeros
    const small = wrapper.vm.$options.methods.getPadding.call(wrapper.vm, { x: 360, y: 640 }, { width: 1200, height: 800 });
    expect(small).toEqual({ top: 0, bottom: 0, left: 0, right: 0 });
  });

  it("KeyI is ignored when dialog is not visible", async () => {
    const wrapper = mountLightbox();
    expect(localStorage.getItem("lightbox.info")).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyI" });
    expect(localStorage.getItem("lightbox.info")).toBeNull();
  });

  it("getViewport falls back to window size without content ref", () => {
    const wrapper = mountLightbox();
    const vp = wrapper.vm.$options.methods.getViewport.call(wrapper.vm);
    expect(vp.x).toBeGreaterThan(0);
    expect(vp.y).toBeGreaterThan(0);
  });

  it("menuActions marks Download action visible when allowed", () => {
    const wrapper = mountLightbox();
    const ctx = {
      $gettext: VTUConfig.global.mocks.$gettext,
      $pgettext: VTUConfig.global.mocks.$pgettext,
      // minimal state needed by menuActions visibility checks
      canManageAlbums: false,
      canArchive: false,
      canDownload: true,
      collection: null,
      context: contexts.Default,
      model: {},
    };
    const actions = wrapper.vm.$options.methods.menuActions.call(ctx);
    const download = actions.find((a) => a?.name === "download");
    expect(download).toBeTruthy();
    expect(download.visible).toBe(true);
  });

  it("sets a collection cover from the detailed photo file hash", async () => {
    const wrapper = mountLightbox();
    const hash = "bc2eb607c7a70a378d4393b5f24fc12574ea3a9a";
    const album = new Album({ UID: "atezhkib1587p7fo" });
    album.setCover = vi.fn().mockResolvedValue(album);

    const ctx = {
      canManageAlbums: true,
      collection: album,
      model: {
        Hash: "",
        fileHash: () => hash,
      },
      pauseSlideshow: vi.fn(),
      log: vi.fn(),
      $notify: {
        success: vi.fn(),
      },
      $gettext: VTUConfig.global.mocks.$gettext,
    };

    await wrapper.vm.$options.methods.onSetCollectionCover.call(ctx);

    expect(album.setCover).toHaveBeenCalledWith(hash);
    expect(ctx.log).not.toHaveBeenCalled();
    expect(ctx.$notify.success).toHaveBeenCalledWith("Changes successfully saved");
  });

  it("binds the active theme to the dialog and updates it on refresh", async () => {
    const subscriptions = {};

    const wrapper = mountLightbox({
      global: {
        stubs: {
          "v-dialog": {
            name: "VDialog",
            props: ["theme"],
            template: '<div class="v-dialog-stub" :data-theme="theme"><slot /></div>',
          },
        },
        mocks: {
          $config: {
            ...VTUConfig.global.mocks.$config,
            themeName: "default",
          },
          $event: {
            ...VTUConfig.global.mocks.$event,
            subscribe: (event, handler) => {
              subscriptions[event] = handler;
              return event;
            },
          },
        },
      },
    });

    expect(wrapper.find(".v-dialog-stub").attributes("data-theme")).toBe("default");

    await subscriptions["view.refresh"]("view.refresh", { themeName: "nordic" });
    await nextTick();

    expect(wrapper.find(".v-dialog-stub").attributes("data-theme")).toBe("nordic");
  });

  it("only exposes zoom toggling when the current slide has a distinct zoom target", () => {
    const wrapper = mountLightbox();
    const button = document.createElement("button");
    const methods = wrapper.vm.$options.methods;
    const baseContext = {
      isZoomable: true,
      isZoomedIn: true,
      zoomToggleElement: button,
      $gettext: VTUConfig.global.mocks.$gettext,
      updateZoomToggleButton() {
        return methods.updateZoomToggleButton.call(this);
      },
      canToggleZoom(slide) {
        return methods.canToggleZoom.call(this, slide);
      },
    };

    const noOpZoomContext = {
      ...baseContext,
      pswp: () => ({
        currSlide: {
          currZoomLevel: 1,
          zoomLevels: {
            initial: 1,
            secondary: 1,
          },
        },
      }),
    };

    methods.syncZoomState.call(noOpZoomContext);

    expect(noOpZoomContext.isZoomable).toBe(false);
    expect(noOpZoomContext.isZoomedIn).toBe(false);
    expect(button.getAttribute("title")).toBe("Zoom in");

    const zoomedContext = {
      ...baseContext,
      pswp: () => ({
        currSlide: {
          currZoomLevel: 1.6,
          zoomLevels: {
            initial: 1,
            secondary: 2,
          },
        },
      }),
    };

    methods.syncZoomState.call(zoomedContext);

    expect(zoomedContext.isZoomable).toBe(true);
    expect(zoomedContext.isZoomedIn).toBe(true);
    expect(button.getAttribute("title")).toBe("Zoom out");
  });

  it("formats under-image text from the configured card metadata layout", () => {
    const wrapper = mountLightbox({
      global: {
        mocks: {
          $config: {
            ...VTUConfig.global.mocks.$config,
            getSettings: () => ({
              features: {
                edit: true,
                favorites: true,
                download: true,
                archive: true,
              },
              display: {
                metadata: {
                  cards: ["keywords", "date", "caption"],
                },
              },
            }),
          },
        },
      },
    });

    const html = wrapper.vm.$options.methods.formatCaption.call(wrapper.vm.$.proxy, {
      Caption: "Test caption",
      DetailsKeywords: "keyword-one, keyword-two",
      TakenAtLocal: "2026-01-11T12:44:32Z",
      TimeZone: "America/New_York",
    });

    expect(html).toContain("pswp__dynamic-caption-field--keywords");
    expect(html).toContain("pswp__dynamic-caption-field--clickable");
    expect(html).toContain("keyword-one");
    expect(html).toContain("keyword-two");
    expect(html).toContain("Jan 11, 2026");
    expect(html).toContain("Test caption");
  });

  it("hides empty under-image fields from the configured card metadata layout", () => {
    const wrapper = mountLightbox({
      global: {
        mocks: {
          $config: {
            ...VTUConfig.global.mocks.$config,
            getSettings: () => ({
              features: {
                edit: true,
                favorites: true,
                download: true,
                archive: true,
              },
              display: {
                metadata: {
                  cards: ["keywords", "caption"],
                },
              },
            }),
          },
        },
      },
    });

    const html = wrapper.vm.$options.methods.formatCaption.call(wrapper.vm.$.proxy, {
      Caption: "Only caption",
      DetailsKeywords: "",
    });

    expect(html).not.toContain("pswp__dynamic-caption-field--keywords");
    expect(html).toContain("Only caption");
  });

  it("uses richer current-photo metadata for captions without replacing slide models", async () => {
    const wrapper = mountLightbox({
      global: {
        mocks: {
          $config: {
            ...VTUConfig.global.mocks.$config,
            getSettings: () => ({
              features: {
                edit: true,
                favorites: true,
                download: true,
                archive: true,
              },
              display: {
                metadata: {
                  cards: ["keywords", "caption"],
                },
              },
            }),
          },
        },
      },
    });

    const slideModel = {
      UID: "photo-1",
      Caption: "",
      DetailsKeywords: "",
      Thumbs: {
        fit_1920: {
          src: "/static/example.jpg",
          w: 1920,
          h: 1080,
        },
      },
      Hash: "abc123",
    };
    const detailPhoto = {
      UID: "photo-1",
      Caption: "Detailed caption",
      DetailsKeywords: "keyword-one, keyword-two",
      TakenAtLocal: "2026-01-11T12:44:32Z",
      TimeZone: "America/New_York",
    };

    const slide = { index: 0 };

    await wrapper.setData({
      model: detailPhoto,
      models: [slideModel],
      index: 0,
    });

    const activeModel = wrapper.vm.$options.methods.activeCaptionModel.call(wrapper.vm.$.proxy, slide);
    const html = wrapper.vm.$options.methods.formatCaption.call(wrapper.vm.$.proxy, activeModel);
    await nextTick();

    expect(wrapper.vm.$data.models[0]).toEqual(slideModel);
    expect(wrapper.vm.$data.models[0].Thumbs.fit_1920.src).toBe("/static/example.jpg");
    expect(wrapper.vm.$data.model.UID).toBe("photo-1");
    expect(wrapper.vm.$data.model.DetailsKeywords).toBe("keyword-one, keyword-two");
    expect(wrapper.vm.$data.model.Caption).toBe("Detailed caption");
    expect(activeModel).toBe(wrapper.vm.$data.model);
    expect(html).toContain("keyword-one");
    expect(html).toContain("keyword-two");
  });
});
