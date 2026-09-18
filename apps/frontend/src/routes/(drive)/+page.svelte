<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import type { FileItem, FolderItem } from "$lib/types";
  import { formatFileSize } from "$lib/constants";
  import { view } from "$lib/stores/view.svelte";
  import { uploadStore } from "$lib/stores/upload.svelte";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
  import FileContextMenu from "$lib/components/drive/FileContextMenu.svelte";
  import FolderContextMenu from "$lib/components/drive/FolderContextMenu.svelte";
  import CreateFolderModal from "$lib/components/modals/CreateFolderModal.svelte";
  import RenameModal from "$lib/components/modals/RenameModal.svelte";
  import MoveModal from "$lib/components/modals/MoveModal.svelte";
  import DeleteConfirmModal from "$lib/components/modals/DeleteConfirmModal.svelte";
  import CloudUploadIcon from "@lucide/svelte/icons/cloud-upload";
  import RefreshCwIcon from "@lucide/svelte/icons/refresh-cw";
  import XIcon from "@lucide/svelte/icons/x";
  import LoaderIcon from "@lucide/svelte/icons/loader";
  import { toast } from "svelte-sonner";

  const folderId = $derived(page.url.searchParams.get("folder") ?? null);

  let folders = $state<FolderItem[]>([]);
  let files = $state<FileItem[]>([]);
  let breadcrumbs = $state<{ id: string | null; name: string }[]>([{ id: null, name: "Home" }]);
  let loading = $state(true);
  let dragOver = $state(false);

  const sortedFolders = $derived(() => {
    const sorted = [...folders];
    sorted.sort((a, b) => {
      const dir = view.sortOrder === "asc" ? 1 : -1;
      if (view.sortBy === "name") return a.name.localeCompare(b.name) * dir;
      if (view.sortBy === "updated_at") return (new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime()) * dir;
      if (view.sortBy === "created_at") return (new Date(a.created_at).getTime() - new Date(b.created_at).getTime()) * dir;
      return a.name.localeCompare(b.name) * dir;
    });
    return sorted;
  });

  const sortedFiles = $derived(() => {
    const sorted = [...files];
    sorted.sort((a, b) => {
      const dir = view.sortOrder === "asc" ? 1 : -1;
      if (view.sortBy === "name") return a.name.localeCompare(b.name) * dir;
      if (view.sortBy === "size") return (a.size - b.size) * dir;
      if (view.sortBy === "updated_at") return (new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime()) * dir;
      if (view.sortBy === "created_at") return (new Date(a.created_at).getTime() - new Date(b.created_at).getTime()) * dir;
      return a.name.localeCompare(b.name) * dir;
    });
    return sorted;
  });

  const isEmpty = $derived(folders.length === 0 && files.length === 0 && !loading);

  async function fetchContents(currentFolderId: string | null) {
    loading = true;
    try {
      const [folderRes, fileRes] = await Promise.all([
        foldersApi.listFolders(currentFolderId),
        filesApi.listFiles({ folder_id: currentFolderId }),
      ]);
      folders = folderRes ?? [];
      files = fileRes ?? [];

      if (currentFolderId) {
        try {
          const folder = await foldersApi.getFolder(currentFolderId);
          breadcrumbs = [
            { id: null, name: "Home" },
            { id: folder.id, name: folder.name },
          ];
        } catch {
          breadcrumbs = [{ id: null, name: "Home" }];
        }
      } else {
        breadcrumbs = [{ id: null, name: "Home" }];
      }
    } catch {
      toast.error("Gagal memuat konten folder");
      folders = [];
      files = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    fetchContents(folderId);
  });

  function navigateToFolder(id: string | null) {
    if (id) {
      goto(`/?folder=${id}`);
    } else {
      goto("/");
    }
  }

  function openFile(file: FileItem) {
    toast.info("Preview belum diimplementasi");
  }

  // --- Context Menu State ---
  let fileContextMenuOpen = $state(false);
  let fileContextMenuX = $state(0);
  let fileContextMenuY = $state(0);
  let fileContextMenuItem = $state<FileItem | null>(null);

  let folderContextMenuOpen = $state(false);
  let folderContextMenuX = $state(0);
  let folderContextMenuY = $state(0);
  let folderContextMenuItem = $state<FolderItem | null>(null);

  function handleFolderContextMenu(e: MouseEvent, folder: FolderItem) {
    e.preventDefault();
    folderContextMenuItem = folder;
    folderContextMenuX = e.clientX;
    folderContextMenuY = e.clientY;
    folderContextMenuOpen = true;
  }

  function handleFileContextMenu(e: MouseEvent, file: FileItem) {
    e.preventDefault();
    fileContextMenuItem = file;
    fileContextMenuX = e.clientX;
    fileContextMenuY = e.clientY;
    fileContextMenuOpen = true;
  }

  // --- Modal State ---
  let createFolderOpen = $state(false);
  let renameOpen = $state(false);
  let renameItem = $state<FileItem | FolderItem | null>(null);
  let renameType = $state<"file" | "folder">("file");
  let moveOpen = $state(false);
  let moveItem = $state<FileItem | FolderItem | null>(null);
  let moveType = $state<"file" | "folder">("file");
  let deleteConfirmOpen = $state(false);
  let deleteConfirmItem = $state<FileItem | FolderItem | null>(null);
  let deleteConfirmType = $state<"file" | "folder">("file");

  function handleNewFolder() {
    createFolderOpen = true;
  }

  // --- File Context Menu Actions ---
  function handleFilePreview(item: FileItem) {
    openFile(item);
  }

  function handleFileDownload(item: FileItem) {
    const url = filesApi.downloadUrl(item.id);
    const a = document.createElement("a");
    a.href = url;
    a.download = item.name;
    a.click();
  }

  function handleFileRename(item: FileItem) {
    renameItem = item;
    renameType = "file";
    renameOpen = true;
  }

  function handleFileMove(item: FileItem) {
    moveItem = item;
    moveType = "file";
    moveOpen = true;
  }

  async function handleFileToggleStar(item: FileItem) {
    try {
      await filesApi.toggleFileStarred(item.id, !item.starred);
      toast.success(item.starred ? "Dihapus dari bintang" : "Ditandai bintang");
      fetchContents(folderId);
    } catch {
      toast.error("Gagal mengubah status bintang");
    }
  }

  function handleFileTrash(item: FileItem) {
    deleteConfirmItem = item;
    deleteConfirmType = "file";
    deleteConfirmOpen = true;
  }

  // --- Folder Context Menu Actions ---
  function handleFolderOpen(item: FolderItem) {
    navigateToFolder(item.id);
  }

  function handleFolderRename(item: FolderItem) {
    renameItem = item;
    renameType = "folder";
    renameOpen = true;
  }

  function handleFolderMove(item: FolderItem) {
    moveItem = item;
    moveType = "folder";
    moveOpen = true;
  }

  async function handleFolderToggleStar(item: FolderItem) {
    try {
      await foldersApi.toggleFolderStarred(item.id, !item.starred);
      toast.success(item.starred ? "Dihapus dari bintang" : "Ditandai bintang");
      fetchContents(folderId);
    } catch {
      toast.error("Gagal mengubah status bintang");
    }
  }

  function handleFolderTrash(item: FolderItem) {
    deleteConfirmItem = item;
    deleteConfirmType = "folder";
    deleteConfirmOpen = true;
  }

  // --- Drag & Drop ---
  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    dragOver = true;
  }

  function handleDragLeave() {
    dragOver = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    dragOver = false;
    const droppedFiles = e.dataTransfer?.files;
    if (!droppedFiles || droppedFiles.length === 0) return;

    for (let i = 0; i < droppedFiles.length; i++) {
      uploadStore.addUpload(droppedFiles[i], folderId);
    }
    toast.success(`${droppedFiles.length} file ditambahkan ke antrian upload`);
  }

  function handleDone() {
    fetchContents(folderId);
  }
</script>

<Toolbar
  {breadcrumbs}
  onNavigate={navigateToFolder}
  onNewFolder={handleNewFolder}
/>

<div
  class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto relative"
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  role="region"
  aria-label="File browser"
>
  {#if dragOver}
    <div class="absolute inset-0 z-30 bg-primary/5 border-2 border-dashed border-primary rounded-lg flex items-center justify-center pointer-events-none">
      <div class="flex flex-col items-center gap-2 text-primary">
        <CloudUploadIcon class="size-10" />
        <span class="text-sm font-semibold">Drop file di sini untuk upload</span>
      </div>
    </div>
  {/if}

  <div
    class="relative rounded-lg border border-dashed border-border/80 bg-muted/15 p-5 flex flex-col sm:flex-row items-center justify-between gap-4 transition-colors hover:border-primary/50"
    style="background-image: radial-gradient(circle, var(--border) 1px, transparent 1px); background-size: 20px 20px;"
  >
    <div class="flex items-center gap-4">
      <div class="size-10 rounded-lg bg-muted/60 border border-border/70 flex items-center justify-center text-primary shrink-0">
        <CloudUploadIcon class="size-5" />
      </div>
      <div>
        <div class="text-xs font-semibold text-foreground">
          Drag & drop files here to upload directly to this folder
        </div>
        <div class="text-[11px] text-muted-foreground font-mono mt-0.5">
          Instant encryption · Direct stream ingestion · Max single file: 100 MB
        </div>
      </div>
    </div>
    <div class="text-[11px] font-mono text-muted-foreground bg-muted/50 border border-border/70 px-2 py-1 rounded shrink-0">
      Shift + U
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="flex flex-col items-center gap-3">
        <LoaderIcon class="size-6 text-primary animate-spin" />
        <span class="text-xs font-mono text-muted-foreground">Memuat konten...</span>
      </div>
    </div>
  {:else if isEmpty}
    <EmptyState variant="folder" />
  {:else if view.viewMode === "grid"}
    <GridView
      folders={sortedFolders()}
      files={sortedFiles()}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {:else}
    <ListView
      folders={sortedFolders()}
      files={sortedFiles()}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {/if}
</div>

<!-- Floating Upload Progress Bar Widget -->
{#if uploadStore.hasActive}
  {@const activeUpload = uploadStore.uploads.find((u) => u.status === "uploading")}
  {#if activeUpload}
    <div class="fixed bottom-4 right-4 z-50 w-80 rounded-lg border border-border/90 bg-card shadow-xl p-3.5 space-y-2.5 animate-in fade-in slide-in-from-bottom-3 duration-200">
      <div class="flex items-center justify-between text-xs">
        <div class="flex items-center gap-2 font-medium text-foreground truncate">
          <RefreshCwIcon class="size-3.5 text-primary animate-spin" />
          <span class="truncate">Uploading {activeUpload.file.name}</span>
        </div>
        <button
          type="button"
          onclick={() => uploadStore.removeUpload(activeUpload.id)}
          class="text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
          aria-label="Cancel upload"
        >
          <XIcon class="size-3.5" />
        </button>
      </div>
      <div class="h-1.5 w-full bg-muted rounded-full overflow-hidden">
        <div
          class="h-full bg-primary rounded-full transition-all duration-300"
          style="width: {activeUpload.progress}%"
        ></div>
      </div>
      <div class="flex items-center justify-between text-[11px] font-mono text-muted-foreground">
        <span>{activeUpload.progress}% of {formatFileSize(activeUpload.file.size)}</span>
        <span>{uploadStore.uploads.filter((u) => u.status === "uploading" || u.status === "pending").length} file(s)</span>
      </div>
    </div>
  {/if}
{/if}

<!-- Context Menus -->
<FileContextMenu
  bind:open={fileContextMenuOpen}
  x={fileContextMenuX}
  y={fileContextMenuY}
  item={fileContextMenuItem}
  context="drive"
  onPreview={handleFilePreview}
  onDownload={handleFileDownload}
  onRename={handleFileRename}
  onMove={handleFileMove}
  onToggleStar={handleFileToggleStar}
  onTrash={handleFileTrash}
/>

<FolderContextMenu
  bind:open={folderContextMenuOpen}
  x={folderContextMenuX}
  y={folderContextMenuY}
  item={folderContextMenuItem}
  context="drive"
  onOpen={handleFolderOpen}
  onRename={handleFolderRename}
  onMove={handleFolderMove}
  onToggleStar={handleFolderToggleStar}
  onTrash={handleFolderTrash}
/>

<!-- Modals -->
<CreateFolderModal
  bind:open={createFolderOpen}
  parentFolderId={folderId}
  onDone={handleDone}
/>

<RenameModal
  bind:open={renameOpen}
  item={renameItem}
  itemType={renameType}
  onDone={handleDone}
/>

<MoveModal
  bind:open={moveOpen}
  item={moveItem}
  itemType={moveType}
  onDone={handleDone}
/>

<DeleteConfirmModal
  bind:open={deleteConfirmOpen}
  item={deleteConfirmItem}
  itemType={deleteConfirmType}
  onDone={handleDone}
/>
