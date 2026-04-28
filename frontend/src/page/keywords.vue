<template>
  <div ref="page" tabindex="-1" class="p-page p-page-keywords not-selectable" :class="$config.aclClasses('keywords')">
    <v-toolbar flat :density="$vuetify.display.smAndDown ? 'compact' : 'default'" color="secondary" class="page-toolbar">
      <v-toolbar-title class="page__title">{{ $gettext("Keywords") }}</v-toolbar-title>
      <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
    </v-toolbar>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content pa-4">
      <div v-if="groups.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-tag-off-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">{{ $gettext("No keywords found") }}</div>
          <div class="mt-2">
            {{ $gettext("Keywords are extracted from image metadata during indexing.") }}
          </div>
        </v-alert>
      </div>
      <template v-else>
        <div v-for="group in groups" :key="group.name" class="keywords-group mb-6">
          <div class="keywords-group__title text-overline text-medium-emphasis mb-2">{{ group.name }}</div>
          <div class="keywords-group__list">
            <button
              v-for="kw in group.keywords"
              :key="kw.full"
              class="keywords-list-item"
              @click="browseKeyword(kw.leaf)"
            >{{ kw.leaf }}</button>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script>
import $api from "common/api";
import PLoading from "component/loading.vue";
import PActionMenu from "component/action/menu.vue";

export default {
  name: "PPageKeywords",
  components: {
    PLoading,
    PActionMenu,
  },
  data() {
    return {
      loading: true,
      groups: [],
    };
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
          this.groups = this.buildGroups(resp.data || []);
        })
        .catch(() => {
          this.groups = [];
        })
        .finally(() => {
          this.loading = false;
        });
    },
    buildGroups(keywords) {
      const map = new Map();
      const NO_CATEGORY = this.$gettext("No category");

      for (const kw of keywords) {
        const raw = kw.Keyword || "";
        if (!raw) continue;

        const sep = raw.indexOf("|");
        const category = sep === -1 ? NO_CATEGORY : raw.slice(0, sep);
        const leaf = sep === -1 ? raw : raw.slice(sep + 1);

        if (!map.has(category)) {
          map.set(category, []);
        }
        map.get(category).push({ full: raw, leaf });
      }

      // Sort groups alphabetically, "No category" last.
      const groups = [];
      let noCategory = null;

      for (const [name, kws] of map.entries()) {
        if (name === NO_CATEGORY) {
          noCategory = { name, keywords: kws };
        } else {
          groups.push({ name, keywords: kws });
        }
      }

      groups.sort((a, b) => a.name.localeCompare(b.name));

      if (noCategory) {
        groups.push(noCategory);
      }

      return groups;
    },
    browseKeyword(fullKeyword) {
      this.$router.push({ name: "browse", query: { q: fullKeyword } });
    },
  },
};
</script>
