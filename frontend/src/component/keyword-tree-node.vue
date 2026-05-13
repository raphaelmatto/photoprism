<template>
  <div class="keyword-tree__node">
    <div class="keyword-tree__row" :style="indentStyle">
      <button
        v-if="hasChildren"
        type="button"
        class="keyword-tree__toggle"
        :aria-expanded="expanded"
        :title="expanded ? $gettext('Collapse') : $gettext('Expand')"
        @click.stop="toggle"
      >
        <i class="mdi" :class="expanded ? 'mdi-chevron-down' : 'mdi-chevron-right'" />
      </button>
      <span v-else class="keyword-tree__toggle keyword-tree__toggle--placeholder" aria-hidden="true"></span>
      <button type="button" class="keyword-tree__label" @click="onBrowse">
        {{ node.name }}<span v-if="node.count > 0" class="keyword-tree__count"> ({{ node.count }})</span>
      </button>
    </div>
    <div v-if="expanded && hasChildren" class="keyword-tree__children">
      <keyword-tree-node
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :default-expanded="false"
        :is-expanded="isExpanded"
        :set-expanded="setExpanded"
        @browse="$emit('browse', $event)"
      />
    </div>
  </div>
</template>

<script>
// KeywordTreeNode renders a single keyword or folder in a hierarchical tree
// and recursively renders its children when expanded. Pass `isExpanded` and
// `setExpanded` callbacks to share collapse state across the whole tree
// (e.g. persisted in localStorage by the parent page); when omitted, each
// node falls back to local per-instance state seeded by `defaultExpanded`.
export default {
  name: "KeywordTreeNode",
  props: {
    node: { type: Object, required: true },
    depth: { type: Number, default: 0 },
    defaultExpanded: { type: Boolean, default: false },
    isExpanded: { type: Function, default: null },
    setExpanded: { type: Function, default: null },
  },
  emits: ["browse"],
  data() {
    return { localExpanded: this.defaultExpanded };
  },
  computed: {
    hasChildren() {
      return Array.isArray(this.node.children) && this.node.children.length > 0;
    },
    expanded() {
      if (typeof this.isExpanded === "function") {
        return !!this.isExpanded(this.node.path);
      }
      return this.localExpanded;
    },
    indentStyle() {
      // Each level indents by a fixed amount; the chevron column already
      // takes care of the first visual offset.
      return { paddingInlineStart: `${this.depth * 16}px` };
    },
  },
  methods: {
    toggle() {
      if (!this.hasChildren) {
        return;
      }
      const next = !this.expanded;
      if (typeof this.setExpanded === "function") {
        this.setExpanded(this.node.path, next);
      } else {
        this.localExpanded = next;
      }
    },
    onBrowse() {
      // Emit the whole node so each page can route on the fields it cares
      // about (name for keyword search, uid/slug for folder navigation, etc.).
      this.$emit("browse", this.node);
    },
  },
};
</script>
