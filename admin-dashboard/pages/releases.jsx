import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import UpdatesView from '../components/UpdatesView';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Releases() {
  const [updates, setUpdates] = useState([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const headers = { Authorization: `Bearer ${token}` };
      const res = await fetch(`${API_BASE}/api/admin/updates?limit=100`, { headers });
      if (res.ok) {
        const data = await res.json();
        setUpdates(data.updates || []);
      }
    } catch (err) {
      console.error('Data fetch error:', err);
    }
  };

  const handleDeleteUpdate = async (id) => {
    if (!window.confirm('Are you sure you want to delete this update?')) return;
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const res = await fetch(`${API_BASE}/api/admin/updates/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) fetchData();
    } catch (e) {
      alert('Failed to delete update');
    }
  };

  return (
    <Layout>
      <UpdatesView updates={updates} onDelete={handleDeleteUpdate} />
    </Layout>
  );
}
