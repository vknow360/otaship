import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ cookies }) => {
		cookies.delete('otaship_token', { path: '/' });
		throw redirect(303, '/login');
	}
};

export async function load() {
	throw redirect(303, '/login');
}
