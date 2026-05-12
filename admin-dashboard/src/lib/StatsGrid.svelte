<script>
	let { stats } = $props();
</script>

<div class="space-y-6">
	<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
		<!-- Card 1 -->
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
			<div class="mb-4 flex items-center justify-between">
				<span class="text-sm font-medium text-neutral-500">Total Downloads</span>
				<div class="text-neutral-400">
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="7 10 12 15 17 10" /><line x1="12" y1="15" x2="12" y2="3" /></svg>
				</div>
			</div>
			<div class="flex flex-col gap-1">
				<span class="text-2xl font-bold tracking-tight">{stats?.total_downloads?.toLocaleString()}</span>
				<span class="text-xs text-neutral-500">Across all platforms</span>
			</div>
		</div>

		<!-- Card 2: Recent -->
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
			<div class="mb-4 flex items-center justify-between">
				<span class="text-sm font-medium text-neutral-500">Recent (24h)</span>
				<div class="text-neutral-400">
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12" /></svg>
				</div>
			</div>
			<div class="flex flex-col gap-1">
				<span class="text-2xl font-bold tracking-tight">{stats?.recent_downloads?.toLocaleString()}</span>
				<span class="text-xs text-neutral-500">Downloads in last 24 hours</span>
			</div>
		</div>

		<!-- Card 3: Channels -->
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
			<div class="mb-4 flex items-center justify-between">
				<span class="text-sm font-medium text-neutral-500">Channels</span>
				<div class="text-neutral-400">
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2" /></svg>
				</div>
			</div>
			<div class="flex flex-col gap-1">
				<span class="text-2xl font-bold tracking-tight">{stats?.by_channel?.length || 0}</span>
				<span class="text-xs text-neutral-500">{stats?.by_channel?.map(c => c.channel).join(', ') || 'None'}</span>
			</div>
		</div>

		<!-- Card 4: Platforms -->
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
			<div class="mb-4 flex items-center justify-between">
				<span class="text-sm font-medium text-neutral-500">Platforms</span>
				<div class="text-neutral-400">
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2" /><line x1="12" y1="18" x2="12.01" y2="18" /></svg>
				</div>
			</div>
			<div class="flex flex-col gap-1">
				<span class="text-2xl font-bold tracking-tight">{stats?.by_platform?.length || 0}</span>
				<span class="text-xs text-neutral-500">Active device targets</span>
			</div>
		</div>
	</div>

	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-6">
			<h3 class="mb-6 text-sm font-medium tracking-wider text-neutral-400 uppercase">Distribution by Channel</h3>
			<div class="space-y-4">
				{#each stats?.by_channel || [] as channel}
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">{channel.channel}</span>
						<div class="flex items-center gap-3">
							<div class="h-1.5 w-32 overflow-hidden rounded-full bg-neutral-800">
								<div
									class="h-full bg-white"
									style="width: {(channel.count / stats.total_downloads) * 100}%"
								></div>
							</div>
							<span class="w-10 text-right text-sm text-neutral-400">{channel.count}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-6">
			<h3 class="mb-6 text-sm font-medium tracking-wider text-neutral-400 uppercase">Distribution by Platform</h3>
			<div class="space-y-4">
				{#each stats?.by_platform || [] as platform}
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">{platform.platform}</span>
						<div class="flex items-center gap-3">
							<div class="h-1.5 w-32 overflow-hidden rounded-full bg-neutral-800">
								<div
									class="h-full bg-neutral-400"
									style="width: {(platform.count / stats.total_downloads) * 100}%"
								></div>
							</div>
							<span class="w-10 text-right text-sm text-neutral-400">{platform.count}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
