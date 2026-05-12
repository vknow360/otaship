import { apiGet } from '$lib/api.js';

export async function load({ cookies }) {
	try {
		const token = cookies.get('otaship_token');
		const updates = await apiGet('api/admin/updates?limit=50', token);
		return { updates, token };
	} catch (error) {
		console.error('Error loading global releases:', error);
		return { updates: { updates: [] }, token: null };
	}
}
