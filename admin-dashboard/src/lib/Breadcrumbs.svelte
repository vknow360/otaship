<script>
	import { resolve } from '$app/paths';
	let { items } = $props();
</script>

<nav class="flex text-sm text-neutral-500">
	{#each items as item, i (item.label || i)}
		{#if item.href}
			<a href={resolve(item.href)} class="transition-colors hover:text-white">{item.label}</a>
		{:else}
			<span class={i === items.length - 1 ? 'font-medium text-white' : ''}>{item.label}</span>
		{/if}
		{#if i < items.length - 1}
			<span class="mx-2 text-neutral-700">/</span>
		{/if}
	{/each}
</nav>
