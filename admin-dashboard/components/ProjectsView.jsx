import React, { useState } from 'react';
import { Plus, Box, Trash2, Copy } from 'lucide-react';

const ProjectsView = ({ projects, onCreate, onDelete, apiBase }) => {
  const [showModal, setShowModal] = useState(false);
  const [newProject, setNewProject] = useState({ slug: '', name: '', description: '' });

  const handleSubmit = (e) => {
    e.preventDefault();
    onCreate(newProject);
    setShowModal(false);
    setNewProject({ slug: '', name: '', description: '' });
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold text-white">Projects</h2>
          <p className="text-gray-400">Organize your apps and services.</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors font-medium shadow-lg shadow-blue-500/20">
          <Plus className="w-5 h-5 mr-2" />
          New Project
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {projects.map((p) => (
          <div
            key={p.slug}
            className="group bg-gray-900/50 backdrop-blur-sm border border-gray-800 rounded-xl p-6 hover:border-blue-500/30 hover:bg-gray-800/60 transition-all duration-300 relative">
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 bg-blue-500/10 rounded-lg text-blue-400 group-hover:text-white group-hover:bg-blue-500 transition-colors">
                <Box className="w-6 h-6" />
              </div>
              <button
                onClick={() => onDelete(p.slug)}
                className="text-gray-500 hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 p-2 hover:bg-red-500/10 rounded"
                title="Delete Project">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
            <h3 className="text-lg font-bold text-white mb-1 group-hover:text-blue-400 transition-colors">
              {p.name}
            </h3>
            <code className="text-xs font-mono text-gray-500 bg-gray-950 px-2 py-1 rounded border border-gray-800">
              {p.slug}
            </code>

            {/* Manifest URL Section */}
            <div className="mt-4 pt-4 border-t border-gray-800">
              <p className="text-[10px] text-gray-500 mb-1 uppercase font-semibold tracking-wider">
                Manifest URL
              </p>
              <div className="flex items-center bg-black/40 rounded border border-gray-800 p-2 group/url hover:border-blue-500/30 transition-colors">
                <code className="flex-1 text-xs text-blue-400 truncate font-mono select-all">
                  {`${apiBase}/api/${p.slug}/manifest`}
                </code>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(`${apiBase}/api/${p.slug}/manifest`);
                  }}
                  className="ml-2 p-1.5 text-gray-500 hover:text-white bg-gray-800/50 hover:bg-gray-700 rounded transition-all"
                  title="Copy URL">
                  <Copy className="w-3 h-3" />
                </button>
              </div>
            </div>

            <p className="text-gray-400 text-sm mt-4 line-clamp-2">
              {p.description || 'No description provided.'}
            </p>
          </div>
        ))}

        {/* Empty State */}
        {projects.length === 0 && (
          <div className="col-span-full py-12 text-center border-2 border-dashed border-gray-800 rounded-xl">
            <Box className="w-12 h-12 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-medium text-gray-400">No projects yet</h3>
            <p className="text-gray-500 mb-6">Create your first project to get started.</p>
            <button
              onClick={() => setShowModal(true)}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors">
              Create Project
            </button>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-gray-900 border border-gray-800 rounded-xl w-full max-w-md p-6 shadow-2xl relative animate-in zoom-in-95 duration-200">
            <h3 className="text-xl font-bold text-white mb-6">Create New Project</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">
                  Slug (app.json slug)
                </label>
                <input
                  type="text"
                  required
                  value={newProject.slug}
                  onChange={(e) => setNewProject({ ...newProject, slug: e.target.value })}
                  className="w-full bg-gray-950 border border-gray-800 rounded-lg px-4 py-2 text-white focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 outline-none"
                  placeholder="my-cool-app"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Name</label>
                <input
                  type="text"
                  required
                  value={newProject.name}
                  onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                  className="w-full bg-gray-950 border border-gray-800 rounded-lg px-4 py-2 text-white focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 outline-none"
                  placeholder="My Cool App"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Description</label>
                <textarea
                  rows="3"
                  value={newProject.description}
                  onChange={(e) => setNewProject({ ...newProject, description: e.target.value })}
                  className="w-full bg-gray-950 border border-gray-800 rounded-lg px-4 py-2 text-white focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 outline-none resize-none"
                  placeholder="What does this app do?"
                />
              </div>
              <div className="flex gap-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
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
          </div>
        </div>
      )}
    </div>
  );
};
export default ProjectsView;
