import { apiGet } from '$lib/api.js';

export async function load({ cookies }) {
	const token = cookies.get('otaship_token');

	const [statsRes, updatesRes] = await Promise.allSettled([
		apiGet('api/admin/stats', token),
		apiGet('api/admin/updates?limit=10', token),
	]);

	return {
		stats: statsRes.status === 'fulfilled' ? statsRes.value : null,
		updates: updatesRes.status === 'fulfilled' ? updatesRes.value : null,
	};
}
