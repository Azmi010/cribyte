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
  import PermanentDeleteModal from "$lib/components/modals/PermanentDeleteModal.svelte";
  import EmptyTrashModal from "$lib/components/modals/EmptyTrashModal.svelte";
  import FilePreviewModal from "$lib/components/modals/FilePreviewModal.svelte";
  import TrashIcon from "@lucide/svelte/icons/trash";
  import RotateCcwIcon from "@lucide/svelte/icons/rotate-ccw";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import XIcon from "@lucide/svelte/icons/x";
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

  // --- Selection ---
  const sel = createSelection(() => folders, () => files);
  let contentEl = $state<HTMLDivElement | null>(null);
  const mq = createMarquee(
    () => contentEl,
    () => sel.selectedIds,
    (v) => (sel.selectedIds = v),
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
  let permanentDeleteOpen = $state(false);
  let permanentDeleteItem = $state<FileItem | FolderItem | null>(null);
  let permanentDeleteType = $state<"file" | "folder">("file");
  let emptyTrashOpen = $state(false);
  let bulkEntries = $state<{ id: string; name: string; type: "file" | "folder" }[]>([]);

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

  // --- Bulk Actions ---
  async function handleBulkRestore() {
    const entries = sel.selectedEntries;
    if (entries.length === 0) return;
    try {
      await Promise.all(
        entries.map((e) =>
          e.type === "folder" ? foldersApi.restoreFolder(e.id) : filesApi.restoreFile(e.id),
        ),
      );
      toast.success(`${entries.length} item dipulihkan`);
      sel.clear();
      fetchTrash();
    } catch (err) {
      handleApiError(err, "Gagal memulihkan");
    }
  }

  function handleBulkPermanentDelete() {
    if (sel.selectedEntries.length === 0) return;
    bulkEntries = sel.selectedEntries;
    permanentDeleteItem = null;
    permanentDeleteOpen = true;
  }

  function handleDone() {
    sel.clear();
    bulkEntries = [];
    fetchTrash();
  }

  // --- Keyboard Shortcuts ---
  function handleShortcut(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null;
    const typing =
      target &&
      (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable);

    if (e.key === "Escape") {
      if (!permanentDeleteOpen && !emptyTrashOpen && !previewOpen) sel.clear();
      return;
    }
    if (typing || permanentDeleteOpen || emptyTrashOpen || previewOpen) return;

    if ((e.metaKey || e.ctrlKey) && (e.key === "a" || e.key === "A")) {
      if (sel.orderedIds.length > 0) {
        e.preventDefault();
        sel.selectAll();
      }
    }
  }
</script>

<svelte:head>
  <title>Trash — CriByte</title>
</svelte:head>

<svelte:window onkeydown={handleShortcut} />

<Toolbar {breadcrumbs} onNavigate={navigateToFolder} />

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  bind:this={contentEl}
  class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto relative"
  onmousedown={mq.onStart}
  role="region"
  aria-label="Trash browser"
>
  {#if mq.marquee}
    <div
      class="absolute z-20 border border-primary/70 bg-primary/10 pointer-events-none rounded-sm"
      style="left: {mq.marquee.x}px; top: {mq.marquee.y}px; width: {mq.marquee.w}px; height: {mq.marquee.h}px;"
    ></div>
  {/if}

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
      onclick={handleBulkRestore}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full hover:bg-muted transition-colors cursor-pointer"
    >
      <RotateCcwIcon class="size-3.5" />
      <span>Restore</span>
    </button>
    <button
      type="button"
      onclick={handleBulkPermanentDelete}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
    >
      <Trash2Icon class="size-3.5" />
      <span>Hapus permanen</span>
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
  entries={permanentDeleteItem ? null : bulkEntries}
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
