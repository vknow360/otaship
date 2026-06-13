<script>
	import { apiPost } from '$lib/api';
	import { invalidateAll } from '$app/navigation';

	let { projectId, token } = $props();

	let open = $state(false);
	let name = $state('');
	let isCreating = $state(false);
	let apiKey = $state('');

	async function handleCreateKey() {
		isCreating = true;
		try {
			const res = await apiPost(`api/admin/projects/${projectId}/keys`, { name }, token);
			if (res && res.api_key) {
				apiKey = res.api_key;
				name = '';
				await invalidateAll();
			} else {
				open = false;
				alert('Failed to create API Key');
			}
		} catch {
			alert('Error creating API key');
			open = false;
		} finally {
			isCreating = false;
		}
	}

	function copyToClipboard() {
		navigator.clipboard.writeText(apiKey);
		alert('Copied to clipboard!');
	}
</script>

<div class="relative">
	{#if open}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
			<div
				class="flex min-w-96 flex-col gap-4 rounded-2xl border border-neutral-800 bg-neutral-900 p-6"
			>
				{#if apiKey}
					<div class="flex h-full flex-col justify-between">
						<div>
							<h2 class="mb-2 text-xl font-bold tracking-tight text-white">
								Key Created Successfully
							</h2>
							<p class="mb-6 text-sm text-neutral-400">
								Copy this key now. You won't be able to see it again.
							</p>

							<div
								class="mb-6 flex items-center justify-between rounded-lg border border-neutral-800 bg-black/50 p-3 font-mono text-sm break-all text-neutral-300"
							>
								<span>{apiKey}</span>
								<button
									onclick={copyToClipboard}
									class="p-2 text-neutral-600 transition-colors hover:text-white"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										width="16"
										height="16"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										stroke-linecap="round"
										stroke-linejoin="round"
										><rect width="14" height="14" x="8" y="8" rx="2" ry="2" /><path
											d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"
										/></svg
									>
								</button>
							</div>
						</div>
						<button
							onclick={() => {
								open = false;
								apiKey = '';
							}}
							class="w-full rounded-lg bg-neutral-200 px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-neutral-400"
						>
							I've saved it
						</button>
					</div>
				{:else}
					<h2 class="text-xl font-bold tracking-tight text-white">Create New API Key</h2>
					<div class="flex flex-col gap-2">
						<label for="name" class="text-sm font-medium text-neutral-400">Name</label>
						<input
							type="text"
							autofocus
							placeholder="e.g. Production CLI"
							bind:value={name}
							class="rounded-lg border border-neutral-800 bg-neutral-900 p-2 text-sm text-white focus:border-neutral-600 focus:outline-none"
						/>
					</div>
					<div class="mt-4 flex flex-row justify-end gap-2">
						<button
							disabled={!name || !projectId || isCreating}
							onclick={handleCreateKey}
							class="rounded-lg bg-neutral-200 px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-neutral-400 disabled:opacity-50"
						>
							{isCreating ? 'Creating...' : 'Create'}
						</button>
						<button
							disabled={isCreating}
							onclick={() => (open = false)}
							class="rounded-lg bg-neutral-800 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-neutral-700"
						>
							Cancel
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
	<button
		onclick={() => (open = true)}
		class="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-neutral-200"
	>
		Create New Key
	</button>
</div>
