<template>
  <div class="p-photos p-photo-view-column" :style="columnStyle">
    <div v-if="photos.length === 0" class="pa-3">
      <v-alert color="surface-variant" :icon="isSharedView ? 'mdi-image-off' : 'mdi-lightbulb-outline'" class="no-results" variant="outlined">
        <div class="font-weight-bold">
          {{ $gettext(`No pictures found`) }}
        </div>
        <div class="mt-2">
          {{ $gettext(`Try again using other filters or keywords.`) }}
          <template v-if="!isSharedView">
            {{ $gettext(`In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`) }}
          </template>
        </div>
      </v-alert>
    </div>
    <div v-else class="search-results photo-results column-view" :class="{ 'select-results': selectMode }">
      <div
        v-for="(photo, index) in photos"
        :key="photo.ID"
        :data-id="photo.ID"
        :data-uid="photo.UID"
        class="column-item result"
        :class="photo.classes()"
        @contextmenu.stop="onContextMenu($event, index)"
      >
        <div class="column-photo" :style="columnPhotoStyle(photo)">
          <div
            class="column-image"
            @touchstart.passive="input.touchStart($event, index)"
            @touchend.stop="onClick($event, index)"
            @touchmove.stop
            @mousedown.stop="input.mouseDown($event, index)"
            @click.stop.prevent="onClick($event, index)"
          >
            <img v-bind="thumbInfo(photo)" loading="lazy" class="column-image__img" />
            <div class="column-image__overlay"></div>

            <button
              class="input-select"
              @touchstart.stop="input.touchStart($event, index)"
              @touchend.stop="onSelect($event, index)"
              @touchmove.stop
              @click.stop.prevent="onSelect($event, index)"
            >
              <i class="mdi mdi-check-circle select-on" />
              <i class="mdi mdi-circle-outline select-off" />
            </button>

            <button
              v-if="!isSharedView && $config.feature('favorites')"
              class="input-favorite"
              @touchstart.stop="input.touchStart($event, index)"
              @touchend.stop="toggleLike($event, index)"
              @touchmove.stop
              @click.stop.prevent="toggleLike($event, index)"
            >
              <i v-if="photo.Favorite" class="mdi mdi-star text-favorite favorite-on" />
              <i v-else class="mdi mdi-star-outline favorite-off" />
            </button>
          </div>

          <div class="column-meta">
            <div class="meta-details meta-fields">
              <component
                :is="item.clickable ? 'button' : 'div'"
                v-for="item in columnMetadataItems(photo)"
                :key="`${photo.ID}-${item.key}`"
                :title="item.label"
                :class="item.className"
                @click.exact="onMetadataAction(item.action, index)"
              >
                <i v-if="item.icon" class="mdi" :class="item.icon" />
                {{ item.text }}
              </component>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { hasMetadataText, metadataIcon, metadataLabel, metadataLayout, metadataText, MetadataView } from "common/metadata";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import Thumb from "model/thumb";
import * as contexts from "options/contexts";

export default {
  name: "PPhotoViewColumn",
  props: {
    photos: {
      type: Array,
      default: () => [],
    },
    openPhoto: {
      type: Function,
      default: () => {},
    },
    editPhoto: {
      type: Function,
      default: () => {},
    },
    openDate: {
      type: Function,
      default: () => {},
    },
    openLocation: {
      type: Function,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    context: {
      type: String,
      default: "",
    },
    selectMode: Boolean,
    isSharedView: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const input = new Input();
    const settings = this.$config.getSettings();
    const featPlaces = settings.features.places;

    return {
      input,
      featPlaces,
      contexts,
    };
  },
  computed: {
    columnLayout() {
      return metadataLayout(this.$config.getSettings(), MetadataView.Cards);
    },
    columnStyle() {
      const width = Number(this.$config?.getSettings?.()?.display?.lightboxBorder);
      const borderWidth = Number.isFinite(width) ? Math.max(0, width) : 1;
      return { "--p-lightbox-border-width": `${borderWidth}px` };
    },
  },
  methods: {
    thumbInfo(photo) {
      const display = this.$config.getSettings()?.display;
      const dpr = display?.retinaLightbox ? window.devicePixelRatio || 1 : 1;
      const thumbs = Thumb.fromPhoto(photo).Thumbs;
      const original = display?.originals ? thumbs?.original : null;
      const source = original || this.$util.thumb(thumbs, Math.round(window.innerWidth * dpr), Math.round(window.innerHeight * dpr));
      return {
        src: source.src,
        width: Math.round(source.w / dpr),
        height: Math.round(source.h / dpr),
        alt: photo.Title,
      };
    },
    // Cap the wrapper at the image's natural width so a long description does
    // not expand the framed container past the image. The wrapper still
    // narrows on smaller viewports because max-width yields to available space.
    columnPhotoStyle(photo) {
      const info = this.thumbInfo(photo);
      if (!info?.width) {
        return null;
      }
      return { maxWidth: `${info.width}px` };
    },
    columnMetadataItems(photo) {
      return this.columnLayout
        .map((fieldId, index) => {
          if (fieldId === "location" && (!this.featPlaces || photo?.Country === "zz")) {
            return null;
          } else if (!hasMetadataText(photo, fieldId)) {
            return null;
          }

          const action = this.metadataAction(fieldId);

          return {
            key: `${fieldId}-${index}`,
            label: metadataLabel(fieldId),
            text: metadataText(photo, fieldId),
            icon: metadataIcon(fieldId, photo),
            clickable: action !== "",
            className: this.metadataClass(fieldId, action !== ""),
            action,
          };
        })
        .filter(Boolean);
    },
    metadataAction(fieldId) {
      switch (fieldId) {
        case "date":
          return "date";
        case "location":
          return "location";
        case "filename":
          return this.isSharedView ? "open" : "files";
        case "camera":
        case "lens":
        case "exposure":
        case "fileInfo":
          return this.isSharedView ? "open" : "details";
        case "title":
        case "caption":
        case "keywords":
        default:
          return this.isSharedView ? "open" : "edit";
      }
    },
    metadataClass(fieldId, clickable) {
      const classes = ["meta-field", `meta-field--${fieldId}`, `meta-${fieldId}`];

      if (clickable) {
        classes.push("clickable");
      }

      return classes.join(" ");
    },
    onMetadataAction(action, index) {
      switch (action) {
        case "date":
          this.openDate(index);
          break;
        case "location":
          this.openLocation(index);
          break;
        case "files":
          this.editPhoto(index, "files");
          break;
        case "details":
          this.editPhoto(index, "details");
          break;
        case "edit":
          this.editPhoto(index);
          break;
        case "open":
          this.openPhoto(index);
          break;
        default:
          break;
      }
    },
    toggleLike(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const photo = this.photos[index];

      if (!photo) {
        return;
      }

      photo.toggleLike();
    },
    onSelect(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.toggle(this.photos[index]);
      }
    },
    toggle(photo) {
      this.$clipboard.toggle(photo);
      this.$forceUpdate();
    },
    onClick(ev, index) {
      const inputType = this.input.eval(ev, index);
      const longClick = inputType === ClickLong;

      if (inputType === InputInvalid) {
        return;
      }

      if (longClick || this.selectMode) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index);
        } else {
          this.toggle(this.photos[index]);
        }
      } else {
        this.openPhoto(index);
      }
    },
    onContextMenu(ev, index) {
      if (this.$isMobile) {
        ev.preventDefault();
        ev.stopPropagation();
        this.selectRange(index);
      }
    },
    selectRange(index) {
      this.$clipboard.addRange(index, this.photos);
      this.$forceUpdate();
    },
  },
};
</script>
