<script lang="ts">
  import HomeIcon from "@lucide/svelte/icons/home";
  import FolderIcon from "@lucide/svelte/icons/folder";
  import FolderPlusIcon from "@lucide/svelte/icons/folder-plus";
  import LayoutGridIcon from "@lucide/svelte/icons/layout-grid";
  import ListIcon from "@lucide/svelte/icons/list";
  import ArrowDownIcon from "@lucide/svelte/icons/arrow-down";
  import ArrowUpIcon from "@lucide/svelte/icons/arrow-up";
  import ChevronRightIcon from "@lucide/svelte/icons/chevron-right";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
  import { view } from "$lib/stores/view.svelte";

  interface BreadcrumbSegment {
    id: string | null;
    name: string;
  }

  let {
    breadcrumbs = [{ id: null, name: "Home" }],
    onNavigate,
    onNewFolder,
  }: {
    breadcrumbs?: BreadcrumbSegment[];
    onNavigate?: (id: string | null) => void;
    onNewFolder?: () => void;
  } = $props();

  const sortOptions = [
    { value: "updated_at", label: "Modified" },
    { value: "name", label: "Name" },
    { value: "size", label: "Size" },
    { value: "created_at", label: "Created" },
  ] as const;

  const currentSortLabel = $derived(
    sortOptions.find((opt) => opt.value === view.sortBy)?.label ?? "Modified"
  );
</script>

<div class="flex flex-wrap items-center justify-between gap-3 px-4 md:px-6 py-3 border-b border-border/60 bg-background select-none">
  <nav class="flex items-center gap-1.5 text-xs text-muted-foreground overflow-x-auto py-0.5">
    {#each breadcrumbs as crumb, index (crumb.id ?? "root")}
      {@const isLast = index === breadcrumbs.length - 1}
      {#if index > 0}
        <ChevronRightIcon class="size-3.5 text-muted-foreground/50 shrink-0" />
      {/if}
      <button
        type="button"
        onclick={() => onNavigate?.(crumb.id)}
        class="flex items-center gap-1.5 px-1.5 py-1 rounded hover:bg-muted/60 transition-colors shrink-0 {isLast
          ? 'font-semibold text-foreground'
          : 'text-muted-foreground hover:text-foreground'}"
      >
        {#if index === 0}
          <HomeIcon class="size-3.5" />
          <span>{crumb.name}</span>
        {:else if isLast}
          <FolderIcon class="size-3.5 text-primary" />
          <span>{crumb.name}</span>
        {:else}
          <span>{crumb.name}</span>
        {/if}
      </button>
    {/each}
  </nav>

  <div class="flex items-center gap-2.5 shrink-0">
    <DropdownMenu.Root>
      <DropdownMenu.Trigger class="outline-none">
        <button
          type="button"
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md border border-border/70 text-xs text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors font-mono cursor-pointer"
        >
          <span>Sort: {currentSortLabel}</span>
          {#if view.sortOrder === "asc"}
            <ArrowUpIcon class="size-3 text-primary" />
          {:else}
            <ArrowDownIcon class="size-3 text-primary" />
          {/if}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end" class="w-40">
        <DropdownMenu.Label class="text-xs text-muted-foreground font-mono">Sort by</DropdownMenu.Label>
        <DropdownMenu.Separator />
        {#each sortOptions as opt (opt.value)}
          <DropdownMenu.Item
            onclick={() => {
              if (view.sortBy === opt.value) {
                view.toggleSortOrder();
              } else {
                view.setSortBy(opt.value);
              }
            }}
            class="flex items-center justify-between text-xs cursor-pointer {view.sortBy === opt.value ? 'text-primary font-medium' : ''}"
          >
            <span>{opt.label}</span>
            {#if view.sortBy === opt.value}
              <span class="text-[10px] font-mono uppercase">{view.sortOrder}</span>
            {/if}
          </DropdownMenu.Item>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <div class="flex items-center p-0.5 rounded-md border border-border/70 bg-muted/30">
      <button
        type="button"
        onclick={() => view.setViewMode("grid")}
        class="p-1 rounded transition-colors cursor-pointer {view.viewMode === 'grid'
          ? 'bg-muted text-foreground shadow-2xs'
          : 'text-muted-foreground hover:text-foreground'}"
        title="Grid view"
        aria-label="Grid view"
      >
        <LayoutGridIcon class="size-3.5" />
      </button>
      <button
        type="button"
        onclick={() => view.setViewMode("list")}
        class="p-1 rounded transition-colors cursor-pointer {view.viewMode === 'list'
          ? 'bg-muted text-foreground shadow-2xs'
          : 'text-muted-foreground hover:text-foreground'}"
        title="List view"
        aria-label="List view"
      >
        <ListIcon class="size-3.5" />
      </button>
    </div>

    <button
      type="button"
      onclick={onNewFolder}
      class="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-border/80 hover:bg-muted/60 text-xs font-medium text-foreground transition-colors cursor-pointer"
    >
      <FolderPlusIcon class="size-3.5 text-muted-foreground" />
      <span>New Folder</span>
    </button>
  </div>
</div>
