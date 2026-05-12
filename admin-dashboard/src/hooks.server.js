import { redirect } from '@sveltejs/kit';

export async function handle({ event, resolve }) {
	const token = event.cookies.get('otaship_token');

	const protectedRoutes = ['/', '/releases', '/projects', '/settings'];
	const isProtected = protectedRoutes.some(
		path => event.url.pathname === path || event.url.pathname.startsWith(path + '/')
	);

	if (isProtected && !token) {
		throw redirect(303, '/login');
	}

	if (event.url.pathname === '/login' && token) {
		throw redirect(303, '/');
	}

	return resolve(event);
}
