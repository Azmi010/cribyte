<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import * as filesApi from "$lib/api/files";
  import * as foldersApi from "$lib/api/folders";
  import { toast } from "svelte-sonner";
  import PencilIcon from "@lucide/svelte/icons/pencil";
  import type { FileItem, FolderItem } from "$lib/types";

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

  let name = $state("");
  let loading = $state(false);

  $effect(() => {
    if (item) {
      name = item.name;
    }
  });

  async function handleSubmit() {
    const trimmed = name.trim();
    if (!trimmed || !item) return;
    if (trimmed === item.name) {
      open = false;
      return;
    }

    loading = true;
    try {
      if (itemType === "folder") {
        await foldersApi.renameFolder(item.id, trimmed);
      } else {
        await filesApi.renameFile(item.id, trimmed);
      }
      toast.success("Berhasil di-rename");
      open = false;
      onDone?.();
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Gagal rename";
      toast.error(msg);
    } finally {
      loading = false;
    }
  }

  function handleOpen() {
    if (item) name = item.name;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") handleSubmit();
  }
</script>

<Dialog.Root bind:open onOpenChange={handleOpen}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <PencilIcon class="size-4" />
        Rename {itemType === "folder" ? "folder" : "file"}
      </Dialog.Title>
      <Dialog.Description>
        Masukkan nama baru untuk {itemType === "folder" ? "folder" : "file"} ini.
      </Dialog.Description>
    </Dialog.Header>

    <Input
      placeholder="Nama baru"
      bind:value={name}
      onkeydown={handleKeydown}
      disabled={loading}
      autofocus
    />

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)} disabled={loading}>
        Batal
      </Button>
      <Button onclick={handleSubmit} disabled={!name.trim() || loading}>
        {loading ? "Menyimpan..." : "Simpan"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
