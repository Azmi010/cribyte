import { writable } from "svelte/store";

function createSearchStore() {
  const { subscribe, set } = writable("");

  return {
    subscribe,
    set,
    clear: () => set(""),
  };
}

export const searchQuery = createSearchStore();
