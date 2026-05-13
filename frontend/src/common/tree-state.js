// tree-state.js persists the set of expanded paths for a hierarchical tree
// view (Keywords, Folders, ...) in localStorage. Each tree gets its own
// storage key so they don't collide. Failures (private mode, quota,
// disabled storage) degrade silently to "everything collapsed".

// loadExpandedPaths returns a Set of expanded tree paths previously stored
// under the given key, or an empty Set when no usable state exists.
export function loadExpandedPaths(key) {
  if (!key || typeof window === "undefined") {
    return new Set();
  }
  try {
    const raw = window.localStorage?.getItem(key);
    if (!raw) {
      return new Set();
    }
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return new Set();
    }
    return new Set(parsed.filter((entry) => typeof entry === "string"));
  } catch {
    return new Set();
  }
}

// saveExpandedPaths writes the current Set of expanded paths back to
// localStorage. Quiet on quota or access errors.
export function saveExpandedPaths(key, paths) {
  if (!key || typeof window === "undefined") {
    return;
  }
  try {
    const serialized = JSON.stringify([...paths]);
    window.localStorage?.setItem(key, serialized);
  } catch {
    // localStorage may be disabled (private mode) or full; ignore.
  }
}

// filterTreeByName returns a copy of the tree with nodes whose name does not
// match `query` removed, keeping every ancestor whose subtree contains a
// match. Case-insensitive substring match on node.name only. An empty or
// whitespace-only query returns the original tree unchanged.
export function filterTreeByName(tree, query) {
  if (!Array.isArray(tree) || tree.length === 0) {
    return tree;
  }
  const needle = (query || "").trim().toLowerCase();
  if (!needle) {
    return tree;
  }
  const out = [];
  for (const node of tree) {
    const children = filterTreeByName(node.children, needle);
    const selfMatches = typeof node.name === "string" && node.name.toLowerCase().includes(needle);
    if (selfMatches || (Array.isArray(children) && children.length > 0)) {
      out.push({ ...node, children: Array.isArray(children) ? children : [] });
    }
  }
  return out;
}

// loadTreeReverse returns the saved sort-reverse flag for a tree (true or
// false). Defaults to false when nothing is stored or storage is unreadable.
export function loadTreeReverse(key) {
  if (!key || typeof window === "undefined") {
    return false;
  }
  try {
    return window.localStorage?.getItem(`${key}.reverse`) === "1";
  } catch {
    return false;
  }
}

// saveTreeReverse persists the sort-reverse flag for a tree. Stored as "1"
// or "0" so the storage value is human-readable and tiny.
export function saveTreeReverse(key, value) {
  if (!key || typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage?.setItem(`${key}.reverse`, value ? "1" : "0");
  } catch {
    // localStorage unavailable; ignore.
  }
}

// collectFilterExpansions returns the Set of paths that should be visible
// (i.e. expanded) when the tree is filtered by `query`, so every match is
// reachable from the root. Includes ancestors of matching nodes; matching
// nodes themselves only get expanded if they have descendant matches.
export function collectFilterExpansions(tree, query) {
  const expand = new Set();
  const needle = (query || "").trim().toLowerCase();
  if (!Array.isArray(tree) || tree.length === 0 || !needle) {
    return expand;
  }
  const walk = (nodes) => {
    let subtreeMatched = false;
    for (const node of nodes) {
      const childMatched = Array.isArray(node.children) ? walk(node.children) : false;
      const selfMatches = typeof node.name === "string" && node.name.toLowerCase().includes(needle);
      if (childMatched) {
        expand.add(node.path);
      }
      if (childMatched || selfMatches) {
        subtreeMatched = true;
      }
    }
    return subtreeMatched;
  };
  walk(tree);
  return expand;
}
