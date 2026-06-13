import { apiGet } from '$lib/api.js';

export async function load({ cookies, params }) {
	const token = cookies.get('otaship_token');

	const [projectRes, statsRes, updatesRes, apiKeysRes] = await Promise.allSettled([
		apiGet(`api/admin/projects/${params.id}`, token),
		apiGet(`api/admin/projects/${params.id}/stats`, token),
		apiGet(`api/admin/updates?project_id=${params.id}&limit=5`, token),
		apiGet(`api/admin/projects/${params.id}/keys`, token)
	]);

	if (projectRes.status === 'rejected')
		console.error('[Project Load Error] Project call failed:', projectRes.reason);
	if (statsRes.status === 'rejected')
		console.error('[Project Load Error] Stats call failed:', statsRes.reason);
	if (updatesRes.status === 'rejected')
		console.error('[Project Load Error] Updates call failed:', updatesRes.reason);
	if (apiKeysRes.status === 'rejected')
		console.error('[Project Load Error] API Keys call failed:', apiKeysRes.reason);

	return {
		project: projectRes?.status === 'fulfilled' ? projectRes.value : null,
		stats: statsRes?.status === 'fulfilled' ? statsRes.value : null,
		updates: updatesRes?.status === 'fulfilled' ? updatesRes.value : null,
		apiKeys: apiKeysRes?.status === 'fulfilled' ? apiKeysRes.value : [],
		token,
		projectId: params.id
	};
}
