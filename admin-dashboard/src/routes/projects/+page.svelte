<script>
	import CreateProjectModel from '$lib/CreateProjectModel.svelte';
	import DeleteProjectModel from '$lib/DeleteProjectModel.svelte';
	import { resolve } from '$app/paths';

	let { data } = $props();
</script>

<div class="flex flex-col gap-6 px-6 py-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Projects</h1>
			<p class="mt-2 text-sm text-neutral-500">Manage your projects and applications.</p>
		</div>

		<CreateProjectModel token={data.token} />
	</div>

	<div class="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
		{#each data.projects as project (project.id)}
			<a
				href={resolve(`/projects/${project.id}`)}
				class="group flex flex-col gap-3 rounded-xl border border-neutral-800 bg-neutral-900 p-6 transition-all hover:border-neutral-700 hover:bg-neutral-800/50"
			>
				<div class="flex items-center justify-between">
					<h2 class="text-xl font-bold tracking-tight text-white">{project.name}</h2>
					<DeleteProjectModel projectId={project.id} token={data.token} />
				</div>
				<p class="line-clamp-2 text-sm text-neutral-500">{project.description}</p>
				<div class="mt-auto flex items-center justify-between border-t border-neutral-800/50 pt-4">
					<span class="text-[10px] font-bold tracking-wider text-neutral-600 uppercase"
						>Created</span
					>

					<span class="text-xs text-neutral-500"
						>{new Date(project.created_at).toLocaleDateString()}</span
					>
				</div>
			</a>
		{/each}
	</div>
</div>
