import React from 'react';
import { useRouter } from 'next/router';
import Sidebar from './Sidebar';
import AuthGuard from './AuthGuard';

const Layout = ({ children }) => {
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem('otaship_admin_token');
    router.push('/login');
  };

  return (
    <AuthGuard>
      <div className="flex min-h-screen bg-gray-950 text-white font-sans selection:bg-blue-500/30">
        <Sidebar onLogout={handleLogout} />
        <main className="flex-1 ml-64 p-8 overflow-x-hidden">
          <div className="max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-500">
            {children}
          </div>
        </main>
      </div>
    </AuthGuard>
  );
};

export default Layout;
