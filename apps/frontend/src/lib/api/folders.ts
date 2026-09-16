import type { FolderItem } from "$lib/types";
import { api } from "./client";

export function listFolders(parentFolderId?: string | null): Promise<FolderItem[]> {
  const query = new URLSearchParams();
  if (parentFolderId) query.set("folder_id", parentFolderId);
  const qs = query.toString();
  return api<FolderItem[]>(`/folders${qs ? `?${qs}` : ""}`);
}

export function listFolderContents(folderId: string): Promise<FolderItem[]> {
  return api<FolderItem[]>(`/folders/${folderId}/contents`);
}

export function listStarredFolders(): Promise<FolderItem[]> {
  return api<FolderItem[]>("/folders/starred");
}

export function listTrashFolders(): Promise<FolderItem[]> {
  return api<FolderItem[]>("/folders/trash");
}

export function getFolder(id: string): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}`);
}

interface CreateFolderPayload {
  name: string;
  parent_folder_id?: string | null;
}

export function createFolder(payload: CreateFolderPayload): Promise<FolderItem> {
  return api<FolderItem>("/folders/", {
    method: "POST",
    body: payload,
  });
}

export function renameFolder(id: string, name: string): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}`, {
    method: "PATCH",
    body: { name },
  });
}

export function moveFolder(id: string, parentFolderId: string | null): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}/move`, {
    method: "POST",
    body: { parent_folder_id: parentFolderId },
  });
}

export function trashFolder(id: string): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}`, {
    method: "DELETE",
  });
}

export function restoreFolder(id: string): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}/restore`, {
    method: "POST",
  });
}

export function permanentDeleteFolder(id: string): Promise<void> {
  return api<void>(`/folders/${id}/permanent-delete`, {
    method: "POST",
  });
}

export function toggleFolderStarred(id: string, starred: boolean): Promise<FolderItem> {
  return api<FolderItem>(`/folders/${id}/star`, {
    method: "POST",
    body: { starred },
  });
}
