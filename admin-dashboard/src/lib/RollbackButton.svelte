<script>
	import { apiPost } from '$lib/api';
	import { invalidateAll } from '$app/navigation';

	let { updateId, token } = $props();
	let isRollingBack = $state(false);

	async function handleRollback(e) {
		e.preventDefault();
		e.stopPropagation();

		const confirmed = window.confirm(
			'Are you sure you want to roll back to this version? This will deactivate the current active release for this channel.'
		);
		if (!confirmed) return;

		isRollingBack = true;
		try {
			await apiPost(`api/admin/updates/${updateId}/rollback`, {}, token);
			await invalidateAll();
		} catch (err) {
			alert('Failed to rollback: ' + err.message);
		} finally {
			isRollingBack = false;
		}
	}
</script>

<button
	disabled={isRollingBack}
	onclick={handleRollback}
	class="flex items-center gap-1.5 rounded-lg border border-amber-500/20 bg-amber-500/10 px-3 py-1.5 text-[10px] font-bold tracking-wider text-amber-500 uppercase transition-all hover:bg-amber-500/20 disabled:opacity-50"
	title="Rollback to this version"
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		width="12"
		height="12"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="3"
		stroke-linecap="round"
		stroke-linejoin="round"
		class={isRollingBack ? 'animate-spin' : ''}
	>
		{#if isRollingBack}
			<path d="M21 12a9 9 0 1 1-6.219-8.56" />
		{:else}
			<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
			<path d="M3 3v5h5" />
		{/if}
	</svg>
	{isRollingBack ? 'Rolling back...' : 'Rollback'}
</button>
