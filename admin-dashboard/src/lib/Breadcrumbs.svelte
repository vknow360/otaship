<script>
    let { items } = $props(); 
</script>

<nav class="flex text-sm text-neutral-500">
    {#each items as item, i}
        {#if item.href}
            <a href={item.href} class="hover:text-white transition-colors">{item.label}</a>
        {:else}
            <span class={i === items.length - 1 ? 'text-white font-medium' : ''}>{item.label}</span>
        {/if}
        {#if i < items.length - 1}
            <span class="mx-2 text-neutral-700">/</span>
        {/if}
    {/each}
</nav>
