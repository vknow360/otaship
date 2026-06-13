import { apiGet } from '$lib/api.js';

export async function load({ cookies }) {
	const token = cookies.get('otaship_token');

	const [statsRes, updatesRes] = await Promise.allSettled([
		apiGet('api/admin/stats', token),
		apiGet('api/admin/updates?limit=10', token)
	]);

	if (statsRes.status === 'rejected') {
		console.error('[Dashboard Load Error] Stats call failed:', statsRes.reason);
	}
	if (updatesRes.status === 'rejected') {
		console.error('[Dashboard Load Error] Updates call failed:', updatesRes.reason);
	}

	return {
		stats: statsRes.status === 'fulfilled' ? statsRes.value : null,
		updates: updatesRes.status === 'fulfilled' ? updatesRes.value : null
	};
}
