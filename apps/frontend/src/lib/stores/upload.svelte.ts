import { uploadFile as apiUpload } from "$lib/api/files";

export type UploadStatus = "pending" | "uploading" | "done" | "error";

export interface UploadEntry {
  id: string;
  file: File;
  parentFolderId: string | null;
  progress: number; // 0–100
  status: UploadStatus;
  error?: string;
  abort?: () => void;
  onDone?: () => void;
}

const uploads: UploadEntry[] = $state([]);

let idCounter = 0;

export const uploadStore = {
  get uploads() {
    return uploads;
  },

  get hasActive() {
    return uploads.some((u) => u.status === "uploading" || u.status === "pending");
  },

  get hasAny() {
    return uploads.length > 0;
  },

  addUpload(file: File, parentFolderId: string | null, onDone?: () => void): string {
    const id = `upload-${++idCounter}-${Date.now()}`;

    const entry: UploadEntry = $state({
      id,
      file,
      parentFolderId,
      progress: 0,
      status: "pending",
      onDone,
    });

    uploads.push(entry);
    startUpload(entry);
    return id;
  },

  removeUpload(id: string) {
    const idx = uploads.findIndex((u) => u.id === id);
    if (idx !== -1) {
      const entry = uploads[idx];
      if (entry.abort && (entry.status === "pending" || entry.status === "uploading")) {
        entry.abort();
      }
      uploads.splice(idx, 1);
    }
  },

  clearCompleted() {
    const remaining = uploads.filter((u) => u.status !== "done");
    uploads.length = 0;
    uploads.push(...remaining);
  },

  cancelAll() {
    for (const entry of uploads) {
      if (entry.abort && (entry.status === "pending" || entry.status === "uploading")) {
        entry.abort();
      }
    }
    uploads.length = 0;
  },
};

function startUpload(entry: UploadEntry) {
  entry.status = "uploading";

  const { promise, abort } = apiUpload({
    file: entry.file,
    parentFolderId: entry.parentFolderId,
    onProgress(loaded, total) {
      entry.progress = Math.round((loaded / total) * 100);
    },
  });

  entry.abort = abort;

  promise
    .then(() => {
      entry.status = "done";
      entry.progress = 100;
      entry.onDone?.();
    })
    .catch((err) => {
      if (err instanceof Error && err.message === "Upload cancelled") {
        return;
      }
      entry.status = "error";
      entry.error = err instanceof Error ? err.message : "Upload gagal";
    });
}
