import { apiGet } from '$lib/api.js';

export async function load({ cookies }) {
	try {
		const token = cookies.get('otaship_token');
		const projects = await apiGet('api/admin/projects', token);
		return { projects, token };
	} catch (error) {
		console.error('Error loading projects:', error);
		return { projects: [], token: null };
	}
}
