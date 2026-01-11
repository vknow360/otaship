import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import SettingsView from '../components/SettingsView';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Settings() {
  const [keys, setKeys] = useState([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const headers = { Authorization: `Bearer ${token}` };
      const res = await fetch(`${API_BASE}/api/admin/keys`, { headers });
      if (res.ok) {
        const data = await res.json();
        setKeys(data.keys || []);
      }
    } catch (err) {
      console.error('Data fetch error:', err);
    }
  };

  const handleCreateKey = async (name) => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const res = await fetch(`${API_BASE}/api/admin/keys`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name, scopes: ['admin'] }),
      });
      const data = await res.json();
      if (res.ok) {
        fetchData();
        return data;
      } else {
        alert(data.error || 'Failed to create key');
      }
    } catch (e) {
      alert('Failed to create key');
    }
    return null;
  };

  const handleDeleteKey = async (id) => {
    if (!window.confirm('Revoke this API key?')) return;
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const res = await fetch(`${API_BASE}/api/admin/keys/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) fetchData();
    } catch (e) {
      alert('Failed to delete key');
    }
  };

  return (
    <Layout>
      <SettingsView keys={keys} onCreateKey={handleCreateKey} onDeleteKey={handleDeleteKey} />
    </Layout>
  );
}
