<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth.svelte";

  let { children } = $props();

  // If already authenticated, redirect to drive
  $effect(() => {
    if (!auth.loading && auth.authenticated) {
      goto("/");
    }
  });
</script>

<div class="min-h-screen bg-background text-foreground">
  {@render children()}
</div>

