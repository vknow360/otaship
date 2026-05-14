<script>
	import { enhance } from '$app/forms';
	import icon from '$lib/assets/icon-512.png';

	let { form } = $props();
	let token = $state('');
	let loading = $state(false);
</script>

<div
	class="flex min-h-screen items-center justify-center bg-black font-sans text-neutral-100 antialiased"
>
	<div class="w-full max-w-sm rounded-lg border border-neutral-800 px-6 py-8">
		<div class="mb-10 text-center">
			<img src={icon} alt="OTAShip" width="64" height="64" class="mx-auto mb-2 rounded-lg" />
			<h1 class="text-xl font-medium tracking-tight">OTAShip</h1>
			<p class="mt-2 text-sm text-neutral-500">Sign in to dashboard</p>
		</div>

		<form method="POST" use:enhance={() => {
			loading = true;
			return async ({ update }) => {
				await update();
				loading = false;
			};
		}} class="space-y-4">
			<div>
				<label for="adminToken" class="mb-4 text-neutral-500">Enter admin token</label>
				<input
					id="adminToken"
					name="token"
					type="password"
					bind:value={token}
					placeholder="Admin token..."
					class="mt-2 w-full rounded-lg border border-neutral-800 bg-neutral-900 px-4 py-2.5 text-sm text-neutral-200 placeholder-neutral-600 transition-colors focus:border-neutral-600 focus:outline-none"
					required
				/>
			</div>

			{#if form?.error}
				<div class="px-1 text-sm text-red-500">{form?.error}</div>
			{/if}

			<button
				type="submit"
				disabled={loading || !token.trim()}
				class="w-full rounded-lg bg-white px-4 py-2.5 text-sm font-medium text-black transition-colors hover:bg-neutral-200 disabled:opacity-50"
			>
				{loading ? 'Verifying...' : 'Continue'}
			</button>
		</form>
	</div>
</div>
