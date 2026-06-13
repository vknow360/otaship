<script>
	import Breadcrumbs from '$lib/Breadcrumbs.svelte';
	import RollbackButton from '$lib/RollbackButton.svelte';
	import DeleteReleaseButton from '$lib/DeleteReleaseButton.svelte';

	let { data } = $props();
	let update = $derived(data.update);
	let project = $derived(data.project);

	function formatBytes(bytes, decimals = 2) {
		if (!+bytes) return '0 Bytes';
		const k = 1024;
		const dm = decimals < 0 ? 0 : decimals;
		const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
	}
</script>

<svelte:head>
	<title>{project.name} - Release {update.id.split('-')[0]} - OTAShip</title>
</svelte:head>

<div class="mx-auto max-w-7xl px-8 py-10">
	<div class="mb-10">
		<Breadcrumbs
			items={[
				{ label: 'Projects', href: '/projects' },
				{ label: project.name, href: `/projects/${project.id}` },
				{ label: 'Releases' },
				{ label: update.id.split('-')[0] }
			]}
		/>
		<header class="mt-4 flex items-start justify-between">
			<div>
				<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">
					{update.message || 'No release notes'}
				</h1>
				<p class="font-mono text-sm text-neutral-500">{update.id}</p>
			</div>
			<div class="flex gap-2">
				{#if !update.is_active}
					<RollbackButton updateId={update.id} token={data.token} />
				{/if}
				<DeleteReleaseButton
					updateId={update.id}
					token={data.token}
					redirect={`/projects/${project.id}`}
				/>
			</div>
		</header>
	</div>

	<div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-4">
			<h3 class="text-xs font-semibold tracking-widest text-neutral-500 uppercase">Platform</h3>
			<p class="mt-2 text-lg font-semibold text-white capitalize">{update.platform}</p>
		</div>
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-4">
			<h3 class="text-xs font-semibold tracking-widest text-neutral-500 uppercase">Channel</h3>
			<p class="mt-2 text-lg font-semibold text-white capitalize">{update.channel}</p>
		</div>
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-4">
			<h3 class="text-xs font-semibold tracking-widest text-neutral-500 uppercase">Rollout</h3>
			<p class="mt-2 text-lg font-semibold text-white">{update.rollout_percentage}%</p>
		</div>
		<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-4">
			<h3 class="text-xs font-semibold tracking-widest text-neutral-500 uppercase">Downloads</h3>
			<p class="mt-2 text-lg font-semibold text-white">
				{update.download_count?.toLocaleString() || '0'}
			</p>
		</div>
	</div>

	<section class="mt-12">
		<h2 class="mb-4 text-lg font-semibold text-white">Assets Payload</h2>
		{#await data.streamed.assets}
			<div
				class="flex items-center justify-center rounded-xl border border-neutral-800 bg-neutral-900/50 p-8"
			>
				<span class="animate-pulse text-sm text-neutral-500">Loading assets...</span>
			</div>
		{:then assets}
			<div class="overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900">
				<table class="w-full text-left text-sm text-neutral-400">
					<thead
						class="bg-neutral-800/50 text-xs font-semibold tracking-widest text-neutral-500 uppercase"
					>
						<tr>
							<th class="px-6 py-4">Filename</th>
							<th class="px-6 py-4 text-right">Size</th>
							<th class="px-6 py-4 text-right">Storage Provider</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-neutral-800">
						{#each assets as asset (asset.file_name)}
							<tr class="transition-colors hover:bg-white/[0.02]">
								<td class="px-6 py-4 font-mono text-neutral-300">{asset.file_name}</td>
								<td class="px-6 py-4 text-right font-mono">{formatBytes(asset.size)}</td>
								<td class="px-6 py-4 text-right font-mono text-neutral-300"
									>{asset.storage_provider}</td
								>
							</tr>
						{:else}
							<tr>
								<td colspan="3" class="px-6 py-8 text-center text-neutral-500"
									>No assets found for this release.</td
								>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:catch error}
			<div class="rounded-xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-400">
				Failed to load assets: {error.message}
			</div>
		{/await}
	</section>
</div>
