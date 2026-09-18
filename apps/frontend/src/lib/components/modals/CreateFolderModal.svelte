<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import * as foldersApi from "$lib/api/folders";
  import { toast } from "svelte-sonner";
  import { handleApiError } from "$lib/api/errors";
  import FolderPlusIcon from "@lucide/svelte/icons/folder-plus";
  import LoaderIcon from "@lucide/svelte/icons/loader";

  let {
    open = $bindable(false),
    parentFolderId = null,
    onDone,
  }: {
    open: boolean;
    parentFolderId?: string | null;
    onDone?: () => void;
  } = $props();

  let name = $state("");
  let loading = $state(false);

  async function handleSubmit() {
    const trimmed = name.trim();
    if (!trimmed) return;

    loading = true;
    try {
      await foldersApi.createFolder({
        name: trimmed,
        parent_folder_id: parentFolderId,
      });
      toast.success("Folder dibuat");
      open = false;
      name = "";
      onDone?.();
    } catch (err) {
      handleApiError(err, "Gagal membuat folder");
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") handleSubmit();
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <FolderPlusIcon class="size-4" />
        Folder baru
      </Dialog.Title>
      <Dialog.Description>
        Masukkan nama folder yang ingin dibuat.
      </Dialog.Description>
    </Dialog.Header>

    <Input
      placeholder="Nama folder"
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
        {#if loading}<LoaderIcon class="size-3.5 mr-1.5 animate-spin" />Membuat...{:else}Buat{/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
