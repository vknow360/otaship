<script>
	import Sidebar from '$lib/Sidebar.svelte';
	import { page } from '$app/state';
	import './layout.css';

	let { children } = $props();
</script>

<div class="flex h-screen overflow-hidden bg-gray-950 text-white font-sans antialiased">
	{#if page.url.pathname !== '/login'}
		<Sidebar />
	{/if}
	<main class="flex-1 overflow-y-auto">
		{@render children()}
	</main>
</div>
