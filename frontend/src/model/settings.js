import $api from "common/api";
import { defaultMetadataLayout, MetadataView } from "common/metadata";
import Model from "./model";

// Settings stores the nested user/admin settings tree used across the UI.
export class Settings extends Model {
  changed(area, key) {
    if (typeof this.__originalValues[area] === "undefined") {
      return false;
    }

    return this[area][key] !== this.__originalValues[area][key];
  }

  setValues(values, scalarOnly) {
    if (!values) return;

    if (values.maps?.style === "basic" || values.maps?.style === "offline") {
      values.maps.style = "";
    }

    if (!values.index) {
      values.index = {
        addAIKeywords: true,
      };
    } else if (typeof values.index.addAIKeywords === "undefined" || values.index.addAIKeywords === null) {
      values.index.addAIKeywords = true;
    }

    // Ensure display settings exist with defaults.
    if (!values.display) {
      values.display = {
        originals: false,
        imagePacking: false,
        lightboxBorder: 1,
        retinaLightbox: false,
        retinaThumbnails: false,
        metadata: {
          cards: defaultMetadataLayout(MetadataView.Cards),
          list: defaultMetadataLayout(MetadataView.List),
          lightbox: defaultMetadataLayout(MetadataView.Lightbox),
        },
      };
    } else {
      if (typeof values.display.originals === "undefined") {
        values.display.originals = false;
      }

      if (typeof values.display.imagePacking === "undefined") {
        values.display.imagePacking = false;
      }

      if (typeof values.display.lightboxBorder === "undefined") {
        values.display.lightboxBorder = 1;
      }

      if (typeof values.display.retinaLightbox === "undefined") {
        values.display.retinaLightbox = false;
      }

      if (typeof values.display.retinaThumbnails === "undefined") {
        values.display.retinaThumbnails = false;
      }
    }

    if (!values.display.metadata) {
      values.display.metadata = {};
    }

    if (!Array.isArray(values.display.metadata.cards)) {
      values.display.metadata.cards = defaultMetadataLayout(MetadataView.Cards);
    }

    if (!Array.isArray(values.display.metadata.list)) {
      values.display.metadata.list = defaultMetadataLayout(MetadataView.List);
    }

    if (!Array.isArray(values.display.metadata.lightbox)) {
      values.display.metadata.lightbox = defaultMetadataLayout(MetadataView.Lightbox);
    }

    // Filter visibility toggles for the photo search toolbar. Default every
    // filter to visible so existing users (and fresh installs without an
    // explicit Filters block in settings.yml) see the historical UI.
    const filterDefaults = {
      country: true,
      camera: true,
      view: true,
      order: true,
      year: true,
      month: true,
      color: true,
      category: true,
    };
    if (!values.display.filters || typeof values.display.filters !== "object") {
      values.display.filters = { ...filterDefaults };
    } else {
      for (const key of Object.keys(filterDefaults)) {
        if (typeof values.display.filters[key] !== "boolean") {
          values.display.filters[key] = filterDefaults[key];
        }
      }
    }

    super.setValues(values, scalarOnly);

    return this;
  }

  load() {
    return $api.get("settings").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return $api.post("settings", this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
  }
}

export default Settings;
