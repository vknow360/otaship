<script>
	import RollbackButton from './RollbackButton.svelte';
	import DeleteReleaseButton from './DeleteReleaseButton.svelte';
	let { updates = [], showControls, token, projectId } = $props();
</script>

<div class="mt-10 overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900">
	<div class="flex items-center justify-between border-b border-neutral-800 px-6 py-4">
		<h2 class="text-sm font-semibold tracking-widest text-neutral-400 uppercase">
			Recent Releases
		</h2>
		{#if projectId}
			<a
				href={`/projects/${projectId}/releases`}
				class="text-xs text-neutral-500 transition-colors hover:text-white">View All Activity</a
			>
		{/if}
	</div>
	<div class="divide-y divide-neutral-800">
		{#each updates as update (update.id)}
			<div class="group p-6 transition-colors hover:bg-white/[0.02]">
				<div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
					<div class="space-y-1">
						<div class="flex items-center gap-3">
							<h3 class="font-semibold text-white transition-colors group-hover:text-blue-400">
								v{update.runtime_version}
							</h3>
							{#if update.is_active}
								<span
									class="rounded border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-emerald-500 uppercase"
									>Active</span
								>
							{:else}
								<span
									class="rounded bg-neutral-800 px-2 py-0.5 text-[10px] font-bold tracking-wider text-neutral-400 uppercase"
									>Inactive</span
								>
							{/if}
							{#if showControls && update.is_rollback}
								<span
									class="rounded border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-amber-500 uppercase"
									>Rollback</span
								>
							{/if}
						</div>
						<p class="line-clamp-1 max-w-xl text-sm text-neutral-500">
							{update.message || 'No release notes provided for this version.'}
						</p>
						<div class="flex items-center gap-1.5 font-mono text-[10px] text-neutral-600">
							<span class="uppercase">ID:</span>
							<span>{update.id}</span>
						</div>
					</div>

					<div class="flex items-center gap-6 text-sm text-neutral-500">
						<div class="flex flex-col items-end gap-1">
							<div class="flex items-center gap-2">
								<span
									class="rounded border border-blue-500/20 bg-blue-500/10 px-2 py-0.5 text-[10px] font-bold tracking-wider text-blue-500 uppercase"
									>{update.platform}</span
								>
								<span
									class="rounded bg-neutral-800 px-2 py-0.5 text-[10px] font-bold tracking-wider text-neutral-400 uppercase"
									>{update.channel}</span
								>
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
							<span class="font-mono font-bold text-white">{update.rollout_percentage}%</span>
							<span class="text-[10px] text-neutral-600 uppercase">Rollout</span>
						</div>
						<div class="flex items-center gap-2">
							{#if showControls && !update.is_active}
								<RollbackButton updateId={update.id} {token} />
							{/if}
							{#if showControls}
								<DeleteReleaseButton updateId={update.id} {token} />
							{/if}
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
