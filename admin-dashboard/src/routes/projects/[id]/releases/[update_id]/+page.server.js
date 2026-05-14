import { apiGet } from '$lib/api';

export async function load({ params, cookies }) {
    const token = cookies.get('otaship_token');

    const [projectRes, updateRes] = await Promise.all([
        apiGet(`api/admin/projects/${params.id}`, token),
        apiGet(`api/admin/updates/${params.update_id}`, token)
    ]);

    const assetsPromise = apiGet(`api/admin/updates/${params.update_id}/assets`, token);

    return {
        project: projectRes,
        update: updateRes,
        streamed: {
            assets: assetsPromise
        },
        token
    };
}
