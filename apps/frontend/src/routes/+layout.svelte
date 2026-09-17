<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import { onMount } from "svelte";
  import { ModeWatcher } from "mode-watcher";
  import { Toaster } from "$lib/components/ui/sonner";
  import { auth } from "$lib/stores/auth.svelte";

  let { children } = $props();

  onMount(() => {
    auth.fetchUser();
  });
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>CriByte - Self-Hosted Cloud Storage</title>
</svelte:head>

<ModeWatcher defaultMode="dark" />
<Toaster position="bottom-right" />

{@render children()}
