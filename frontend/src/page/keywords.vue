<template>
  <div ref="page" tabindex="-1" class="p-page p-page-keywords not-selectable" :class="$config.aclClasses('keywords')">
    <div class="p-keywords-search p-page__navigation">
      <v-toolbar :density="$vuetify.display.smAndDown ? 'compact' : 'default'" color="secondary" class="page-toolbar">
        <v-text-field
          v-model="query"
          :density="$vuetify.display.smAndDown ? 'compact' : 'comfortable'"
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
    </div>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content pa-4">
      <div v-if="tree.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-tag-off-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">{{ $gettext("No keywords found") }}</div>
          <div class="mt-2">
            {{ $gettext("Keywords are extracted from image metadata during indexing.") }}
          </div>
        </v-alert>
      </div>
      <div v-else-if="visibleTree.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-magnify-close" class="no-results" variant="outlined">
          <div class="font-weight-bold">{{ $gettext("No matching keywords") }}</div>
          <div class="mt-2">
            {{ $gettext("Try a different search term or clear the filter to see all keywords.") }}
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
          @browse="browseKeyword"
        />
      </div>
    </div>
  </div>
</template>

<script>
import $api from "common/api";
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

const STORAGE_KEY = "photoprism.tree.keywords";

export default {
  name: "PPageKeywords",
  components: {
    PLoading,
    PActionMenu,
    KeywordTreeNode,
  },
  data() {
    return {
      loading: true,
      rawTree: [],
      query: "",
      reverse: loadTreeReverse(STORAGE_KEY),
      // Persisted set of expanded tree paths. Reassigned on every change so
      // Vue reactivity picks it up (Set mutations aren't tracked deeply).
      expandedPaths: loadExpandedPaths(STORAGE_KEY),
    };
  },
  computed: {
    // tree applies the reverse-sort flag on top of the raw alphabetical
    // tree. Reversing recursively keeps sibling order consistent at every
    // depth so a Z-A toggle shows Z..A at the top level and within each
    // expanded branch.
    tree() {
      return this.reverse ? this.reverseTree(this.rawTree) : this.rawTree;
    },
    // visibleTree narrows the rendered tree to entries whose name (or a
    // descendant name) matches the search query. Empty query passes through.
    visibleTree() {
      return filterTreeByName(this.tree, this.query);
    },
    // filterExpansions holds the paths that must be expanded while filtering
    // so each match is reachable. Empty when no filter is active.
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
    load() {
      this.loading = true;
      $api
        .get("keywords")
        .then((resp) => {
          this.rawTree = this.buildTree(resp.data || []);
        })
        .catch(() => {
          this.rawTree = [];
        })
        .finally(() => {
          this.loading = false;
        });
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
    // buildTree converts the flat keyword list returned by the API into a
    // nested {name, path, children} structure of arbitrary depth, splitting
    // each Keyword on "|". Single-segment keywords get a synthetic
    // "No category" parent so they group together visually below the named
    // categories.
    buildTree(keywords) {
      const noCategory = this.$gettext("No category");
      const root = new Map();
      const standalone = new Map();

      for (const kw of keywords) {
        const raw = (kw.Keyword || "").trim();
        if (!raw) continue;

        const segments = raw
          .split("|")
          .map((s) => s.trim())
          .filter(Boolean);

        if (segments.length === 0) continue;

        if (segments.length === 1) {
          if (!standalone.has(segments[0])) {
            standalone.set(segments[0], {
              name: segments[0],
              path: segments[0],
              children: new Map(),
            });
          }
          continue;
        }

        let parentMap = root;
        const acc = [];
        for (const seg of segments) {
          acc.push(seg);
          if (!parentMap.has(seg)) {
            parentMap.set(seg, {
              name: seg,
              path: acc.join("|"),
              children: new Map(),
            });
          }
          parentMap = parentMap.get(seg).children;
        }
      }

      const tree = this.mapToSortedArray(root);

      if (standalone.size > 0) {
        tree.push({
          name: noCategory,
          path: `__no_category__`,
          children: this.mapToSortedArray(standalone),
        });
      }

      return tree;
    },
    mapToSortedArray(map) {
      return [...map.values()]
        .map((node) => ({
          name: node.name,
          path: node.path,
          children: this.mapToSortedArray(node.children),
        }))
        .sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: "base" }));
    },
    isExpanded(path) {
      // While filtering, force ancestors of matches open so results are
      // reachable. The persisted state is left untouched in this mode.
      if (this.filterExpansions) {
        return this.filterExpansions.has(path);
      }
      return this.expandedPaths.has(path);
    },
    setExpanded(path, value) {
      // Ignore manual toggles while filtering — filter mode is transient,
      // and persisting auto-expansions would clobber the saved state.
      if (this.filterExpansions) {
        return;
      }
      // Replace the Set rather than mutating it so Vue's reactivity picks
      // up the change and re-renders dependent nodes.
      const next = new Set(this.expandedPaths);
      if (value) {
        next.add(path);
      } else {
        next.delete(path);
      }
      this.expandedPaths = next;
      saveExpandedPaths(STORAGE_KEY, next);
    },
    browseKeyword(node) {
      // Wrap the keyword in quotes so the backend treats it as a literal
      // phrase: multi-word keywords match the exact tag, and single-word
      // keywords skip the prefix-LIKE that would otherwise pull in unrelated
      // words sharing the same prefix.
      const name = node?.name;
      if (!name) return;
      this.$router.push({ name: "browse", query: { q: `"${name}"` } });
    },
  },
};
</script>
