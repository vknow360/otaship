import { apiGet } from "$lib/api";

export async function load({ url, params, cookies }) {
    let projectId = params.id;
    let limit = Number(url.searchParams.get('limit') || 10);
    let offset = Number(url.searchParams.get('offset') || 0);
    let token = cookies.get('otaship_token');

    const [updatesRes] = await Promise.allSettled([
        apiGet(`api/admin/updates?project_id=${projectId}&limit=${limit}&offset=${offset}`, token),
    ]);

    const updatesData = updatesRes.status === 'fulfilled'
        ? updatesRes.value
        : { updates: [], total: 0, limit, offset };

    return {
        updates: updatesData.updates || [],
        total: updatesData.total || 0,
        limit: updatesData.limit || limit,
        offset: updatesData.offset || offset,
        projectId,
        token,
    };
}
