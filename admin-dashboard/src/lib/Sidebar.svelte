<script>
	import { page } from '$app/state';
	import { enhance } from '$app/forms';
	import icon from '$lib/assets/icon-512.png';
</script>

<div class="flex min-h-screen w-52 flex-col border-r border-gray-800">
	<div>
		<div class="flex items-center justify-center gap-2 p-8 text-2xl font-bold">
			<img src={icon} alt="OTAShip" width="32" height="32" class="rounded-lg" /> OTAShip
		</div>
		<hr class="mb-10 h-px border-0 border-gray-800 bg-gray-800" />
		<div class="px-4">
			<nav>
				<ul>
					{#each [{ href: '/', label: 'Overview' }, { href: '/projects', label: 'Projects' }, { href: '/settings', label: 'Settings' }] as link}
						<li class="mb-2">
							<a
								href={link.href}
								class="block rounded-lg px-4 py-2 transition-colors {page.url.pathname === link.href
									? 'bg-white font-medium text-black'
									: 'text-gray-400 hover:bg-gray-800 hover:text-white'}"
							>
								{link.label}
							</a>
						</li>
					{/each}
				</ul>
			</nav>
		</div>
	</div>

	<div class="mt-auto w-full px-4 pb-8">
		<hr class="mb-5 h-px border-0 border-gray-800 bg-gray-800" />
		<form action="/logout" method="POST" use:enhance class="w-full">
			<button
				type="submit"
				class="w-full cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-white transition-colors hover:bg-gray-700"
				>Logout</button
			>
		</form>
	</div>
</div>
