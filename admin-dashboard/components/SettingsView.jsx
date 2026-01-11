import React, { useState } from 'react';
import { Key, Copy, Plus, Trash2, Server } from 'lucide-react';

const SettingsView = ({ keys, onCreateKey, onDeleteKey }) => {
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [createdKey, setCreatedKey] = useState(null);

  const handleCreate = async (e) => {
    e.preventDefault();
    const result = await onCreateKey(newKeyName);
    if (result) {
      setCreatedKey(result); // { apiKey: {...}, key: "..." }
      setNewKeyName('');
    } else {
      setShowKeyModal(false);
    }
  };

  const closeKeyModal = () => {
    setShowKeyModal(false);
    setCreatedKey(null);
  };

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      {/* Server Info */}
      <section>
        <div>
          <h2 className="text-3xl font-bold text-white mb-2">Settings</h2>
          <p className="text-gray-400">Configure your OTAShip server.</p>
        </div>
      </section>

      {/* API Keys */}
      <section className="bg-gray-900/50 backdrop-blur-sm border border-gray-800 rounded-xl overflow-hidden shadow-xl">
        <div className="p-4 sm:p-6 border-b border-gray-800 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div>
            <h3 className="text-lg font-semibold text-white flex items-center">
              <Key className="w-5 h-5 mr-2 text-yellow-500" />
              API Keys
            </h3>
            <p className="text-sm text-gray-400 mt-1">Manage access tokens for CLI and CI/CD.</p>
          </div>
          <button
            onClick={() => setShowKeyModal(true)}
            className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors font-medium shadow-lg shadow-blue-500/20">
            <Plus className="w-4 h-4 mr-2" />
            Create Key
          </button>
        </div>

        <div className="divide-y divide-gray-800">
          {keys.map((k) => (
            <div
              key={k.id}
              className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-gray-800/40 transition-colors group">
              <div className="min-w-0 flex-1">
                <p className="font-medium text-white">{k.name}</p>
                <div className="flex flex-wrap items-center gap-2 mt-1">
                  <code className="text-xs bg-gray-950 px-2 py-0.5 rounded text-blue-400 border border-gray-800 font-mono">
                    {k.prefix}••••••••
                  </code>
                  <span className="text-xs text-gray-500">
                    • Created {new Date(k.createdAt).toLocaleDateString()}
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-4">
                <div className="text-left sm:text-right">
                  <span className="text-xs text-gray-500 uppercase tracking-wider block">
                    Last Used
                  </span>
                  <span className="text-sm text-gray-300">
                    {k.lastUsedAt ? new Date(k.lastUsedAt).toLocaleDateString() : 'Never'}
                  </span>
                </div>
                <button
                  onClick={() => onDeleteKey(k.id)}
                  className="p-2 text-gray-500 hover:text-red-400 hover:bg-red-900/10 rounded-lg transition-colors"
                  title="Revoke Key">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
          {keys.length === 0 && (
            <div className="p-8 text-center text-gray-500">
              No API keys found. Create one to use the CLI.
            </div>
          )}
        </div>
      </section>

      {/* Server Config (Static for now) */}
      <section className="bg-gray-900/50 backdrop-blur-sm border border-gray-800 rounded-xl p-6 shadow-xl opacity-50">
        <div className="flex items-center gap-2 mb-4">
          <Server className="w-5 h-5 text-gray-500" />
          <h3 className="text-lg font-semibold text-gray-400">Server Configuration</h3>
        </div>
        <p className="text-gray-500 text-sm">
          Server settings are currently managed via environment variables.
        </p>
      </section>

      {/* Create Key Modal */}
      {showKeyModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-gray-900 border border-gray-800 rounded-xl w-full max-w-md p-6 shadow-2xl relative animate-in zoom-in-95 duration-200">
            {createdKey ? (
              <div className="text-center space-y-6">
                <div className="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center mx-auto">
                  <Key className="w-8 h-8 text-green-500" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-white mb-2">API Key Created</h3>
                  <p className="text-gray-400 text-sm">
                    Copy this key now. You won't see it again!
                  </p>
                </div>

                <div className="bg-black/50 border border-gray-800 rounded-lg p-4 flex items-center gap-2 group relative">
                  <code className="flex-1 font-mono text-blue-400 text-sm break-all">
                    {createdKey.key}
                  </code>
                  <button
                    onClick={() => {
                      navigator.clipboard.writeText(createdKey.key);
                    }}
                    className="p-2 hover:bg-white/10 rounded-lg transition-colors text-gray-400 hover:text-white"
                    title="Copy to clipboard">
                    <Copy className="w-4 h-4" />
                  </button>
                </div>

                <button
                  onClick={closeKeyModal}
                  className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors font-medium">
                  Done
                </button>
              </div>
            ) : (
              <form onSubmit={handleCreate} className="space-y-4">
                <h3 className="text-xl font-bold text-white mb-6">Create API Key</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Name</label>
                  <input
                    type="text"
                    required
                    autoFocus
                    value={newKeyName}
                    onChange={(e) => setNewKeyName(e.target.value)}
                    className="w-full bg-gray-950 border border-gray-800 rounded-lg px-4 py-2 text-white focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 outline-none"
                    placeholder="e.g. CI/CD Pipeline"
                  />
                </div>
                <div className="flex gap-3 mt-6">
                  <button
                    type="button"
                    onClick={() => setShowKeyModal(false)}
                    className="flex-1 px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg transition-colors">
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors font-medium">
                    Create
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default SettingsView;
