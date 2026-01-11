import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import Overview from '../components/Overview';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [updates, setUpdates] = useState([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const headers = { Authorization: `Bearer ${token}` };

      const [statsRes, updatesRes] = await Promise.all([
        fetch(`${API_BASE}/api/admin/stats`, { headers }),
        fetch(`${API_BASE}/api/admin/updates?limit=5`, { headers }),
      ]);

      if (statsRes.ok) setStats(await statsRes.json());
      if (updatesRes.ok) setUpdates((await updatesRes.json()).updates || []);
    } catch (err) {
      console.error('Data fetch error:', err);
    }
  };

  return (
    <Layout>
      <Overview stats={stats} updates={updates} />
    </Layout>
  );
}
