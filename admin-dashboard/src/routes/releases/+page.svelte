<script>
	import RecentUpdates from '$lib/RecentUpdates.svelte';
	let { data } = $props();
</script>

<div class="mx-auto max-w-7xl px-8 py-10">
	<header class="mb-10">
		<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">Global Releases</h1>
		<p class="text-sm text-neutral-500">A timeline of all software updates across your projects.</p>
	</header>

	<RecentUpdates updates={data.updates?.updates || []} showControls={false} />
</div>
