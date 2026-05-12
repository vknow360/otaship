import { apiGet } from '$lib/api.js';

export async function load({ cookies }) {
	const token = cookies.get('otaship_token');
	try {
		const settings = await apiGet('api/admin/settings', token);

		const usage = apiGet('api/admin/settings/storage/usage', token);

		return {
			settings,
			streamed: { usage },
			token
		};
	} catch (error) {
		console.error('Error loading settings:', error);
		return { settings: null, usage: null, token };
	}
}
