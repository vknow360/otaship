import React from 'react';
import { GitBranch, Bell } from 'lucide-react';

const Overview = ({ stats, updates }) => {
  if (!stats) return <div className="text-gray-400 p-8">Loading stats...</div>;

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div>
        <h2 className="text-3xl font-bold text-white mb-2">Welcome Back</h2>
        <p className="text-gray-400">Here's what's happening with your OTA updates.</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Active Updates"
          value={updates.filter((u) => u.isActive).length}
          icon={<GitBranch className="w-5 h-5 text-orange-400" />}
          gradient="from-orange-500/10 to-orange-500/5"
          borderColor="border-orange-500/20"
        />
      </div>

      {/* Recent Updates (Simplified List) */}
      <div className="bg-gray-900/50 backdrop-blur-sm rounded-xl border border-gray-800 p-6 shadow-xl">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-semibold text-white flex items-center">
            <Bell className="w-5 h-5 mr-2 text-yellow-500" />
            Recent Activity
          </h3>
        </div>
        {/* Limit to 5 updates */}
        <div className="space-y-4">
          {updates.slice(0, 5).map((u) => (
            <div
              key={u.id}
              className="flex items-center justify-between p-4 bg-gray-800/50 rounded-lg border border-gray-700/50 hover:border-gray-600 transition-colors group">
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm text-blue-400 group-hover:text-blue-300 transition-colors">
                    {u.runtimeVersion}
                  </span>
                  <span className="text-gray-400 text-sm">•</span>
                  <span className="text-white font-medium">{u.projectSlug}</span>
                </div>
                <div className="text-sm text-gray-500 mt-1 font-mono">
                  {u.updateId?.substring(0, 8)}... uploaded on{' '}
                  {new Date(u.createdAt).toLocaleDateString()}
                </div>
              </div>
              <div className="text-right">
                <div className="text-xl font-bold text-white tabular-nums">
                  {u.downloads?.toLocaleString() || 0}
                </div>
                <div className="text-xs text-gray-500 uppercase tracking-wider">Downloads</div>
              </div>
            </div>
          ))}
          {updates.length === 0 && (
            <div className="text-gray-500 text-center py-4">No recent activity</div>
          )}
        </div>
      </div>
    </div>
  );
};

function StatCard({ title, value, icon, gradient, borderColor }) {
  return (
    <div
      className={`relative overflow-hidden bg-gray-900/40 rounded-xl p-6 border ${borderColor || 'border-gray-800'} shadow-lg group hover:bg-gray-800/60 transition-all duration-300`}>
      <div
        className={`absolute inset-0 bg-gradient-to-br ${gradient} opacity-20 group-hover:opacity-30 transition-opacity`}
      />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <p className="text-gray-400 text-sm font-medium">{title}</p>
          <div className="p-2 bg-gray-800/80 rounded-lg group-hover:bg-gray-700 transition-colors border border-gray-700/50">
            {icon}
          </div>
        </div>
        <p className="text-3xl font-bold text-white tracking-tight tabular-nums">{value || 0}</p>
      </div>
    </div>
  );
}

export default Overview;
