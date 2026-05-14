<script>
	import { afterNavigate } from '$app/navigation';

	let { data } = $props();

	let hasPreviousPage = $state(false);
	let nextOffset = $state(0);
	let hasNextPage = $state(false);
	let previousOffset = $state(0);

	$effect(() => {
		if (data.offset == null || data.limit == null) return;
		hasPreviousPage = data.offset > 0;
		nextOffset = data.offset + data.limit;
		hasNextPage = nextOffset < data.total;
		previousOffset = Math.max(0, data.offset - data.limit);
	});

	afterNavigate(() => {
		document.querySelector('main')?.scrollTo({ top: 0, behavior: 'instant' });
	});
</script>

<div class="mx-auto max-w-7xl px-8 py-10">
	<header class="mb-10 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">All Releases</h1>
			<p class="text-sm text-neutral-500">Browse the full release history for this project.</p>
		</div>
		<p class="text-sm text-neutral-500">
			Showing {data.updates.length} of {data.total}
		</p>
	</header>

	<div class="overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900">
		<div class="divide-y divide-neutral-800">
			{#if data.updates.length === 0}
				<div class="p-8 text-sm text-neutral-500">No releases found for this project.</div>
			{:else}
				{#each data.updates as update (update.id)}
					<div class="group p-6 transition-colors hover:bg-white/[0.02]">
						<div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
							<div class="space-y-1">
								<div class="flex flex-wrap items-center gap-3">
									<h2 class="font-semibold text-white transition-colors group-hover:text-blue-400">
										v{update.runtime_version}
									</h2>
									{#if update.is_active}
										<span
											class="rounded border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-emerald-500 uppercase"
										>
											Active
										</span>
									{:else}
										<span
											class="rounded bg-neutral-800 px-2 py-0.5 text-[10px] font-bold tracking-wider text-neutral-400 uppercase"
										>
											Inactive
										</span>
									{/if}
									{#if update.is_rollback}
										<span
											class="rounded border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-amber-500 uppercase"
										>
											Rollback
										</span>
									{/if}
								</div>

								<p class="max-w-xl text-sm text-neutral-500">
									{update.message || 'No release notes provided for this version.'}
								</p>

								<div class="flex items-center gap-1.5 font-mono text-[10px] text-neutral-600">
									<span class="uppercase">ID:</span>
									<span>{update.id}</span>
								</div>
							</div>

							<div class="flex flex-wrap items-center gap-6 text-sm text-neutral-500">
								<div class="flex flex-col items-start gap-1 md:items-end">
									<div class="flex items-center gap-2">
										<span
											class="rounded border border-blue-500/20 bg-blue-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-blue-500 uppercase"
										>
											{update.platform}
										</span>
										<span
											class="rounded bg-neutral-800 px-2 py-0.5 text-[10px] font-bold tracking-wider text-neutral-400 uppercase"
										>
											{update.channel}
										</span>
									</div>
									<span class="text-[10px] tabular-nums">
										{new Date(update.created_at).toLocaleDateString(undefined, {
											month: 'short',
											day: 'numeric',
											year: 'numeric'
										})}
									</span>
								</div>

								<div class="flex flex-col items-center">
									<span class="font-mono font-bold text-white">{update.download_count?.toLocaleString() || '0'}</span>
									<span class="text-[10px] text-neutral-600 uppercase">Downloads</span>
								</div>

								<div class="flex items-center gap-2">
									<a 
										href={`/projects/${data.projectId}/releases/${update.id}`} 
										class="ml-auto rounded p-2 text-neutral-500 hover:bg-neutral-800 hover:text-white transition-colors"
									>
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="9 18 15 12 9 6"></polyline>
										</svg>
									</a>
								</div>
							</div>
						</div>
					</div>
				{/each}
			{/if}
		</div>
	</div>

	{#if data.total > data.limit}
		<div class="mt-6 flex items-center justify-between">
			<a
				href={`/projects/${data.projectId}/releases?limit=${data.limit}&offset=${previousOffset}`}
				aria-disabled={!hasPreviousPage}
				class={`rounded-md border px-4 py-2 text-sm transition ${
					hasPreviousPage
						? 'border-neutral-700 text-white hover:border-neutral-500 hover:bg-neutral-900'
						: 'pointer-events-none border-neutral-800 text-neutral-600'
				}`}
			>
				Previous
			</a>

			<p class="text-sm text-neutral-500">
				Page {Math.floor(data.offset / data.limit) + 1} of {Math.max(
					1,
					Math.ceil(data.total / data.limit)
				)}
			</p>

			<a
				href={`/projects/${data.projectId}/releases?limit=${data.limit}&offset=${nextOffset}`}
				aria-disabled={!hasNextPage}
				class={`rounded-md border px-4 py-2 text-sm transition ${
					hasNextPage
						? 'border-neutral-700 text-white hover:border-neutral-500 hover:bg-neutral-900'
						: 'pointer-events-none border-neutral-800 text-neutral-600'
				}`}
			>
				Next
			</a>
		</div>
	{/if}
</div>
