<script>
	import StatsGrid from '$lib/StatsGrid.svelte';
	import RecentUpdates from '$lib/RecentUpdates.svelte';
	import CreateApiModal from '$lib/CreateApiModal.svelte';
	import DeleteApiKeyButton from '$lib/DeleteApiKeyButton.svelte';
	import ManifestUrl from '$lib/ManifestUrl.svelte';

	let { data } = $props();

	function formatApiKey(suffix) {
		return `••••••••••••${suffix}`;
	}
</script>

{#if data.project}
	<div class="mx-auto max-w-7xl px-8 py-10">
		<header class="mb-10">
			<h1 class="mb-2 text-3xl font-bold tracking-tight text-white">{data.project?.name}</h1>
			<p class="text-sm text-neutral-500">{data.project?.description}</p>
		</header>

		<StatsGrid stats={data.stats || {}} />

		<div class="mt-12">
			<ManifestUrl projectId={data.project?.id} />
		</div>

		<div class="mt-12">
			<RecentUpdates
				projectId={data.project?.id}
				updates={data.updates?.updates || []}
				showControls={true}
				token={data.token}
			/>
		</div>

		<div class="mt-16">
			<div class="mb-8 flex items-center justify-between">
				<div>
					<h2 class="text-2xl font-bold tracking-tight text-white">API Keys</h2>
					<p class="mt-1 text-sm text-neutral-500">Manage authentication keys for this project.</p>
				</div>
				<CreateApiModal projectId={data.projectId} token={data.token} />
			</div>

			<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
				{#each data.apiKeys as key (key.id)}
					<div
						class="group rounded-xl border border-neutral-800 bg-neutral-900 p-4 transition-all hover:border-neutral-700"
					>
						<div class="mb-3 flex items-center justify-between">
							<h3 class="truncate pr-2 text-sm font-bold text-white" title={key.name}>
								{key.name}
							</h3>
							<DeleteApiKeyButton projectId={data.projectId} keyId={key.id} token={data.token} />
						</div>

						<div
							class="mb-4 flex items-center justify-between rounded-lg border border-neutral-800 bg-black/50 px-3 py-2 font-mono text-xs text-neutral-300"
						>
							<span>{formatApiKey(key.key_suffix)}</span>
						</div>

						<div
							class="flex items-center justify-between text-[9px] font-bold tracking-tight text-neutral-500 uppercase"
						>
							<span
								>Used: {key.last_used
									? new Date(key.last_used).toLocaleDateString()
									: 'Never'}</span
							>
							<span>Created: {new Date(key.created_at).toLocaleDateString()}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
{:else}
	<div class="flex h-screen items-center justify-center text-white">Project not found.</div>
{/if}
