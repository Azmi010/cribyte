<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { SvelteSet } from "svelte/reactivity";
  import type { FileItem, FolderItem } from "$lib/types";
  import { formatFileSize } from "$lib/constants";
  import { view } from "$lib/stores/view.svelte";
  import { uploadStore } from "$lib/stores/upload.svelte";
  import { driveRefresh } from "$lib/stores/drive.svelte";
  import { searchQuery } from "$lib/stores/search";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
  import DriveSkeleton from "$lib/components/drive/DriveSkeleton.svelte";
  import FileContextMenu from "$lib/components/drive/FileContextMenu.svelte";
  import FolderContextMenu from "$lib/components/drive/FolderContextMenu.svelte";
  import CreateFolderModal from "$lib/components/modals/CreateFolderModal.svelte";
  import RenameModal from "$lib/components/modals/RenameModal.svelte";
  import MoveModal from "$lib/components/modals/MoveModal.svelte";
  import DeleteConfirmModal from "$lib/components/modals/DeleteConfirmModal.svelte";
  import FilePreviewModal from "$lib/components/modals/FilePreviewModal.svelte";
  import CloudUploadIcon from "@lucide/svelte/icons/cloud-upload";
  import RefreshCwIcon from "@lucide/svelte/icons/refresh-cw";
  import CheckIcon from "@lucide/svelte/icons/check";
  import CircleAlertIcon from "@lucide/svelte/icons/circle-alert";
  import XIcon from "@lucide/svelte/icons/x";
  import StarIcon from "@lucide/svelte/icons/star";
  import FolderInputIcon from "@lucide/svelte/icons/folder-input";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import { toast } from "svelte-sonner";
  import { handleApiError } from "$lib/api/errors";

  const folderId = $derived(page.url.searchParams.get("folder") ?? null);

  let folders = $state<FolderItem[]>([]);
  let files = $state<FileItem[]>([]);
  let breadcrumbs = $state<{ id: string | null; name: string }[]>([{ id: null, name: "Home" }]);
  let loading = $state(true);
  let dragOver = $state(false);
  let currentSearch = $state("");

  const isSearching = $derived(currentSearch.length > 0);

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

  // --- Selection (multi-item) ---
  const selectedIds = new SvelteSet<string>();
  let lastSelectedId = $state<string | null>(null);
  let uploadInputRef = $state<HTMLInputElement | null>(null);

  const orderedIds = $derived([
    ...sortedFolders().map((f) => f.id),
    ...sortedFiles().map((f) => f.id),
  ]);

  const singleSelectedId = $derived(selectedIds.size === 1 ? [...selectedIds][0] : null);
  const selectedFile = $derived(singleSelectedId ? (files.find((f) => f.id === singleSelectedId) ?? null) : null);
  const selectedFolder = $derived(singleSelectedId ? (folders.find((f) => f.id === singleSelectedId) ?? null) : null);

  const selectedFilesList = $derived(files.filter((f) => selectedIds.has(f.id)));
  const selectedFoldersList = $derived(folders.filter((f) => selectedIds.has(f.id)));
  const selectedEntries = $derived([
    ...selectedFoldersList.map((f) => ({ id: f.id, name: f.name, type: "folder" as const })),
    ...selectedFilesList.map((f) => ({ id: f.id, name: f.name, type: "file" as const })),
  ]);

  function clearSelection() {
    selectedIds.clear();
    lastSelectedId = null;
  }

  function replaceSelection(ids: Iterable<string>) {
    selectedIds.clear();
    for (const id of ids) selectedIds.add(id);
  }

  function handleSelect(id: string, e: MouseEvent) {
    if (e.shiftKey && lastSelectedId) {
      const a = orderedIds.indexOf(lastSelectedId);
      const b = orderedIds.indexOf(id);
      if (a !== -1 && b !== -1) {
        const [lo, hi] = a < b ? [a, b] : [b, a];
        for (let i = lo; i <= hi; i++) selectedIds.add(orderedIds[i]);
      }
      return;
    }
    if (e.ctrlKey || e.metaKey) {
      if (selectedIds.has(id)) selectedIds.delete(id);
      else selectedIds.add(id);
      lastSelectedId = id;
      return;
    }
    // klik biasa: kalau cuma item ini yang terpilih, batalkan; selain itu pilih tunggal
    if (selectedIds.size === 1 && selectedIds.has(id)) {
      clearSelection();
      return;
    }
    replaceSelection([id]);
    lastSelectedId = id;
  }

  // --- Bulk actions ---
  let bulkEntries = $state<{ id: string; name: string; type: "file" | "folder" }[]>([]);

  async function handleBulkStar() {
    const entries = selectedEntries;
    if (entries.length === 0) return;
    const allStarred = [...selectedFoldersList, ...selectedFilesList].every((i) => i.starred);
    const target = !allStarred;
    try {
      await Promise.all(
        entries.map((e) =>
          e.type === "folder"
            ? foldersApi.toggleFolderStarred(e.id, target)
            : filesApi.toggleFileStarred(e.id, target),
        ),
      );
      toast.success(target ? "Ditandai bintang" : "Dihapus dari bintang");
      clearSelection();
      fetchContents(folderId);
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
    }
  }

  function handleBulkMove() {
    if (selectedEntries.length === 0) return;
    bulkEntries = selectedEntries;
    moveItem = null;
    moveOpen = true;
  }

  function handleBulkTrash() {
    if (selectedEntries.length === 0) return;
    bulkEntries = selectedEntries;
    deleteConfirmItem = null;
    deleteConfirmOpen = true;
  }

  // --- Marquee (rubber-band) selection ---
  let contentEl = $state<HTMLDivElement | null>(null);
  let marquee = $state<{ x: number; y: number; w: number; h: number } | null>(null);
  let marqueeStart = { x: 0, y: 0 };
  let marqueeAdditive = false;
  let marqueeBase: string[] = [];

  function handleMarqueeStart(e: MouseEvent) {
    if (e.button !== 0 || !contentEl) return;
    const target = e.target as HTMLElement;
    // Abaikan kalau mulai dari item/tombol; biar klik item normal jalan
    if (target.closest("[data-select-id]") || target.closest("button")) return;

    marqueeAdditive = e.ctrlKey || e.metaKey || e.shiftKey;
    marqueeBase = marqueeAdditive ? [...selectedIds] : [];
    if (!marqueeAdditive) clearSelection();

    const rect = contentEl.getBoundingClientRect();
    marqueeStart = {
      x: e.clientX - rect.left + contentEl.scrollLeft,
      y: e.clientY - rect.top + contentEl.scrollTop,
    };
    marquee = { x: marqueeStart.x, y: marqueeStart.y, w: 0, h: 0 };

    window.addEventListener("mousemove", handleMarqueeMove);
    window.addEventListener("mouseup", handleMarqueeEnd);
  }

  function handleMarqueeMove(e: MouseEvent) {
    if (!contentEl || !marquee) return;
    const rect = contentEl.getBoundingClientRect();
    const curX = e.clientX - rect.left + contentEl.scrollLeft;
    const curY = e.clientY - rect.top + contentEl.scrollTop;
    const x = Math.min(marqueeStart.x, curX);
    const y = Math.min(marqueeStart.y, curY);
    const w = Math.abs(curX - marqueeStart.x);
    const h = Math.abs(curY - marqueeStart.y);
    marquee = { x, y, w, h };

    // Hitung item yang beririsan dengan kotak
    const boxLeft = x + rect.left - contentEl.scrollLeft;
    const boxTop = y + rect.top - contentEl.scrollTop;
    const boxRight = boxLeft + w;
    const boxBottom = boxTop + h;

    // Set lokal non-reaktif untuk hitung irisan
    // eslint-disable-next-line svelte/prefer-svelte-reactivity
    const next = new Set(marqueeBase);
    const nodes = contentEl.querySelectorAll<HTMLElement>("[data-select-id]");
    nodes.forEach((node) => {
      const r = node.getBoundingClientRect();
      const intersects =
        r.left < boxRight && r.right > boxLeft && r.top < boxBottom && r.bottom > boxTop;
      const id = node.dataset.selectId;
      if (id && intersects) next.add(id);
    });
    for (const id of [...selectedIds]) {
      if (!next.has(id)) selectedIds.delete(id);
    }
    for (const id of next) selectedIds.add(id);
  }

  function handleMarqueeEnd() {
    marquee = null;
    if (selectedIds.size > 0) {
      lastSelectedId = [...selectedIds][selectedIds.size - 1] ?? null;
    }
    window.removeEventListener("mousemove", handleMarqueeMove);
    window.removeEventListener("mouseup", handleMarqueeEnd);
  }

  async function fetchContents(currentFolderId: string | null, search?: string) {
    loading = true;
    try {
      if (search) {
        const searchRes = await filesApi.searchFiles(search);
        files = searchRes ?? [];
        folders = [];
        breadcrumbs = [
          { id: null, name: "Home" },
          { id: null, name: `Hasil pencarian "${search}"` },
        ];
      } else {
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
      }
    } catch (err) {
      handleApiError(err, "Gagal memuat konten");
      folders = [];
      files = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    fetchContents(folderId);
  });

  $effect(() => {
    const fn = () => fetchContents(folderId);
    driveRefresh.register(fn, folderId);
    return () => driveRefresh.unregister(fn);
  });

  $effect(() => {
    const unsub = searchQuery.subscribe((q) => {
      currentSearch = q;
      if (q) {
        fetchContents(null, q);
      } else {
        fetchContents(folderId);
      }
    });
    return unsub;
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
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
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
    } catch (err) {
      handleApiError(err, "Gagal mengubah status bintang");
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
      uploadStore.addUpload(droppedFiles[i], folderId, () => fetchContents(folderId));
    }
    toast.success(`${droppedFiles.length} file ditambahkan ke antrian upload`);
  }

  function handleDone() {
    clearSelection();
    bulkEntries = [];
    fetchContents(folderId);
  }

  // --- Keyboard Shortcuts ---
  const anyModalOpen = $derived(
    createFolderOpen || renameOpen || moveOpen || deleteConfirmOpen || previewOpen,
  );

  function handleShortcut(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null;
    const typing =
      target &&
      (target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable);

    // Escape: batal seleksi (modal punya handler sendiri)
    if (e.key === "Escape") {
      if (!anyModalOpen) clearSelection();
      return;
    }

    // Ctrl/Cmd + U: trigger upload
    if ((e.metaKey || e.ctrlKey) && (e.key === "u" || e.key === "U")) {
      e.preventDefault();
      uploadInputRef?.click();
      return;
    }

    if (typing || anyModalOpen) return;

    // Ctrl/Cmd + A: pilih semua item
    if ((e.metaKey || e.ctrlKey) && (e.key === "a" || e.key === "A")) {
      if (orderedIds.length > 0) {
        e.preventDefault();
        replaceSelection(orderedIds);
        lastSelectedId = orderedIds[orderedIds.length - 1];
      }
      return;
    }

    if (selectedIds.size === 0) return;

    // Delete: trash item terpilih
    if (e.key === "Delete" || e.key === "Backspace") {
      e.preventDefault();
      if (selectedIds.size > 1) {
        handleBulkTrash();
      } else if (selectedFolder) {
        handleFolderTrash(selectedFolder);
      } else if (selectedFile) {
        handleFileTrash(selectedFile);
      }
      return;
    }

    // F2: rename item terpilih (hanya saat satu item)
    if (e.key === "F2") {
      e.preventDefault();
      if (selectedFolder) handleFolderRename(selectedFolder);
      else if (selectedFile) handleFileRename(selectedFile);
    }
  }

  function handleUploadInput(e: Event) {
    const input = e.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;
    const count = input.files.length;
    for (let i = 0; i < input.files.length; i++) {
      uploadStore.addUpload(input.files[i], folderId, () => fetchContents(folderId));
    }
    input.value = "";
    toast.success(`${count} file ditambahkan ke antrian upload`);
  }
</script>

<svelte:window onkeydown={handleShortcut} />

<input
  bind:this={uploadInputRef}
  type="file"
  multiple
  class="hidden"
  onchange={handleUploadInput}
/>

<Toolbar
  {breadcrumbs}
  onNavigate={navigateToFolder}
  onNewFolder={handleNewFolder}
/>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  bind:this={contentEl}
  class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto relative"
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  onmousedown={handleMarqueeStart}
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

  {#if marquee}
    <div
      class="absolute z-20 border border-primary/70 bg-primary/10 pointer-events-none rounded-sm"
      style="left: {marquee.x}px; top: {marquee.y}px; width: {marquee.w}px; height: {marquee.h}px;"
    ></div>
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
      ⌘/Ctrl + U
    </div>
  </div>

  {#if loading}
    <DriveSkeleton />
  {:else if isEmpty}
    <EmptyState variant={isSearching ? "search" : "folder"} query={currentSearch} />
  {:else if view.viewMode === "grid"}
    <GridView
      folders={sortedFolders()}
      files={sortedFiles()}
      {selectedIds}
      onSelect={handleSelect}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {:else}
    <ListView
      folders={sortedFolders()}
      files={sortedFiles()}
      {selectedIds}
      onSelect={handleSelect}
      onOpenFolder={navigateToFolder}
      onOpenFile={(id) => { const f = files.find((x) => x.id === id); if (f) openFile(f); }}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {/if}
</div>

<!-- Floating Upload Progress Bar Widget -->
{#if uploadStore.hasAny}
  {@const activeCount = uploadStore.uploads.filter((u) => u.status === "uploading" || u.status === "pending").length}
  {@const doneCount = uploadStore.uploads.filter((u) => u.status === "done").length}
  <div class="fixed bottom-4 right-4 z-50 w-80 rounded-lg border border-border/90 bg-card shadow-xl animate-in fade-in slide-in-from-bottom-3 duration-200 overflow-hidden">
    <div class="flex items-center justify-between px-3.5 py-2.5 border-b border-border/70">
      <div class="text-xs font-semibold text-foreground">
        {#if activeCount > 0}
          Uploading {activeCount} file(s)
        {:else}
          Upload selesai — {doneCount} file
        {/if}
      </div>
      <div class="flex items-center gap-1">
        {#if activeCount === 0}
          <button
            type="button"
            onclick={() => uploadStore.clearCompleted()}
            class="text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
            aria-label="Tutup"
          >
            <XIcon class="size-3.5" />
          </button>
        {/if}
      </div>
    </div>
    <div class="max-h-64 overflow-y-auto divide-y divide-border/50">
      {#each uploadStore.uploads as u (u.id)}
        <div class="px-3.5 py-2.5 space-y-2">
          <div class="flex items-center justify-between text-xs gap-2">
            <div class="flex items-center gap-2 font-medium text-foreground truncate min-w-0">
              {#if u.status === "done"}
                <CheckIcon class="size-3.5 text-green-500 shrink-0" />
              {:else if u.status === "error"}
                <CircleAlertIcon class="size-3.5 text-destructive shrink-0" />
              {:else}
                <RefreshCwIcon class="size-3.5 text-primary animate-spin shrink-0" />
              {/if}
              <span class="truncate">{u.file.name}</span>
            </div>
            <button
              type="button"
              onclick={() => uploadStore.removeUpload(u.id)}
              class="text-muted-foreground hover:text-foreground cursor-pointer p-0.5 shrink-0"
              aria-label={u.status === "uploading" || u.status === "pending" ? "Batalkan upload" : "Hapus dari daftar"}
            >
              <XIcon class="size-3.5" />
            </button>
          </div>
          {#if u.status === "uploading" || u.status === "pending"}
            <div class="h-1.5 w-full bg-muted rounded-full overflow-hidden">
              <div
                class="h-full bg-primary rounded-full transition-all duration-300"
                style="width: {u.progress}%"
              ></div>
            </div>
            <div class="text-[11px] font-mono text-muted-foreground">
              {u.progress}% of {formatFileSize(u.file.size)}
            </div>
          {:else if u.status === "done"}
            <div class="text-[11px] font-mono text-green-500">
              Berhasil · {formatFileSize(u.file.size)}
            </div>
          {:else}
            <div class="text-[11px] font-mono text-destructive truncate">
              {u.error ?? "Upload gagal"}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>
{/if}

<!-- Bulk Selection Action Bar -->
{#if selectedIds.size > 1}
  <div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-40 flex items-center gap-1 rounded-full border border-border/90 bg-card shadow-xl px-2 py-1.5 animate-in fade-in slide-in-from-bottom-3 duration-200">
    <span class="text-xs font-medium text-foreground px-2.5">{selectedIds.size} dipilih</span>
    <div class="w-px h-5 bg-border/70"></div>
    <button
      type="button"
      onclick={handleBulkStar}
      class="flex items-center gap-1.5 text-xs font-medium px-2.5 py-1.5 rounded-full hover:bg-muted transition-colors cursor-pointer"
    >
      <StarIcon class="size-3.5" />
      <span>Star</span>
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
      onclick={clearSelection}
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
