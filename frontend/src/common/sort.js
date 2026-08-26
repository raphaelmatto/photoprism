// Sort value helpers that map between the directionless frontend order
// ("date", "name", "added", ...) and the backend wire protocol, which
// still uses "newest"/"oldest" for the two date-sort variants.
//
// The UI exposes a single "Date" option plus a Reverse Sort toggle; the
// helpers collapse "newest"/"oldest" on read and map "date" to "newest"
// on write, letting the backend's `reverse=true|false` flag carry the
// direction for all sort orders.

// Maps the frontend order to the backend wire value. "date" is the only
// directionless alias on the frontend and maps to "newest"; the caller
// sends `reverse=true|false` alongside to invert direction.
export function toWireOrder(order) {
  if (order === "date") return "newest";
  return order;
}

// Normalizes legacy stored/URL values. Returns the canonical frontend
// order plus an optional reverse override (null = caller keeps its own
// reverse state, true/false = caller should adopt this direction).
export function fromWireOrder(storedOrder) {
  if (storedOrder === "newest") return { order: "date", reverse: false };
  if (storedOrder === "oldest") return { order: "date", reverse: true };
  return { order: storedOrder, reverse: null };
}

// Resolves reverse-sort state in precedence order: explicit URL direction,
// legacy directional URL order, stored user preference, then model default.
export function resolveReverse({ queryReverse, queryOrder, storedReverse, defaultOrder }) {
  if (queryReverse === "true" || queryReverse === "false") {
    return queryReverse === "true";
  }

  if (queryOrder) {
    const query = fromWireOrder(queryOrder);
    if (query.reverse !== null) return query.reverse;
  }

  if (storedReverse === "true" || storedReverse === "false") {
    return storedReverse === "true";
  }

  if (defaultOrder) {
    const fallback = fromWireOrder(defaultOrder);
    if (fallback.reverse !== null) return fallback.reverse;
  }

  return false;
}
