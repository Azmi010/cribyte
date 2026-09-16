type ViewMode = "grid" | "list";
type SortBy = "name" | "size" | "updated_at" | "created_at";
type SortOrder = "asc" | "desc";

interface ViewState {
  viewMode: ViewMode;
  sortBy: SortBy;
  sortOrder: SortOrder;
}

function loadFromStorage(): ViewState {
  if (typeof localStorage === "undefined") {
    return { viewMode: "grid", sortBy: "name", sortOrder: "asc" };
  }
  return {
    viewMode: (localStorage.getItem("viewMode") as ViewMode) || "grid",
    sortBy: (localStorage.getItem("sortBy") as SortBy) || "name",
    sortOrder: (localStorage.getItem("sortOrder") as SortOrder) || "asc",
  };
}

const state: ViewState = $state(loadFromStorage());

function persist() {
  if (typeof localStorage === "undefined") return;
  localStorage.setItem("viewMode", state.viewMode);
  localStorage.setItem("sortBy", state.sortBy);
  localStorage.setItem("sortOrder", state.sortOrder);
}

export const view = {
  get viewMode() {
    return state.viewMode;
  },
  get sortBy() {
    return state.sortBy;
  },
  get sortOrder() {
    return state.sortOrder;
  },

  setViewMode(mode: ViewMode) {
    state.viewMode = mode;
    persist();
  },

  setSortBy(by: SortBy) {
    state.sortBy = by;
    persist();
  },

  setSortOrder(order: SortOrder) {
    state.sortOrder = order;
    persist();
  },

  toggleSortOrder() {
    state.sortOrder = state.sortOrder === "asc" ? "desc" : "asc";
    persist();
  },
};
