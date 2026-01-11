import React, { useState } from 'react';
import { Search, Filter, Trash2, Smartphone, Globe, GitCommit } from 'lucide-react';

const UpdatesView = ({ updates, onDelete }) => {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('all'); // all, production, staging

  const filtered = updates.filter((u) => {
    // Handle the access to updateId and id carefully as DB model uses updateId but frontend might use id for key
    const idStr = u.updateId || '';
    const matchesSearch =
      u.projectSlug.toLowerCase().includes(search.toLowerCase()) ||
      idStr.includes(search) ||
      u.runtimeVersion.includes(search);
    const matchesFilter = filter === 'all' || u.channel === filter;
    return matchesSearch && matchesFilter;
  });

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h2 className="text-3xl font-bold text-white">Releases</h2>
          <p className="text-gray-400">Manage and track your OTA deployments.</p>
        </div>
        <div className="flex gap-3 w-full md:w-auto">
          <div className="relative flex-1 md:w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Search projects, IDs..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-gray-900 border border-gray-800 rounded-lg pl-10 pr-4 py-2 text-sm text-gray-300 focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 transition-all outline-none"
            />
          </div>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-gray-900 border border-gray-800 rounded-lg px-4 py-2 text-sm text-gray-300 focus:border-blue-500/50 outline-none appearance-none cursor-pointer">
            <option value="all">All Channels</option>
            <option value="production">Production</option>
            <option value="staging">Staging</option>
          </select>
        </div>
      </div>

      <div className="bg-gray-900/50 backdrop-blur-sm rounded-xl border border-gray-800 overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-gray-800 bg-gray-900/80">
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Project / ID
                </th>
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Runtime
                </th>
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Channel
                </th>
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Platform
                </th>
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="p-4 text-xs font-medium text-gray-500 uppercase tracking-wider text-right">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {filtered.map((u) => (
                <tr key={u.id} className="group hover:bg-gray-800/40 transition-colors">
                  <td className="p-4">
                    <div className="font-medium text-white">{u.projectSlug}</div>
                    <div className="text-xs text-gray-500 font-mono mt-1 flex items-center">
                      <GitCommit className="w-3 h-3 mr-1" />
                      {u.updateId ? u.updateId.substring(0, 8) + '...' : 'Unknown'}
                    </div>
                  </td>
                  <td className="p-4">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
                      v{u.runtimeVersion}
                    </span>
                  </td>
                  <td className="p-4">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize border ${
                        u.channel === 'production'
                          ? 'bg-green-500/10 text-green-400 border-green-500/20'
                          : 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20'
                      }`}>
                      {u.channel}
                    </span>
                  </td>
                  <td className="p-4 text-gray-400">
                    <div className="flex items-center">
                      {u.platform === 'ios' && <Smartphone className="w-4 h-4 mr-2" />}
                      {u.platform === 'android' && <Smartphone className="w-4 h-4 mr-2" />}
                      {u.platform === 'all' && <Globe className="w-4 h-4 mr-2" />}
                      <span className="capitalize">{u.platform}</span>
                    </div>
                  </td>
                  <td className="p-4 text-sm text-gray-400">
                    {new Date(u.createdAt).toLocaleDateString()}
                  </td>
                  <td className="p-4 text-right">
                    <button
                      onClick={() => onDelete(u.id)}
                      className="p-2 text-gray-500 hover:text-red-400 hover:bg-red-900/10 rounded-lg transition-colors"
                      title="Delete Update">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan="6" className="p-8 text-center text-gray-500">
                    No updates found matching your filters.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default UpdatesView;
