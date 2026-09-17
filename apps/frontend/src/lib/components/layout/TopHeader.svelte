<script lang="ts">
  import MenuIcon from "@lucide/svelte/icons/menu";
  import SearchIcon from "@lucide/svelte/icons/search";
  import FolderIcon from "@lucide/svelte/icons/folder";
  import SunIcon from "@lucide/svelte/icons/sun";
  import MoonIcon from "@lucide/svelte/icons/moon";
  import LogOutIcon from "@lucide/svelte/icons/log-out";
  import { toggleMode, mode } from "mode-watcher";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
  import { auth } from "$lib/stores/auth.svelte";

  let {
    onMenuClick,
    search = $bindable(""),
    onSearchSubmit,
  }: {
    onMenuClick?: () => void;
    search?: string;
    onSearchSubmit?: (query: string) => void;
  } = $props();

  function handleKeyDown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "k") {
      e.preventDefault();
      const input = document.getElementById("search-files-input");
      input?.focus();
    }
    if (e.key === "Enter") {
      onSearchSubmit?.(search);
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

<header class="h-14 shrink-0 flex items-center justify-between px-4 md:px-6 border-b border-border bg-background">
  <div class="flex items-center gap-3">
    <button
      type="button"
      onclick={onMenuClick}
      class="md:hidden p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-colors cursor-pointer"
      aria-label="Open sidebar"
    >
      <MenuIcon class="size-5" />
    </button>

    <div class="flex items-center gap-1.5 text-xs font-mono text-muted-foreground bg-muted/40 border border-border/60 rounded px-2.5 py-1">
      <FolderIcon class="size-3.5 text-primary" />
      <span>/ root</span>
    </div>
  </div>

  <div class="flex items-center gap-3">
    <div class="relative flex items-center">
      <SearchIcon class="absolute left-2.5 size-3.5 text-muted-foreground pointer-events-none" />
      <input
        id="search-files-input"
        type="text"
        placeholder="Search files..."
        bind:value={search}
        class="h-8 w-36 sm:w-56 md:w-64 pl-8 pr-9 rounded-md bg-muted/40 border border-border/80 text-xs placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary focus:bg-background transition-all"
      />
      <kbd class="absolute right-2 top-1.5 hidden sm:inline-flex items-center text-[10px] font-mono text-muted-foreground border border-border/70 rounded px-1 bg-muted/30 select-none">
        ⌘K
      </kbd>
    </div>

    <div class="hidden sm:flex items-center gap-2 px-2.5 py-1 rounded-md border border-border/60 bg-muted/20 font-mono text-[11px] select-none">
      <span class="size-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
      <div class="flex flex-col text-left leading-none">
        <span class="text-emerald-500 font-medium">connected</span>
        <span class="text-[9px] text-muted-foreground mt-0.5">local docker</span>
      </div>
    </div>

    <DropdownMenu.Root>
      <DropdownMenu.Trigger class="outline-none">
        <div class="size-8 rounded-full bg-muted border border-border/80 flex items-center justify-center text-xs font-semibold overflow-hidden cursor-pointer hover:ring-1 hover:ring-primary transition-all">
          {#if auth.user?.name}
            {auth.user.name.charAt(0).toUpperCase()}
          {:else}
            U
          {/if}
        </div>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end" class="w-48">
        <DropdownMenu.Label class="text-xs font-normal">
          <div class="font-medium text-foreground">{auth.user?.name || "User"}</div>
          <div class="text-[11px] text-muted-foreground truncate">{auth.user?.email || "user@cribyte.local"}</div>
        </DropdownMenu.Label>
        <DropdownMenu.Separator />
        <DropdownMenu.Item onclick={() => toggleMode()} class="cursor-pointer text-xs">
          {#if mode.current === "dark"}
            <SunIcon class="size-3.5 mr-2" />
            <span>Light Mode</span>
          {:else}
            <MoonIcon class="size-3.5 mr-2" />
            <span>Dark Mode</span>
          {/if}
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item onclick={() => auth.logout()} class="cursor-pointer text-xs text-destructive focus:text-destructive">
          <LogOutIcon class="size-3.5 mr-2" />
          <span>Logout</span>
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>
</header>
