import { PUBLIC_API_URL } from '$env/static/public';
const API_BASE = PUBLIC_API_URL || "http://localhost:8080";

export async function apiGet(path, token) {
  const res = await fetch(`${API_BASE}/${path}`, {
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

export async function apiPost(path, data, token) {
  const res = await fetch(`${API_BASE}/${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API error: ${res.status} - ${text}`);
  }
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

export async function apiDelete(path, token) {
  const res = await fetch(`${API_BASE}/${path}`, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  if (res.status === 204) return true;
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

export async function apiPut(path, data, token) {
  const res = await fetch(`${API_BASE}/${path}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}
