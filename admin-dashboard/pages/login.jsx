import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { Server } from 'lucide-react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Login() {
  const router = useRouter();
  const [adminToken, setAdminToken] = useState('');
  const [health, setHealth] = useState(null);

  useEffect(() => {
    fetchHealth();
    // Redirect if already logged in
    if (localStorage.getItem('otaship_admin_token')) {
      router.push('/');
    }
  }, []);

  const fetchHealth = async () => {
    try {
      const res = await fetch(`${API_BASE}/api/health`);
      setHealth(await res.json());
    } catch (e) {
      console.error('Health check failed', e);
    }
  };

  const handleLogin = (e) => {
    e.preventDefault();
    localStorage.setItem('otaship_admin_token', adminToken);

    // Redirect to returnUrl or home
    const returnUrl = router.query.returnUrl || '/';
    router.push(returnUrl);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950 text-white p-4 font-sans selection:bg-blue-500/30">
      <div className="w-full max-w-md bg-gray-900 rounded-2xl shadow-2xl p-8 border border-gray-800 animate-in fade-in zoom-in duration-300">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500 mb-2">
            🚀 OTAShip
          </h1>
          <p className="text-gray-400">Admin Dashboard</p>
        </div>

        {health && (
          <div
            className={`flex items-center justify-center mb-6 px-4 py-2 rounded-full text-xs font-medium border ${
              health.status === 'ok'
                ? 'bg-green-500/10 text-green-400 border-green-500/20'
                : 'bg-red-500/10 text-red-400 border-red-500/20'
            }`}>
            <Server className="w-3 h-3 mr-2" />v{health.version} •{' '}
            {health.status === 'ok' ? 'System Online' : 'System Offline'}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">Admin Token</label>
            <input
              type="password"
              value={adminToken}
              onChange={(e) => setAdminToken(e.target.value)}
              className="w-full px-4 py-3 bg-gray-950 border border-gray-800 rounded-xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 outline-none transition-all text-white placeholder-gray-600"
              placeholder="Enter your secure token"
              required
            />
          </div>
          <button
            type="submit"
            className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 rounded-xl font-semibold text-white shadow-lg shadow-blue-500/20 transform transition-all hover:scale-[1.02] active:scale-[0.98]">
            Access Dashboard
          </button>
        </form>
      </div>
    </div>
  );
}
