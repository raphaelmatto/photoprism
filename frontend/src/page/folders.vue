<template>
  <div ref="page" tabindex="-1" class="p-page p-page-folders p-page-keywords not-selectable" :class="$config.aclClasses('folders')">
    <v-form ref="form" validate-on="invalid-input" class="p-folders-search p-page__navigation" @submit.prevent="onSubmit()">
      <v-toolbar :density="$vuetify.display.smAndDown ? 'compact' : 'default'" color="secondary" class="page-toolbar">
        <v-text-field
          v-model="query"
          :density="density"
          hide-details
          clearable
          single-line
          rounded="pill"
          variant="solo-filled"
          color="surface-variant"
          autocorrect="off"
          autocapitalize="none"
          autocomplete="off"
          prepend-inner-icon="mdi-tune"
          :placeholder="$gettext('Search')"
          class="input-search background-inherit elevation-0"
          :class="{ 'input-search--expanded': expanded }"
          @keyup.esc.exact="hideExpansionPanel"
          @click:prepend-inner.stop="toggleExpansionPanel"
          @click:clear="query = ''"
        ></v-text-field>
        <v-btn
          :title="$gettext('Reverse Sort')"
          :icon="reverse ? 'mdi-sort-descending' : 'mdi-sort-ascending'"
          class="action-reverse ms-1"
          @click.prevent="toggleReverse"
        ></v-btn>
        <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
      </v-toolbar>

      <div class="toolbar-expansion-panel">
        <v-expand-transition>
          <v-card v-show="expanded" flat color="secondary">
            <v-card-text class="dense">
              <v-row dense>
                <v-col cols="12" sm="4" class="p-year-select">
                  <v-select
                    v-model="filter.year"
                    :label="$gettext('Year')"
                    :menu-props="{ maxHeight: 346 }"
                    single-line
                    hide-details
                    variant="solo-filled"
                    :density="density"
                    :items="yearOptions"
                    item-title="text"
                    item-value="value"
                    @update:model-value="onFilterChanged"
                  >
                  </v-select>
                </v-col>
                <v-col cols="12" sm="4" class="p-category-select">
                  <v-select
                    v-model="filter.category"
                    :label="$gettext('Category')"
                    :menu-props="{ maxHeight: 346 }"
                    single-line
                    hide-details
                    variant="solo-filled"
                    :density="density"
                    :items="categories"
                    item-title="text"
                    item-value="value"
                    @update:model-value="onFilterChanged"
                  >
                  </v-select>
                </v-col>
                <v-col cols="12" sm="4" class="p-sort-select">
                  <v-select
                    v-model="filter.order"
                    :label="$gettext('Sort Order')"
                    :menu-props="{ maxHeight: 400 }"
                    single-line
                    hide-details
                    variant="solo-filled"
                    :density="density"
                    :items="sortingOptions"
                    item-title="text"
                    item-value="value"
                    @update:model-value="onOrderChanged"
                  >
                  </v-select>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-expand-transition>
      </div>
    </v-form>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content pa-4">
      <div v-if="tree.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-folder-off-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">{{ $gettext("No folders found") }}</div>
          <div class="mt-2">
            {{ $gettext("Folders mirror the directory layout of your originals folder.") }}
          </div>
        </v-alert>
      </div>
      <div v-else-if="visibleTree.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-magnify-close" class="no-results" variant="outlined">
          <div class="font-weight-bold">{{ $gettext("No matching folders") }}</div>
          <div class="mt-2">
            {{ $gettext("Try a different search term or clear the filter to see all folders.") }}
          </div>
        </v-alert>
      </div>
      <div v-else class="keywords-tree">
        <keyword-tree-node
          v-for="node in visibleTree"
          :key="node.path"
          :node="node"
          :depth="0"
          :default-expanded="false"
          :is-expanded="isExpanded"
          :set-expanded="setExpanded"
          @browse="browseFolder"
        />
      </div>
    </div>
  </div>
</template>

<script>
import * as options from "options/options";
import Album from "model/album";
import PLoading from "component/loading.vue";
import PActionMenu from "component/action/menu.vue";
import KeywordTreeNode from "component/keyword-tree-node.vue";
import {
  loadExpandedPaths,
  saveExpandedPaths,
  loadTreeReverse,
  saveTreeReverse,
  filterTreeByName,
  collectFilterExpansions,
} from "common/tree-state";

const STORAGE_KEY = "photoprism.tree.folders";
const ORDER_STORAGE_KEY = `${STORAGE_KEY}.order`;

// loadStoredOrder reads the persisted sort key. Falls back to "name" so
// the tree renders alphabetically on first visit, matching the photo
// toolbar's natural default.
function loadStoredOrder() {
  if (typeof window === "undefined") {
    return "name";
  }
  try {
    return window.localStorage?.getItem(ORDER_STORAGE_KEY) || "name";
  } catch {
    return "name";
  }
}

function saveStoredOrder(value) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage?.setItem(ORDER_STORAGE_KEY, value || "name");
  } catch {
    // ignore
  }
}

export default {
  name: "PPageFolders",
  components: {
    PLoading,
    PActionMenu,
    KeywordTreeNode,
  },
  data() {
    // Seed the category list from $config so the dropdown is populated
    // immediately; matches the Albums page's approach.
    let categories = [{ value: "", text: this.$gettext("All Categories") }];
    if (this.$config.albumCategories().length > 0) {
      categories = categories.concat(
        this.$config.albumCategories().map((cat) => ({ value: cat, text: cat }))
      );
    }

    return {
      loading: true,
      rawTree: [],
      query: "",
      expanded: false,
      reverse: loadTreeReverse(STORAGE_KEY),
      expandedPaths: loadExpandedPaths(STORAGE_KEY),
      filter: {
        year: "",
        category: "",
        order: loadStoredOrder(),
      },
      categories,
      sortingOptions: [
        { value: "favorites", text: this.$gettext("Favorites") },
        { value: "name", text: this.$gettext("Name") },
        { value: "place", text: this.$gettext("Location") },
        { value: "date", text: this.$gettext("Date") },
        { value: "added", text: this.$gettext("Added") },
        { value: "edited", text: this.$gettext("Edited") },
      ],
      all: {
        years: [{ value: "", text: this.$gettext("All Years") }],
      },
    };
  },
  computed: {
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
    yearOptions() {
      return this.all.years.concat(options.IndexedYears());
    },
    tree() {
      return this.reverse ? this.reverseTree(this.rawTree) : this.rawTree;
    },
    visibleTree() {
      return filterTreeByName(this.tree, this.query);
    },
    filterExpansions() {
      return this.query.trim() ? collectFilterExpansions(this.tree, this.query) : null;
    },
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
    this.load();
  },
  beforeUnmount() {
    this.$view.leave(this);
  },
  methods: {
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => this.load(),
        },
      ];
    },
    onSubmit() {
      // The form's submit-on-enter would normally re-run the search; for the
      // folder tree the v-model on `query` already filters live, so this is
      // intentionally a no-op.
    },
    onFilterChanged() {
      this.load();
    },
    onOrderChanged(value) {
      saveStoredOrder(value);
      this.load();
    },
    toggleExpansionPanel() {
      this.expanded = !this.expanded;
    },
    hideExpansionPanel() {
      this.expanded = false;
    },
    async load() {
      this.loading = true;
      try {
        const folders = await this.fetchFolders();
        this.rawTree = this.buildTree(folders);
      } catch (e) {
        this.rawTree = [];
      } finally {
        this.loading = false;
      }
    },
    // fetchFolders pulls every matching folder album in batches because the
    // album search endpoint caps `count` at 1000 per request. Year, category,
    // and order are passed straight through so the server does the filtering
    // and ordering work — we just walk the result list and build a tree.
    async fetchFolders() {
      const limit = 1000;
      const all = [];
      let offset = 0;

      const baseParams = {
        type: "folder",
        count: limit,
        order: this.filter.order || "name",
      };
      if (this.filter.year) {
        baseParams.year = this.filter.year;
      }
      if (this.filter.category) {
        baseParams.category = this.filter.category;
      }

      for (;;) {
        const resp = await Album.search({ ...baseParams, offset });
        const batch = resp?.models || [];
        all.push(...batch);
        if (batch.length < limit) {
          break;
        }
        offset += batch.length;
        // Safety stop to avoid infinite loops on a misbehaving backend.
        if (offset > 50000) {
          break;
        }
      }

      return all;
    },
    // buildTree converts a flat list of folder albums into a nested
    // {name, path, uid, slug, children} structure keyed off each folder's
    // slash-separated Path. Intermediate path segments that do happen to
    // correspond to a stored folder record carry that record's uid/slug
    // so clicking them navigates to the folder photos page; purely
    // synthetic intermediates (rare) just expand/collapse.
    //
    // Sibling order at every level preserves the server's order, so the
    // active "Sort Order" filter has a visible effect (Date sort shows
    // siblings by date, etc.). The `reverse` toggle flips this at every
    // depth.
    buildTree(folders) {
      const root = new Map();

      for (const folder of folders) {
        const rawPath = (folder?.Path || "").trim();
        if (!rawPath) continue;

        const segments = rawPath
          .split("/")
          .map((s) => s.trim())
          .filter(Boolean);

        if (segments.length === 0) continue;

        let parentMap = root;
        const acc = [];
        for (let i = 0; i < segments.length; i++) {
          const seg = segments[i];
          acc.push(seg);
          const pathSoFar = acc.join("/");

          if (!parentMap.has(seg)) {
            parentMap.set(seg, {
              name: seg,
              path: pathSoFar,
              uid: "",
              slug: "",
              children: new Map(),
            });
          }

          if (i === segments.length - 1) {
            const node = parentMap.get(seg);
            node.uid = folder.UID || "";
            node.slug = folder.Slug || "";
          }

          parentMap = parentMap.get(seg).children;
        }
      }

      return this.mapToArray(root);
    },
    // mapToArray walks the nested Map structure and emits plain arrays,
    // preserving insertion order (i.e. the order the API returned folders
    // in, which reflects the active Sort Order filter).
    mapToArray(map) {
      return [...map.values()].map((node) => ({
        name: node.name,
        path: node.path,
        uid: node.uid,
        slug: node.slug,
        children: this.mapToArray(node.children),
      }));
    },
    reverseTree(nodes) {
      if (!Array.isArray(nodes) || nodes.length === 0) {
        return [];
      }
      return [...nodes]
        .reverse()
        .map((node) => ({ ...node, children: this.reverseTree(node.children) }));
    },
    toggleReverse() {
      this.reverse = !this.reverse;
      saveTreeReverse(STORAGE_KEY, this.reverse);
    },
    isExpanded(path) {
      if (this.filterExpansions) {
        return this.filterExpansions.has(path);
      }
      return this.expandedPaths.has(path);
    },
    setExpanded(path, value) {
      if (this.filterExpansions) {
        return;
      }
      const next = new Set(this.expandedPaths);
      if (value) {
        next.add(path);
      } else {
        next.delete(path);
      }
      this.expandedPaths = next;
      saveExpandedPaths(STORAGE_KEY, next);
    },
    browseFolder(node) {
      if (!node?.uid || !node?.slug) {
        return;
      }
      this.$router.push({ name: "folder", params: { album: node.uid, slug: node.slug } });
    },
  },
};
</script>
