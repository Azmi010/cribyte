<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import { toast } from "svelte-sonner";
  import { handleApiError } from "$lib/api/errors";
  import type { FileItem, FolderItem } from "$lib/types";
  import AlertTriangleIcon from "@lucide/svelte/icons/alert-triangle";
  import LoaderIcon from "@lucide/svelte/icons/loader";

  let {
    open = $bindable(false),
    item = null,
    itemType = "file",
    onDone,
  }: {
    open: boolean;
    item: FileItem | FolderItem | null;
    itemType?: "file" | "folder";
    onDone?: () => void;
  } = $props();

  let loading = $state(false);

  async function handlePermanentDelete() {
    if (!item) return;

    loading = true;
    try {
      if (itemType === "folder") {
        await foldersApi.permanentDeleteFolder(item.id);
      } else {
        await filesApi.permanentDeleteFile(item.id);
      }
      toast.success("Hapus permanen berhasil");
      open = false;
      onDone?.();
    } catch (err) {
      handleApiError(err, "Gagal menghapus permanen");
    } finally {
      loading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2 text-destructive">
        <AlertTriangleIcon class="size-4" />
        Hapus permanen?
      </Dialog.Title>
      <Dialog.Description>
        "{item?.name}" akan dihapus secara permanen. Tindakan ini tidak bisa dibatalkan.
      </Dialog.Description>
    </Dialog.Header>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)} disabled={loading}>
        Batal
      </Button>
      <Button variant="destructive" onclick={handlePermanentDelete} disabled={loading}>
        {#if loading}<LoaderIcon class="size-3.5 mr-1.5 animate-spin" />Menghapus...{:else}Hapus permanen{/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
