<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import { toast } from "svelte-sonner";
  import type { FileItem, FolderItem } from "$lib/types";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";

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

  async function handleTrash() {
    if (!item) return;

    loading = true;
    try {
      if (itemType === "folder") {
        await foldersApi.trashFolder(item.id);
      } else {
        await filesApi.trashFile(item.id);
      }
      toast.success("Dipindahkan ke trash");
      open = false;
      onDone?.();
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Gagal menghapus";
      toast.error(msg);
    } finally {
      loading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <Trash2Icon class="size-4 text-destructive" />
        Trash {itemType === "folder" ? "folder" : "file"}?
      </Dialog.Title>
      <Dialog.Description>
        "{item?.name}" akan dipindahkan ke trash. Anda bisa memulihkannya nanti.
      </Dialog.Description>
    </Dialog.Header>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)} disabled={loading}>
        Batal
      </Button>
      <Button variant="destructive" onclick={handleTrash} disabled={loading}>
        {loading ? "Menghapus..." : "Trash"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
