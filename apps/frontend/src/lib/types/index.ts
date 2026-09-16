export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface FolderItem {
  id: string;
  name: string;
  starred: boolean;
  parent_folder_id: string | null;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface FileItem {
  id: string;
  name: string;
  mime_type: string;
  size: number;
  extension: string | null;
  starred: boolean;
  parent_folder_id: string | null;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export type PreviewKind = "image" | "video" | "audio" | "pdf" | "text" | "other";

export interface PreviewResult {
  kind: PreviewKind;
  mime_type: string;
  name: string;
  size: number;
  content?: string;
  signed_url?: string;
}

export interface StorageUsage {
  used: number;
  total: number;
}

export type DriveItem = ({ type: "folder" } & FolderItem) | ({ type: "file" } & FileItem);

export interface ApiError {
  error: string;
}
