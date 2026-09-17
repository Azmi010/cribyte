<script lang="ts">
  import type { FileItem, FolderItem } from "$lib/types";
  import { view } from "$lib/stores/view.svelte";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import GridView from "$lib/components/drive/GridView.svelte";
  import ListView from "$lib/components/drive/ListView.svelte";
  import EmptyState from "$lib/components/drive/EmptyState.svelte";
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

  function openFile() {
    toast.info("Preview belum diimplementasi");
  }

  function handleFolderContextMenu(e: MouseEvent) {
    e.preventDefault();
  }

  function handleFileContextMenu(e: MouseEvent) {
    e.preventDefault();
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
      onOpenFile={openFile}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {:else}
    <ListView
      {folders}
      {files}
      onOpenFolder={navigateToFolder}
      onOpenFile={openFile}
      onFolderContextMenu={handleFolderContextMenu}
      onFileContextMenu={handleFileContextMenu}
    />
  {/if}
</div>

