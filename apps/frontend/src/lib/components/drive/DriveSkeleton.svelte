<script lang="ts">
  import { view } from "$lib/stores/view.svelte";

  let { mode }: { mode?: "grid" | "list" } = $props();

  const activeMode = $derived(mode ?? view.viewMode);
</script>

{#if activeMode === "grid"}
  <div class="space-y-6" aria-hidden="true">
    <section>
      <div class="mb-3 h-3 w-24 rounded bg-muted animate-pulse"></div>
      <div class="grid grid-cols-1 min-[480px]:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
        {#each Array(4) as _, i (i)}
          <div class="border border-border/70 bg-card rounded-lg p-3.5 flex flex-col gap-3">
            <div class="size-6 rounded bg-muted animate-pulse"></div>
            <div class="h-3 w-3/4 rounded bg-muted animate-pulse"></div>
            <div class="h-2.5 w-1/3 rounded bg-muted/70 animate-pulse"></div>
          </div>
        {/each}
      </div>
    </section>
    <section>
      <div class="mb-3 h-3 w-20 rounded bg-muted animate-pulse"></div>
      <div class="grid grid-cols-1 min-[480px]:grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
        {#each Array(6) as _, i (i)}
          <div class="border border-border/70 bg-card rounded-lg overflow-hidden">
            <div class="h-28 bg-muted/50 animate-pulse border-b"></div>
            <div class="p-2.5 flex flex-col gap-2">
              <div class="h-3 w-4/5 rounded bg-muted animate-pulse"></div>
              <div class="h-2.5 w-2/3 rounded bg-muted/70 animate-pulse"></div>
            </div>
          </div>
        {/each}
      </div>
    </section>
  </div>
{:else}
  <div class="border rounded-md bg-card overflow-hidden" aria-hidden="true">
    <div class="flex items-center px-4 py-2 border-b bg-muted/30 text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
      <div class="flex-1">Name</div>
      <div class="w-24 text-right">Size</div>
      <div class="w-32 text-right">Modified</div>
      <div class="w-8"></div>
    </div>
    <div class="flex flex-col">
      {#each Array(6) as _, i (i)}
        <div class="flex items-center py-2.5 px-3 border-b border-border/40 gap-3">
          <div class="size-5 rounded bg-muted animate-pulse shrink-0"></div>
          <div class="h-3 flex-1 max-w-[40%] rounded bg-muted animate-pulse"></div>
          <div class="h-2.5 w-16 rounded bg-muted/70 animate-pulse shrink-0"></div>
          <div class="h-2.5 w-20 rounded bg-muted/70 animate-pulse shrink-0"></div>
          <div class="w-8"></div>
        </div>
      {/each}
    </div>
  </div>
{/if}
