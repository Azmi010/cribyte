let refreshFn: (() => void) | null = null;
let currentFolderId: string | null = null;

export const driveRefresh = {
  register(fn: () => void, folderId: string | null) {
    refreshFn = fn;
    currentFolderId = folderId;
  },

  unregister(fn: () => void) {
    if (refreshFn === fn) {
      refreshFn = null;
      currentFolderId = null;
    }
  },

  get folderId() {
    return currentFolderId;
  },

  run() {
    refreshFn?.();
  },
};
