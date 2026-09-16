import type { PreviewKind } from "$lib/types";

export const API_BASE = "/api";

// Sinkron dengan MaxUploadSize di backend
export const MAX_UPLOAD_SIZE = 100 * 1024 * 1024; // 100 MB

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const size = bytes / Math.pow(1024, i);
  return `${size < 10 ? size.toFixed(1) : Math.round(size)} ${units[i]}`;
}

const MIME_TO_PREVIEW: [test: (mime: string) => boolean, kind: PreviewKind][] = [
  [(m) => m.startsWith("image/"), "image"],
  [(m) => m.startsWith("video/"), "video"],
  [(m) => m.startsWith("audio/"), "audio"],
  [(m) => m === "application/pdf", "pdf"],
  [(m) => m.startsWith("text/"), "text"],
  [
    (m) =>
      [
        "application/json",
        "application/xml",
        "application/javascript",
        "application/typescript",
        "application/x-yaml",
        "application/x-sh",
        "application/x-httpd-php",
        "application/sql",
        "application/graphql",
        "application/toml",
      ].includes(m),
    "text",
  ],
];

export function getPreviewKind(mimeType: string): PreviewKind {
  for (const [test, kind] of MIME_TO_PREVIEW) {
    if (test(mimeType)) return kind;
  }
  return "other";
}

export const TEXT_EXTENSIONS = new Set([
  ".txt",
  ".md",
  ".markdown",
  ".json",
  ".yaml",
  ".yml",
  ".toml",
  ".xml",
  ".html",
  ".htm",
  ".css",
  ".scss",
  ".less",
  ".js",
  ".jsx",
  ".ts",
  ".tsx",
  ".mjs",
  ".cjs",
  ".py",
  ".rb",
  ".go",
  ".rs",
  ".java",
  ".kt",
  ".c",
  ".cpp",
  ".h",
  ".hpp",
  ".sh",
  ".bash",
  ".zsh",
  ".fish",
  ".ps1",
  ".sql",
  ".graphql",
  ".gql",
  ".env",
  ".gitignore",
  ".dockerignore",
  ".editorconfig",
  ".svelte",
  ".vue",
  ".astro",
  ".csv",
  ".log",
  ".ini",
  ".cfg",
  ".conf",
  ".prisma",
  ".proto",
  ".tf",
  ".hcl",
]);
