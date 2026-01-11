import React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { LayoutDashboard, List, Box, Settings, LogOut, X } from 'lucide-react';

const Sidebar = ({ onLogout, isOpen, onClose }) => {
  const router = useRouter();
  const currentPath = router.pathname;

  const items = [
    { id: 'overview', label: 'Overview', path: '/', icon: LayoutDashboard },
    { id: 'releases', label: 'Releases', path: '/releases', icon: List },
    { id: 'projects', label: 'Projects', path: '/projects', icon: Box },
    { id: 'settings', label: 'Settings', path: '/settings', icon: Settings },
  ];

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && <div className="fixed inset-0 bg-black/60 z-30 md:hidden" onClick={onClose} />}
      <aside
        className={`w-64 bg-gray-900 border-r border-gray-800 flex flex-col h-screen fixed left-0 top-0 z-40 transition-transform duration-300 ${isOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0`}>
        <div className="p-6 border-b border-gray-800 flex items-center justify-between">
          <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500">
            🚀 OTAShip
          </h1>
          {/* Mobile close button */}
          <button
            onClick={onClose}
            className="md:hidden p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
          {items.map((item) => {
            const Icon = item.icon;
            const isActive = currentPath === item.path;
            return (
              <Link
                key={item.id}
                href={item.path}
                onClick={onClose}
                className={`w-full flex items-center px-4 py-3 rounded-xl transition-all duration-200 group ${
                  isActive
                    ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20 shadow-[0_0_15px_rgba(59,130,246,0.1)]'
                    : 'text-gray-400 hover:text-white hover:bg-gray-800'
                }`}>
                <Icon
                  className={`w-5 h-5 mr-3 transition-colors ${
                    isActive ? 'text-blue-400' : 'text-gray-500 group-hover:text-white'
                  }`}
                />
                <span className="font-medium text-sm">{item.label}</span>
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-gray-800 bg-gray-900/50 backdrop-blur-sm">
          <button
            onClick={onLogout}
            className="w-full flex items-center px-4 py-3 text-red-400 hover:text-red-300 hover:bg-red-900/10 rounded-xl transition-colors border border-transparent hover:border-red-900/30">
            <LogOut className="w-5 h-5 mr-3" />
            <span className="font-medium text-sm">Logout</span>
          </button>
        </div>
      </aside>
    </>
  );
};

export default Sidebar;
