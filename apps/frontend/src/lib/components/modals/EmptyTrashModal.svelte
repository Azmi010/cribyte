<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import { toast } from "svelte-sonner";
  import type { FileItem, FolderItem } from "$lib/types";
  import AlertTriangleIcon from "@lucide/svelte/icons/alert-triangle";

  let {
    open = $bindable(false),
    files = [],
    folders = [],
    onDone,
  }: {
    open: boolean;
    files: FileItem[];
    folders: FolderItem[];
    onDone?: () => void;
  } = $props();

  let loading = $state(false);

  async function handleEmptyTrash() {
    loading = true;
    let deleted = 0;
    let failed = 0;

    for (const file of files) {
      try {
        await filesApi.permanentDeleteFile(file.id);
        deleted++;
      } catch {
        failed++;
      }
    }

    for (const folder of folders) {
      try {
        await foldersApi.permanentDeleteFolder(folder.id);
        deleted++;
      } catch {
        failed++;
      }
    }

    if (failed > 0) {
      toast.error(`${failed} item gagal dihapus`);
    }
    if (deleted > 0) {
      toast.success(`${deleted} item dihapus permanen`);
    }

    loading = false;
    open = false;
    onDone?.();
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2 text-destructive">
        <AlertTriangleIcon class="size-4" />
        Kosongkan trash?
      </Dialog.Title>
      <Dialog.Description>
        Semua {files.length + folders.length} item di trash akan dihapus secara permanen. Tindakan ini tidak bisa dibatalkan.
      </Dialog.Description>
    </Dialog.Header>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)} disabled={loading}>
        Batal
      </Button>
      <Button variant="destructive" onclick={handleEmptyTrash} disabled={loading}>
        {loading ? "Menghapus..." : "Hapus permanen semua"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
