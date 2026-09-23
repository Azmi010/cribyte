<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import * as foldersApi from "$lib/api/folders";
  import * as filesApi from "$lib/api/files";
  import { toast } from "svelte-sonner";
  import { handleApiError } from "$lib/api/errors";
  import type { FileItem, FolderItem } from "$lib/types";
  import FolderIcon from "@lucide/svelte/icons/folder";

  import ChevronRightIcon from "@lucide/svelte/icons/chevron-right";
  import HomeIcon from "@lucide/svelte/icons/home";
  import LoaderIcon from "@lucide/svelte/icons/loader";

  let {
    open = $bindable(false),
    item = null,
    itemType = "file",
    entries = null,
    onDone,
  }: {
    open: boolean;
    item: FileItem | FolderItem | null;
    itemType?: "file" | "folder";
    entries?: { id: string; name: string; type: "file" | "folder" }[] | null;
    onDone?: () => void;
  } = $props();

  const isBulk = $derived(!!entries && entries.length > 0);
  const bulkFolderIds = $derived(new Set((entries ?? []).filter((e) => e.type === "folder").map((e) => e.id)));

  let loading = $state(false);
  let moving = $state(false);
  let subfolders = $state<FolderItem[]>([]);

  interface PathSegment {
    id: string | null;
    name: string;
  }

  let pathStack = $state<PathSegment[]>([{ id: null, name: "Home" }]);
  let currentParentId = $derived(pathStack[pathStack.length - 1].id);

  async function loadFolders(parentId: string | null) {
    loading = true;
    try {
      subfolders = (await foldersApi.listFolders(parentId)) ?? [];
    } catch {
      subfolders = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    if (open) {
      pathStack = [{ id: null, name: "Home" }];
      loadFolders(null);
    }
  });

  function navigateInto(folder: FolderItem) {
    pathStack = [...pathStack, { id: folder.id, name: folder.name }];
    loadFolders(folder.id);
  }

  function navigateTo(index: number) {
    pathStack = pathStack.slice(0, index + 1);
    loadFolders(pathStack[pathStack.length - 1].id);
  }

  async function handleMove() {
    if (isBulk) {
      moving = true;
      try {
        await Promise.all(
          (entries ?? []).map((e) =>
            e.type === "folder"
              ? foldersApi.moveFolder(e.id, currentParentId)
              : filesApi.moveFile(e.id, currentParentId),
          ),
        );
        toast.success(`${entries!.length} item dipindahkan`);
        open = false;
        onDone?.();
      } catch (err) {
        handleApiError(err, "Gagal memindahkan");
      } finally {
        moving = false;
      }
      return;
    }

    if (!item) return;

    moving = true;
    try {
      if (itemType === "folder") {
        await foldersApi.moveFolder(item.id, currentParentId);
      } else {
        await filesApi.moveFile(item.id, currentParentId);
      }
      toast.success("Berhasil dipindahkan");
      open = false;
      onDone?.();
    } catch (err) {
      handleApiError(err, "Gagal memindahkan");
    } finally {
      moving = false;
    }
  }

  function isCurrentFolderDisabled(): boolean {
    if (isBulk) {
      // Tidak boleh pindah folder ke dirinya sendiri
      return currentParentId !== null && bulkFolderIds.has(currentParentId);
    }
    if (!item || itemType !== "folder") return false;
    return currentParentId === item.id;
  }

  const selectedFolderName = $derived(pathStack[pathStack.length - 1].name);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-lg">
    <Dialog.Header>
      <Dialog.Title>Pindahkan {isBulk ? `${entries!.length} item` : itemType === "folder" ? "folder" : "file"}</Dialog.Title>
      <Dialog.Description>
        {#if isBulk}
          Pilih folder tujuan untuk memindahkan {entries!.length} item.
        {:else}
          Pilih folder tujuan untuk memindahkan "{item?.name}".
        {/if}
      </Dialog.Description>
    </Dialog.Header>

    <div class="flex flex-col gap-3">
      <!-- Breadcrumb path -->
      <div class="flex items-center gap-1 text-xs text-muted-foreground overflow-x-auto py-1">
        {#each pathStack as segment, index (segment.id ?? "root")}
          {#if index > 0}
            <ChevronRightIcon class="size-3 shrink-0" />
          {/if}
          <button
            type="button"
            onclick={() => navigateTo(index)}
            class="flex items-center gap-1 px-1.5 py-0.5 rounded hover:bg-muted transition-colors shrink-0 {index === pathStack.length - 1 ? 'font-medium text-foreground' : ''}"
          >
            {#if index === 0}
              <HomeIcon class="size-3" />
            {/if}
            <span>{segment.name}</span>
          </button>
        {/each}
      </div>

      <!-- Folder list -->
      <div class="border rounded-md bg-muted/30 max-h-60 overflow-y-auto">
        {#if loading}
          <div class="flex items-center justify-center py-8">
            <LoaderIcon class="size-5 text-primary animate-spin" />
          </div>
        {:else if subfolders.length === 0}
          <div class="text-center py-8 text-xs text-muted-foreground">
            Tidak ada subfolder
          </div>
        {:else}
          {#each subfolders as folder (folder.id)}
            <button
              type="button"
              onclick={() => navigateInto(folder)}
              class="flex items-center gap-2 w-full px-3 py-2 text-left text-sm hover:bg-muted/60 transition-colors border-b border-border/30 last:border-0 cursor-pointer"
            >
              <FolderIcon class="size-4 text-primary shrink-0" />
              <span class="flex-1 truncate">{folder.name}</span>
              <ChevronRightIcon class="size-3.5 text-muted-foreground shrink-0" />
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)} disabled={moving}>
        Batal
      </Button>
      <Button
        onclick={handleMove}
        disabled={moving || isCurrentFolderDisabled()}
      >
        {#if moving}<LoaderIcon class="size-3.5 mr-1.5 animate-spin" />Memindahkan...{:else}Pindah ke {selectedFolderName}{/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
