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
  import PermanentDeleteModal from "$lib/components/modals/PermanentDeleteModal.svelte";
  import LoaderIcon from "@lucide/svelte/icons/loader";
  import { toast } from "svelte-sonner";
  import { goto } from "$app/navigation";

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
    } catch {
      toast.error("Gagal memuat trash");
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
  let permanentDeleteOpen = $state(false);
  let permanentDeleteItem = $state<FileItem | FolderItem | null>(null);
  let permanentDeleteType = $state<"file" | "folder">("file");

  // --- File Actions ---
  async function handleFileRestore(item: FileItem) {
    try {
      await filesApi.restoreFile(item.id);
      toast.success("Berhasil dipulihkan");
      fetchTrash();
    } catch {
      toast.error("Gagal memulihkan file");
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
    } catch {
      toast.error("Gagal memulihkan folder");
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
  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="flex flex-col items-center gap-3">
        <LoaderIcon class="size-6 text-primary animate-spin" />
        <span class="text-xs font-mono text-muted-foreground">Memuat trash...</span>
      </div>
    </div>
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
