<script lang="ts">
  import type { FileItem, FolderItem } from "$lib/types";
  import { view } from "$lib/stores/view.svelte";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
  import FileContextMenu from "$lib/components/drive/FileContextMenu.svelte";
  import FolderContextMenu from "$lib/components/drive/FolderContextMenu.svelte";
  import RenameModal from "$lib/components/modals/RenameModal.svelte";
  import MoveModal from "$lib/components/modals/MoveModal.svelte";
  import DeleteConfirmModal from "$lib/components/modals/DeleteConfirmModal.svelte";
  import LoaderIcon from "@lucide/svelte/icons/loader";
  import { toast } from "svelte-sonner";
  import { goto } from "$app/navigation";

  let folders = $state<FolderItem[]>([]);
  let files = $state<FileItem[]>([]);
  let loading = $state(true);

  const breadcrumbs = [{ id: null, name: "Starred" }];

  const isEmpty = $derived(folders.length === 0 && files.length === 0 && !loading);

  async function fetchStarred() {
    loading = true;
    try {
      const [folderRes, fileRes] = await Promise.all([
        foldersApi.listStarredFolders(),
        filesApi.listStarred(),
      ]);
      folders = folderRes ?? [];
      files = fileRes ?? [];
    } catch {
      toast.error("Gagal memuat item yang ditandai");
      folders = [];
      files = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    fetchStarred();
  });

  function navigateToFolder(id: string | null) {
    if (id) {
      goto(`/?folder=${id}`);
    } else {
      goto("/");
    }
  }

  function openFile(_file: FileItem) {
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
  let renameOpen = $state(false);
  let renameItem = $state<FileItem | FolderItem | null>(null);
  let renameType = $state<"file" | "folder">("file");
  let moveOpen = $state(false);
  let moveItem = $state<FileItem | FolderItem | null>(null);
  let moveType = $state<"file" | "folder">("file");
  let deleteConfirmOpen = $state(false);
  let deleteConfirmItem = $state<FileItem | FolderItem | null>(null);
  let deleteConfirmType = $state<"file" | "folder">("file");

  // --- File Actions ---
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
      toast.success("Dihapus dari bintang");
      fetchStarred();
    } catch {
      toast.error("Gagal mengubah status bintang");
    }
  }

  function handleFileTrash(item: FileItem) {
    deleteConfirmItem = item;
    deleteConfirmType = "file";
    deleteConfirmOpen = true;
  }

  // --- Folder Actions ---
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
      toast.success("Dihapus dari bintang");
      fetchStarred();
    } catch {
      toast.error("Gagal mengubah status bintang");
    }
  }

  function handleFolderTrash(item: FolderItem) {
    deleteConfirmItem = item;
    deleteConfirmType = "folder";
    deleteConfirmOpen = true;
  }

  function handleDone() {
    fetchStarred();
  }
</script>

<svelte:head>
  <title>Starred — CriByte</title>
</svelte:head>

<Toolbar {breadcrumbs} onNavigate={navigateToFolder} />

<div class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto">
  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="flex flex-col items-center gap-3">
        <LoaderIcon class="size-6 text-primary animate-spin" />
        <span class="text-xs font-mono text-muted-foreground">Memuat item yang ditandai...</span>
      </div>
    </div>
  {:else if isEmpty}
    <EmptyState variant="starred" />
  {:else if view.viewMode === "grid"}
    <GridView
      {folders}
      {files}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {:else}
    <ListView
      {folders}
      {files}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {/if}
</div>

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
