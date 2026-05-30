<script>
	import StatsGrid from '$lib/StatsGrid.svelte';
	import RecentUpdates from '$lib/RecentUpdates.svelte';

	let { data } = $props();
</script>

<div class="mx-auto max-w-7xl px-8 py-10">
	<header class="mb-10 flex flex-col justify-between gap-4 md:flex-row md:items-end">
		<div>
			<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">Dashboard Overview</h1>
			<p class="text-sm text-neutral-500">
				Welcome back. Here's what's happening with your deployments.
			</p>
		</div>
		<div
			class="flex items-center gap-2 rounded-full border border-neutral-800 bg-neutral-900/50 px-3 py-1.5 text-xs font-medium text-neutral-600"
		>
			<div class="h-2 w-2 animate-pulse rounded-full bg-emerald-500"></div>
			<span
				>Live Updates: {new Date().toLocaleTimeString([], {
					hour: '2-digit',
					minute: '2-digit'
				})}</span
			>
		</div>
	</header>

	<div class="space-y-10">
		<StatsGrid stats={data.stats} />
		<RecentUpdates updates={data.updates?.updates || []} showControls={false} />
	</div>
</div>
