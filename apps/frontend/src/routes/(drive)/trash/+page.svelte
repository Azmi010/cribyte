<script lang="ts">
  import type { FileItem, FolderItem } from "$lib/types";
  import { view } from "$lib/stores/view.svelte";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
  import DriveSkeleton from "$lib/components/drive/DriveSkeleton.svelte";
  import FileContextMenu from "$lib/components/drive/FileContextMenu.svelte";
  import FolderContextMenu from "$lib/components/drive/FolderContextMenu.svelte";
  import PermanentDeleteModal from "$lib/components/modals/PermanentDeleteModal.svelte";
  import EmptyTrashModal from "$lib/components/modals/EmptyTrashModal.svelte";
  import FilePreviewModal from "$lib/components/modals/FilePreviewModal.svelte";
  import TrashIcon from "@lucide/svelte/icons/trash";
  import { toast } from "svelte-sonner";
  import { goto } from "$app/navigation";
  import { handleApiError } from "$lib/api/errors";

  let folders = $state<FolderItem[]>([]);
  let files = $state<FileItem[]>([]);
  let loading = $state(true);

  const breadcrumbs = [{ id: null, name: "Trash" }];

  const isEmpty = $derived(folders.length === 0 && files.length === 0 && !loading);

  async function fetchTrash() {
    loading = true;
    try {
      const [folderRes, fileRes] = await Promise.all([
        foldersApi.listTrashFolders(),
        filesApi.listTrash(),
      ]);
      folders = folderRes ?? [];
      files = fileRes ?? [];
    } catch (err) {
      handleApiError(err, "Gagal memuat trash");
      folders = [];
      files = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    fetchTrash();
  });

  function navigateToFolder(id: string | null) {
    if (id) {
      goto(`/?folder=${id}`);
    } else {
      goto("/");
    }
  }

  let previewOpen = $state(false);
  let previewItem = $state<FileItem | null>(null);

  function openFile(file: FileItem) {
    previewItem = file;
    previewOpen = true;
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
    fileContextMenuOpen = false;
    folderContextMenuItem = folder;
    folderContextMenuX = e.clientX;
    folderContextMenuY = e.clientY;
    folderContextMenuOpen = true;
  }

  function handleFileContextMenu(e: MouseEvent, file: FileItem) {
    e.preventDefault();
    folderContextMenuOpen = false;
    fileContextMenuItem = file;
    fileContextMenuX = e.clientX;
    fileContextMenuY = e.clientY;
    fileContextMenuOpen = true;
  }

  // --- Modal State ---
  let permanentDeleteOpen = $state(false);
  let permanentDeleteItem = $state<FileItem | FolderItem | null>(null);
  let permanentDeleteType = $state<"file" | "folder">("file");
  let emptyTrashOpen = $state(false);

  // --- File Actions ---
  async function handleFileRestore(item: FileItem) {
    try {
      await filesApi.restoreFile(item.id);
      toast.success("Berhasil dipulihkan");
      fetchTrash();
    } catch (err) {
      handleApiError(err, "Gagal memulihkan file");
    }
  }

  function handleFilePermanentDelete(item: FileItem) {
    permanentDeleteItem = item;
    permanentDeleteType = "file";
    permanentDeleteOpen = true;
  }

  // --- Folder Actions ---
  async function handleFolderRestore(item: FolderItem) {
    try {
      await foldersApi.restoreFolder(item.id);
      toast.success("Berhasil dipulihkan");
      fetchTrash();
    } catch (err) {
      handleApiError(err, "Gagal memulihkan folder");
    }
  }

  function handleFolderPermanentDelete(item: FolderItem) {
    permanentDeleteItem = item;
    permanentDeleteType = "folder";
    permanentDeleteOpen = true;
  }

  function handleDone() {
    fetchTrash();
  }
</script>

<svelte:head>
  <title>Trash — CriByte</title>
</svelte:head>

<Toolbar {breadcrumbs} onNavigate={navigateToFolder} />

<div class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto">
  {#if !isEmpty && !loading}
    <div class="flex justify-end">
      <button
        type="button"
        onclick={() => (emptyTrashOpen = true)}
        class="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium text-destructive hover:bg-destructive/10 transition-colors"
      >
        <TrashIcon class="size-3.5" />
        Kosongkan trash
      </button>
    </div>
  {/if}
  {#if loading}
    <DriveSkeleton />
  {:else if isEmpty}
    <EmptyState variant="trash" />
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

<!-- Context Menus (trash context) -->
<FileContextMenu
  bind:open={fileContextMenuOpen}
  x={fileContextMenuX}
  y={fileContextMenuY}
  item={fileContextMenuItem}
  context="trash"
  onRestore={handleFileRestore}
  onPermanentDelete={handleFilePermanentDelete}
/>

<FolderContextMenu
  bind:open={folderContextMenuOpen}
  x={folderContextMenuX}
  y={folderContextMenuY}
  item={folderContextMenuItem}
  context="trash"
  onRestore={handleFolderRestore}
  onPermanentDelete={handleFolderPermanentDelete}
/>

<!-- Modals -->
<PermanentDeleteModal
  bind:open={permanentDeleteOpen}
  item={permanentDeleteItem}
  itemType={permanentDeleteType}
  onDone={handleDone}
/>

<FilePreviewModal
  bind:open={previewOpen}
  item={previewItem}
/>

<EmptyTrashModal
  bind:open={emptyTrashOpen}
  {files}
  {folders}
  onDone={handleDone}
/>
