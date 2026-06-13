import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export const actions = {
	default: async ({ request, cookies, url }) => {
		const data = await request.formData();
		const token = data.get('token');

		const API_BASE = env.PUBLIC_API_URL || 'http://localhost:8080';

		try {
			const res = await fetch(`${API_BASE}/api/admin/verify`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (!res.ok) {
				return { error: 'Invalid admin token' };
			}

			const secure = url.protocol === 'https:';
			cookies.set('otaship_token', token.toString(), {
				path: '/',
				maxAge: 86400,
				httpOnly: true,
				secure,
				sameSite: 'strict'
			});
		} catch (err) {
			console.error('[Login Error] Failed to connect to backend at', API_BASE, ':', err);
			return { error: 'Failed to connect to server' };
		}

		throw redirect(303, '/');
	}
};
