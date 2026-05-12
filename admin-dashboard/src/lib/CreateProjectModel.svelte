<script>
	import { apiPost } from '$lib/api';
	import { invalidateAll } from '$app/navigation';

	let { token } = $props();

	let open = $state(false);
	let name = $state('');
	let description = $state('');
	let slug = $state('');
	let isCreating = $state(false);

	async function handleCreateProject() {
		isCreating = true;
		try {
			const res = await apiPost(
				'api/admin/projects',
				{
					name,
					description,
					slug
				},
				token
			);

			if (res && res.id) {
				await invalidateAll();
				name = '';
				description = '';
				slug = '';
				open = false;
			} else {
				alert('Failed to create project');
			}
		} catch (error) {
			alert('Error creating project: ' + JSON.stringify(error));
		} finally {
			isCreating = false;
		}
	}
</script>

<div class="relative">
	{#if open}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
			<div
				class="flex min-w-96 flex-col gap-4 rounded-2xl border border-neutral-800 bg-neutral-900 p-6"
			>
				<h2 class="text-xl font-bold tracking-tight text-white">Create New Project</h2>
				<div class="flex flex-col gap-2">
					<label for="name" class="text-sm font-medium text-neutral-400">Name</label>
					<input
						type="text"
						autofocus
						placeholder="e.g. My First Project"
						bind:value={name}
						class="rounded-lg border border-neutral-800 bg-neutral-900 p-2 text-sm text-white focus:border-neutral-600 focus:outline-none"
					/>
				</div>
				<div class="flex flex-col gap-2">
					<label for="description" class="text-sm font-medium text-neutral-400">Description</label>
					<textarea
						rows="3"
						placeholder="Describe what this project is for"
						bind:value={description}
						class="resize-none rounded-lg border border-neutral-800 bg-neutral-900 p-2 text-sm text-white focus:border-neutral-600 focus:outline-none"
					></textarea>
				</div>
				<div class="flex flex-col gap-2">
					<label for="slug" class="text-sm font-medium text-neutral-400">Project Slug</label>
					<input
						type="text"
						placeholder="e.g. my-first-project"
						bind:value={slug}
						class="rounded-lg border border-neutral-800 bg-neutral-900 p-2 text-sm text-white focus:border-neutral-600 focus:outline-none"
					/>
				</div>
				<div class="mt-4 flex flex-row justify-end gap-2">
					<button
						onclick={handleCreateProject}
						disabled={!name || !slug || isCreating}
						class="rounded-lg bg-neutral-200 px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-neutral-400 disabled:opacity-50"
					>
						{isCreating ? 'Creating...' : 'Create'}
					</button>
					<button
						onclick={() => (open = false)}
						disabled={isCreating}
						class="rounded-lg bg-neutral-800 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-neutral-700"
					>
						Cancel
					</button>
				</div>
			</div>
		</div>
	{/if}
	<button
		onclick={() => (open = true)}
		class="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-neutral-200"
	>
		Create New Project
	</button>
</div>
