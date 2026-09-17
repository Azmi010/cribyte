<script lang="ts">
  import type { Snippet } from "svelte";
  import TopHeader from "./TopHeader.svelte";

  let {
    children,
    onMenuClick,
    search = $bindable(""),
    onSearchSubmit,
  }: {
    children?: Snippet;
    onMenuClick?: () => void;
    search?: string;
    onSearchSubmit?: (query: string) => void;
  } = $props();
</script>

<div class="flex flex-col flex-1 min-w-0 h-screen overflow-hidden bg-background">
  <TopHeader {onMenuClick} bind:search {onSearchSubmit} />
  <main class="flex-1 flex flex-col min-w-0 overflow-y-auto">
    {@render children?.()}
  </main>
</div>
