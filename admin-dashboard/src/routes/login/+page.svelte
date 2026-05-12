<script>
	import { goto } from '$app/navigation';
	import icon from '$lib/assets/icon-512.png';

	let token = $state('');
	let error = $state(null);
	let loading = $state(false);

	async function handleLogin(e) {
		e.preventDefault();
		error = null;
		loading = true;

		try {
			const API_BASE = import.meta.env.PUBLIC_API_BASE || 'http://localhost:8080';
			const res = await fetch(`${API_BASE}/api/admin/verify`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) {
				error = 'Invalid token';
			} else {
				document.cookie = `otaship_token=${token}; path=/; max-age=86400; SameSite=Strict; Secure;`;
				goto('/');
			}
		} catch (err) {
			error = 'Invalid token provided. Please try again.';
		} finally {
			loading = false;
		}
	}
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

		<form onsubmit={handleLogin} class="space-y-4">
			<div>
				<label for="adminToken" class="mb-4 text-neutral-500">Enter admin token</label>
				<input
					id="adminToken"
					type="password"
					bind:value={token}
					placeholder="Admin token..."
					class="mt-2 w-full rounded-lg border border-neutral-800 bg-neutral-900 px-4 py-2.5 text-sm text-neutral-200 placeholder-neutral-600 transition-colors focus:border-neutral-600 focus:outline-none"
					required
				/>
			</div>

			{#if error}
				<div class="px-1 text-sm text-red-500">{error}</div>
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
