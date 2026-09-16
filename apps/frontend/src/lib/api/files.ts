import type { FileItem, PreviewResult } from "$lib/types";
import { API_BASE } from "$lib/constants";
import { api } from "./client";

interface ListFilesParams {
  folder_id?: string | null;
  search?: string;
}

export function listFiles(params: ListFilesParams = {}): Promise<FileItem[]> {
  const query = new URLSearchParams();
  if (params.folder_id) query.set("folder_id", params.folder_id);
  if (params.search) query.set("search", params.search);
  const qs = query.toString();
  return api<FileItem[]>(`/files${qs ? `?${qs}` : ""}`);
}

export function searchFiles(query: string): Promise<FileItem[]> {
  return listFiles({ search: query });
}

export function listStarred(): Promise<FileItem[]> {
  return api<FileItem[]>("/files/starred");
}

export function listTrash(): Promise<FileItem[]> {
  return api<FileItem[]>("/files/trash");
}

export function getFile(id: string): Promise<FileItem> {
  return api<FileItem>(`/files/${id}`);
}

export function previewFile(id: string): Promise<PreviewResult> {
  return api<PreviewResult>(`/files/${id}/preview`);
}

export function serveUrl(id: string): string {
  return `${API_BASE}/files/${id}/serve`;
}

export function downloadUrl(id: string): string {
  return `${API_BASE}/files/${id}/download`;
}

interface UploadParams {
  file: File;
  parentFolderId?: string | null;
  name?: string;
  onProgress?: (loaded: number, total: number) => void;
}

// Pakai XHR supaya bisa track upload progress
export function uploadFile(params: UploadParams): {
  promise: Promise<FileItem>;
  abort: () => void;
} {
  const { file, parentFolderId, name, onProgress } = params;

  const xhr = new XMLHttpRequest();
  const formData = new FormData();
  formData.append("file", file);
  if (parentFolderId) formData.append("parent_folder_id", parentFolderId);
  if (name) formData.append("name", name);

  const promise = new Promise<FileItem>((resolve, reject) => {
    xhr.open("POST", `${API_BASE}/files/upload`);
    xhr.withCredentials = true;

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(e.loaded, e.total);
      }
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText));
        } catch {
          reject(new Error("Invalid JSON response"));
        }
      } else {
        let message = `Upload failed (${xhr.status})`;
        try {
          const err = JSON.parse(xhr.responseText);
          if (err.error) message = err.error;
        } catch {
          // ignore
        }
        reject(new Error(message));
      }
    });

    xhr.addEventListener("error", () => reject(new Error("Network error")));
    xhr.addEventListener("abort", () => reject(new Error("Upload cancelled")));

    xhr.send(formData);
  });

  return { promise, abort: () => xhr.abort() };
}

export function renameFile(id: string, name: string): Promise<FileItem> {
  return api<FileItem>(`/files/${id}`, {
    method: "PATCH",
    body: { name },
  });
}

export function moveFile(id: string, parentFolderId: string | null): Promise<FileItem> {
  return api<FileItem>(`/files/${id}/move`, {
    method: "POST",
    body: { parent_folder_id: parentFolderId },
  });
}

export function trashFile(id: string): Promise<FileItem> {
  return api<FileItem>(`/files/${id}`, {
    method: "DELETE",
  });
}

export function restoreFile(id: string): Promise<FileItem> {
  return api<FileItem>(`/files/${id}/restore`, {
    method: "POST",
  });
}

export function permanentDeleteFile(id: string): Promise<void> {
  return api<void>(`/files/${id}/permanent-delete`, {
    method: "POST",
  });
}

export function toggleFileStarred(id: string, starred: boolean): Promise<FileItem> {
  return api<FileItem>(`/files/${id}/star`, {
    method: "POST",
    body: { starred },
  });
}
