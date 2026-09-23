<script lang="ts">
  import type { FileItem, FolderItem } from "$lib/types";
  import { view } from "$lib/stores/view.svelte";
  import { createSelection, createMarquee } from "$lib/stores/selection.svelte";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
  import DriveSkeleton from "$lib/components/drive/DriveSkeleton.svelte";
  import FileContextMenu from "$lib/components/drive/FileContextMenu.svelte";
  import FolderContextMenu from "$lib/components/drive/FolderContextMenu.svelte";
  import RenameModal from "$lib/components/modals/RenameModal.svelte";
  import MoveModal from "$lib/components/modals/MoveModal.svelte";
  import DeleteConfirmModal from "$lib/components/modals/DeleteConfirmModal.svelte";
  import FilePreviewModal from "$lib/components/modals/FilePreviewModal.svelte";
  import StarOffIcon from "@lucide/svelte/icons/star-off";
  import FolderInputIcon from "@lucide/svelte/icons/folder-input";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import XIcon from "@lucide/svelte/icons/x";
  import { toast } from "svelte-sonner";
  import { goto } from "$app/navigation";
  import { handleApiError } from "$lib/api/errors";

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
    } catch (err) {
      handleApiError(err, "Gagal memuat item yang ditandai");
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

  let previewOpen = $state(false);
  let previewItem = $state<FileItem | null>(null);

  function openFile(file: FileItem) {
    previewItem = file;
    previewOpen = true;
  }

  // --- Selection ---
  const sel = createSelection(() => folders, () => files);
  let contentEl = $state<HTMLDivElement | null>(null);
  const mq = createMarquee(
    () => contentEl,
    sel.selectedIds,
    (id) => (sel.lastSelectedId = id),
    () => sel.clear(),
  );

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
  let renameOpen = $state(false);
  let renameItem = $state<FileItem | FolderItem | null>(null);
  let renameType = $state<"file" | "folder">("file");
  let moveOpen = $state(false);
  let moveItem = $state<FileItem | FolderItem | null>(null);
  let moveType = $state<"file" | "folder">("file");
  let deleteConfirmOpen = $state(false);
  let deleteConfirmItem = $state<FileItem | FolderItem | null>(null);
  let deleteConfirmType = $state<"file" | "folder">("file");
  let bulkEntries = $state<{ id: string; name: string; type: "file" | "folder" }[]>([]);

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
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
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
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
    }
  }

  function handleFolderTrash(item: FolderItem) {
    deleteConfirmItem = item;
    deleteConfirmType = "folder";
    deleteConfirmOpen = true;
  }

  // --- Bulk Actions ---
  async function handleBulkUnstar() {
    const entries = sel.selectedEntries;
    if (entries.length === 0) return;
    try {
      await Promise.all(
        entries.map((e) =>
          e.type === "folder"
            ? foldersApi.toggleFolderStarred(e.id, false)
            : filesApi.toggleFileStarred(e.id, false),
        ),
      );
      toast.success(`${entries.length} item dihapus dari bintang`);
      sel.clear();
      fetchStarred();
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
    }
  }

  function handleBulkMove() {
    if (sel.selectedEntries.length === 0) return;
    bulkEntries = sel.selectedEntries;
    moveItem = null;
    moveOpen = true;
  }

  function handleBulkTrash() {
    if (sel.selectedEntries.length === 0) return;
    bulkEntries = sel.selectedEntries;
    deleteConfirmItem = null;
    deleteConfirmOpen = true;
  }

  function handleDone() {
    sel.clear();
    bulkEntries = [];
    fetchStarred();
  }

  // --- Keyboard Shortcuts ---
  const anyModalOpen = $derived(renameOpen || moveOpen || deleteConfirmOpen || previewOpen);

  function handleShortcut(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null;
    const typing =
      target &&
      (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable);

    if (e.key === "Escape") {
      if (!anyModalOpen) sel.clear();
      return;
    }
    if (typing || anyModalOpen) return;

    if ((e.metaKey || e.ctrlKey) && (e.key === "a" || e.key === "A")) {
      if (sel.orderedIds.length > 0) {
        e.preventDefault();
        sel.selectAll();
      }
    }
  }
</script>

<svelte:head>
  <title>Starred — CriByte</title>
</svelte:head>

<svelte:window onkeydown={handleShortcut} />

<Toolbar {breadcrumbs} onNavigate={navigateToFolder} />

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  bind:this={contentEl}
  class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto relative"
  onmousedown={mq.onStart}
  role="region"
  aria-label="Starred browser"
>
  {#if mq.marquee}
    <div
      class="absolute z-20 border border-primary/70 bg-primary/10 pointer-events-none rounded-sm"
      style="left: {mq.marquee.x}px; top: {mq.marquee.y}px; width: {mq.marquee.w}px; height: {mq.marquee.h}px;"
    ></div>
  {/if}

  {#if loading}
    <DriveSkeleton />
  {:else if isEmpty}
    <EmptyState variant="starred" />
  {:else if view.viewMode === "grid"}
    <GridView
      {folders}
      {files}
      selectedIds={sel.selectedIds}
      onSelect={sel.handleSelect}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {:else}
    <ListView
      {folders}
      {files}
      selectedIds={sel.selectedIds}
      onSelect={sel.handleSelect}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {/if}
</div>

<!-- Bulk Selection Action Bar -->
{#if sel.selectedIds.size > 1}
  <div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-40 flex items-center gap-1 rounded-full border border-border/90 bg-card shadow-xl px-2 py-1.5 animate-in fade-in slide-in-from-bottom-3 duration-200">
    <span class="text-xs font-medium text-foreground px-2.5">{sel.selectedIds.size} dipilih</span>
    <div class="w-px h-5 bg-border/70"></div>
    <button
      type="button"
      onclick={handleBulkUnstar}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full hover:bg-muted transition-colors cursor-pointer"
    >
      <StarOffIcon class="size-3.5" />
      <span>Unstar</span>
    </button>
    <button
      type="button"
      onclick={handleBulkMove}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full hover:bg-muted transition-colors cursor-pointer"
    >
      <FolderInputIcon class="size-3.5" />
      <span>Move</span>
    </button>
    <button
      type="button"
      onclick={handleBulkTrash}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
    >
      <Trash2Icon class="size-3.5" />
      <span>Trash</span>
    </button>
    <div class="w-px h-5 bg-border/70"></div>
    <button
      type="button"
      onclick={() => sel.clear()}
      class="p-1.5 rounded-full text-muted-foreground hover:bg-muted hover:text-foreground transition-colors cursor-pointer"
      aria-label="Batal pilih"
    >
      <XIcon class="size-3.5" />
    </button>
  </div>
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
  entries={moveItem ? null : bulkEntries}
  onDone={handleDone}
/>

<DeleteConfirmModal
  bind:open={deleteConfirmOpen}
  item={deleteConfirmItem}
  itemType={deleteConfirmType}
  entries={deleteConfirmItem ? null : bulkEntries}
  onDone={handleDone}
/>

<FilePreviewModal
  bind:open={previewOpen}
  item={previewItem}
/>
