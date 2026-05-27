<script>
	import { apiPut } from '$lib/api.js';

	let { data } = $props();

	let activeProvider = $state('cloudinary');

	// Use $effect to safely sync the prop to local state without warnings
	$effect(() => {
		if (data.settings?.storage_provider) {
			activeProvider = data.settings.storage_provider;
		}
	});

	let isSaving = $state(false);
	let saveMessage = $state('');

	async function saveSettings() {
		isSaving = true;
		saveMessage = '';
		try {
			await apiPut(
				'api/admin/settings',
				{
					key: 'storage_provider',
					value: activeProvider
				},
				data.token
			);
			saveMessage = 'Settings saved successfully!';
			setTimeout(() => (saveMessage = ''), 3000);
		} catch (error) {
			console.error('Failed to save settings:', error);
			saveMessage = 'Error saving settings.';
		} finally {
			isSaving = false;
		}
	}
</script>

<div class="flex flex-col gap-6 px-6 py-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Settings</h1>
		<p class="mt-2 text-sm text-neutral-500">
			Configure your OTAship instance and storage providers.
		</p>
	</div>

	<!-- Storage Provider Selection -->
	<div class="mt-4 rounded-xl border border-neutral-800 bg-neutral-900/50 p-6">
		<h2 class="text-lg font-medium text-white">Active Storage Provider</h2>
		<p class="mb-6 text-sm text-neutral-400">Select where new OTA updates should be uploaded.</p>

		<div class="flex items-end gap-4">
			<div class="max-w-sm flex-1">
				<label for="provider" class="mb-2 block text-sm font-medium text-neutral-300"
					>Provider</label
				>
				<select
					id="provider"
					bind:value={activeProvider}
					class="block w-full rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-white shadow-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none sm:text-sm"
				>
					{#if data.settings?.providers}
						{#each data.settings.providers as provider}
							<option value={provider}
								>{provider.charAt(0).toUpperCase() + provider.slice(1)}</option
							>
						{/each}
					{/if}
				</select>
			</div>
			<button
				onclick={saveSettings}
				disabled={isSaving}
				class="rounded-md bg-white px-4 py-2 text-sm font-semibold text-neutral-900 shadow-sm transition hover:bg-neutral-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white disabled:opacity-50"
			>
				{isSaving ? 'Saving...' : 'Save Changes'}
			</button>
		</div>
		{#if saveMessage}
			<p class="mt-3 text-sm {saveMessage.includes('Error') ? 'text-red-400' : 'text-green-400'}">
				{saveMessage}
			</p>
		{/if}
	</div>

	<!-- Storage Stats Overview -->
	<div>
		<h2 class="mt-6 mb-4 text-xl font-semibold tracking-tight">Storage Usage Overview</h2>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			{#await data.streamed.usage}
				<div class="col-span-full py-10 text-center">
					<div
						class="mx-auto mb-4 h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent"
					></div>
					<p class="text-sm text-neutral-500">Fetching live storage statistics...</p>
				</div>
			{:then usage}
				{#if usage}
					{#each Object.entries(usage) as [provider, stats]}
						<div
							class="flex flex-col gap-3 rounded-xl border border-neutral-800 bg-neutral-900 p-6"
						>
							<div class="flex items-center justify-between border-b border-neutral-800/50 pb-4">
								<h3 class="text-lg font-bold tracking-tight text-white capitalize">{provider}</h3>
								{#if provider === activeProvider}
									<span
										class="inline-flex items-center rounded-full bg-blue-400/10 px-2.5 py-0.5 text-xs font-medium text-blue-400"
										>Active</span
									>
								{/if}
							</div>

							{#if stats.error}
								<p class="mt-2 text-sm text-red-400">Error: {stats.error}</p>
							{:else if stats.message}
								<!-- Generic stats like S3 -->
								<p class="mt-2 text-sm text-neutral-400">{stats.message}</p>
							{:else if stats.storage}
								<!-- Rich stats like Cloudinary -->
								<div class="mt-2 space-y-4">
									<div>
										<div class="mb-1 flex justify-between text-sm">
											<span class="text-neutral-400">Storage</span>
											<span class="text-white">{stats.storage.usage_mb.toFixed(2)} MB</span>
										</div>
									</div>

									<div>
										<div class="mb-1 flex justify-between text-sm">
											<span class="text-neutral-400">Bandwidth</span>
											<span class="text-white">{stats.bandwidth.usage_gb.toFixed(2)} GB</span>
										</div>
										<div class="h-2 w-full overflow-hidden rounded-full bg-neutral-800">
											<div
												class="h-full bg-purple-500"
												style="width: {Math.min(100, (stats.bandwidth.usage_gb / 25) * 100)}%"
											></div>
										</div>
									</div>

									<div
										class="flex justify-between border-t border-neutral-800/50 pt-2 text-xs text-neutral-500"
									>
										<span>Plan: {stats.plan}</span>
										<span>Updated: {stats.last_updated}</span>
									</div>
								</div>
							{/if}
						</div>
					{/each}
				{:else}
					<p class="text-sm text-neutral-500">No usage data available.</p>
				{/if}
			{:catch error}
				<p class="text-sm text-red-400">Failed to load usage data: {error.message}</p>
			{/await}
		</div>
	</div>
</div>
