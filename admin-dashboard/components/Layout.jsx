import React, { useState } from 'react';
import { useRouter } from 'next/router';
import { Menu } from 'lucide-react';
import Sidebar from './Sidebar';
import AuthGuard from './AuthGuard';

const Layout = ({ children }) => {
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const handleLogout = () => {
    localStorage.removeItem('otaship_admin_token');
    router.push('/login');
  };

  return (
    <AuthGuard>
      <div className="flex min-h-screen bg-gray-950 text-white font-sans selection:bg-blue-500/30">
        <Sidebar
          onLogout={handleLogout}
          isOpen={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
        />
        {/* Mobile header with hamburger menu */}
        <div className="md:hidden fixed top-0 left-0 right-0 z-20 bg-gray-900 border-b border-gray-800 px-4 py-3 flex items-center">
          <button
            onClick={() => setSidebarOpen(true)}
            className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors">
            <Menu className="w-6 h-6" />
          </button>
          <h1 className="ml-3 text-lg font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500">
            🚀 OTAShip
          </h1>
        </div>
        <main className="flex-1 md:ml-64 p-4 md:p-8 pt-20 md:pt-8 overflow-x-hidden">
          <div className="max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-500">
            {children}
          </div>
        </main>
      </div>
    </AuthGuard>
  );
};

export default Layout;
