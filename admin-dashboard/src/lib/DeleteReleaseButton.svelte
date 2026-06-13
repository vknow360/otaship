<script>
	import { apiDelete } from '$lib/api';
	import { invalidateAll, goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { updateId, token, redirect } = $props();
	let isDeleting = $state(false);

	async function handleDelete(e) {
		e.preventDefault();
		e.stopPropagation();

		const confirmed = window.confirm(
			'Are you sure you want to delete this release? This action cannot be undone and will remove all associated asset records.'
		);
		if (!confirmed) return;

		isDeleting = true;
		try {
			await apiDelete(`api/admin/updates/${updateId}`, token);
			if (redirect) {
				await goto(resolve(redirect));
			} else {
				await invalidateAll();
			}
		} catch (err) {
			alert('Failed to delete release: ' + err.message);
		} finally {
			isDeleting = false;
		}
	}
</script>

<button
	disabled={isDeleting}
	onclick={handleDelete}
	class="p-1.5 text-neutral-600 transition-colors hover:text-red-500 disabled:opacity-50"
	title="Delete Release"
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		width="14"
		height="14"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
		class={isDeleting ? 'animate-spin' : ''}
	>
		{#if isDeleting}
			<path d="M21 12a9 9 0 1 1-6.219-8.56" />
		{:else}
			<path d="M3 6h18" />
			<path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
			<path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
		{/if}
	</svg>
</button>
