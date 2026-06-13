<script>
	let { projectId } = $props();
	import { env } from '$env/dynamic/public';

	function copy() {
		navigator.clipboard.writeText(
			`${env.PUBLIC_API_URL || 'http://localhost:8080'}/api/manifest/${projectId}`
		);
	}
</script>

<div class="rounded-xl border border-neutral-800 bg-neutral-900 p-5">
	<h3 class="mb-2 text-sm font-medium">Manifest URL</h3>
	<p class="text-xs text-neutral-500">Use this URL in app.json file</p>
	<div
		class="mt-4 flex items-center justify-between gap-3 rounded-lg border border-neutral-800 bg-black px-3 py-2"
	>
		<code class="text-xs text-neutral-400" id="manifest-url"
			>{env.PUBLIC_API_URL || 'http://localhost:8080'}/api/manifest/{projectId}</code
		>
		<button class="text-neutral-500 hover:text-white" onclick={copy} title="Copy">
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
			>
				<rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
				<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
			</svg>
		</button>
	</div>
</div>

<style>
	button {
		cursor: pointer;
	}
</style>
