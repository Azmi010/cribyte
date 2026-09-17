<script lang="ts">
  import { page } from "$app/state";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import FolderIcon from "@lucide/svelte/icons/folder";
  import StarIcon from "@lucide/svelte/icons/star";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import SettingsIcon from "@lucide/svelte/icons/settings";
  import HelpCircleIcon from "@lucide/svelte/icons/help-circle";
  import XIcon from "@lucide/svelte/icons/x";
  import { uploadStore } from "$lib/stores/upload.svelte";

  let {
    isMobile = false,
    onClose,
  }: {
    isMobile?: boolean;
    onClose?: () => void;
  } = $props();

  let fileInputRef: HTMLInputElement | null = $state(null);

  const currentPath = $derived(page.url.pathname);

  const navItems = [
    { label: "Files", href: "/", icon: FolderIcon, match: (p: string) => p === "/" || p.startsWith("/folder") },
    { label: "Starred", href: "/starred", icon: StarIcon, match: (p: string) => p.startsWith("/starred") },
    { label: "Trash", href: "/trash", icon: Trash2Icon, match: (p: string) => p.startsWith("/trash") },
  ];

  function handleUploadClick() {
    fileInputRef?.click();
  }

  function handleFilesSelected(e: Event) {
    const input = e.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;

    for (let i = 0; i < input.files.length; i++) {
      const file = input.files[i];
      uploadStore.addUpload(file, null);
    }
    input.value = "";
    if (isMobile) onClose?.();
  }
</script>

<aside
  class="flex flex-col h-full w-60 shrink-0 bg-sidebar text-sidebar-foreground border-r border-sidebar-border select-none"
>
  <div class="flex items-center justify-between px-4 py-3.5 border-b border-sidebar-border">
    <div class="flex items-center gap-2.5">
      <svg
        class="size-5.5 text-primary"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="m21 16-9 5-9-5V8l9-5 9 5v8Z" />
        <path d="m3.27 6.96 8.73 4.86 8.73-4.86" />
        <path d="M12 22.08V12" />
      </svg>
      <span class="font-bold text-sm tracking-tight text-foreground">CriByte</span>
    </div>
    <div class="flex items-center gap-1.5">
      <span class="text-[11px] font-mono text-muted-foreground/70">v0.4.2</span>
      {#if isMobile}
        <button
          type="button"
          onclick={onClose}
          class="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/50 cursor-pointer"
          aria-label="Close sidebar"
        >
          <XIcon class="size-4" />
        </button>
      {/if}
    </div>
  </div>

  <div class="px-3 pt-3 pb-2">
    <button
      type="button"
      onclick={handleUploadClick}
      class="w-full flex items-center justify-between bg-primary hover:bg-primary/90 text-primary-foreground font-medium rounded-md px-3 py-2 shadow-xs transition-colors cursor-pointer"
    >
      <div class="flex items-center gap-1.5 text-xs font-semibold">
        <PlusIcon class="size-4 stroke-[2.5]" />
        <span>Upload</span>
      </div>
      <span class="text-[10px] font-mono tracking-wider uppercase px-1.5 py-0.5 rounded bg-black/15 text-primary-foreground/90 font-medium">drop</span>
    </button>
    <input
      bind:this={fileInputRef}
      type="file"
      multiple
      class="hidden"
      onchange={handleFilesSelected}
    />
  </div>

  <nav class="flex-1 px-2 py-2 space-y-0.5">
    {#each navItems as item (item.href)}
      {@const active = item.match(currentPath)}
      {@const Icon = item.icon}
      <a
        href={item.href}
        onclick={() => isMobile && onClose?.()}
        class="flex items-center gap-2.5 px-3 py-2 rounded-md text-xs font-medium transition-colors {active
          ? 'bg-primary/10 text-primary'
          : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
      >
        <Icon class="size-4 {active ? 'text-primary' : 'text-muted-foreground'}" />
        <span>{item.label}</span>
      </a>
    {/each}
  </nav>

  <div class="mt-auto">
    <div class="px-3 py-3 border-t border-sidebar-border">
      <div class="flex items-center justify-between text-xs text-muted-foreground mb-1.5 font-medium">
        <span>Storage</span>
        <span class="font-mono text-[11px]">42%</span>
      </div>
      <div class="h-1.5 w-full bg-muted rounded-full overflow-hidden mb-2">
        <div class="h-full bg-primary rounded-full transition-all" style="width: 42%"></div>
      </div>
      <div class="flex items-center justify-between text-[11px] font-mono text-muted-foreground">
        <span>4.2 GB used</span>
        <span>10 GB</span>
      </div>
    </div>

    <div class="flex items-center justify-between px-3 py-2.5 text-xs text-muted-foreground border-t border-sidebar-border">
      <a
        href="/settings"
        onclick={() => isMobile && onClose?.()}
        class="flex items-center gap-1.5 hover:text-foreground transition-colors"
      >
        <SettingsIcon class="size-3.5" />
        <span>Settings</span>
      </a>
      <a
        href="https://github.com/Azmi010/cribyte"
        target="_blank"
        rel="noreferrer"
        class="flex items-center gap-1.5 hover:text-foreground transition-colors"
      >
        <HelpCircleIcon class="size-3.5" />
        <span>Docs</span>
      </a>
    </div>
  </div>
</aside>
