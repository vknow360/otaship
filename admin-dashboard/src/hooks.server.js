import { redirect } from '@sveltejs/kit';

export async function handle({ event, resolve }) {
	const start = Date.now();
	const token = event.cookies.get('otaship_token');

	const protectedRoutes = ['/', '/releases', '/projects', '/settings'];
	const isProtected = protectedRoutes.some(
		(path) => event.url.pathname === path || event.url.pathname.startsWith(path + '/')
	);

	let redirectTarget = null;
	if (isProtected && !token) {
		redirectTarget = '/login';
	} else if (event.url.pathname === '/login' && token) {
		redirectTarget = '/';
	}

	if (redirectTarget) {
		const duration = Date.now() - start;
		console.log(
			`[SvelteKit] ${event.request.method} ${event.url.pathname} -> 303 Redirect to ${redirectTarget} (${duration}ms)`
		);
		throw redirect(303, redirectTarget);
	}

	const response = await resolve(event);
	const duration = Date.now() - start;
	console.log(
		`[SvelteKit] ${event.request.method} ${event.url.pathname} -> ${response.status} (${duration}ms)`
	);
	return response;
}

export function handleError({ error, event }) {
	console.error(`[SvelteKit Error] on ${event.url.pathname}:`, error);
	return {
		message: error.message ?? 'Internal Server Error'
	};
}
