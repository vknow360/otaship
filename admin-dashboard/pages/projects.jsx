import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import ProjectsView from '../components/ProjectsView';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Projects() {
  const [projects, setProjects] = useState([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const headers = { Authorization: `Bearer ${token}` };
      const res = await fetch(`${API_BASE}/api/admin/projects`, { headers });
      if (res.ok) {
        const data = await res.json();
        setProjects(data.projects || []);
      }
    } catch (err) {
      console.error('Data fetch error:', err);
    }
  };

  const handleCreateProject = async (project) => {
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const res = await fetch(`${API_BASE}/api/admin/projects`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(project),
      });
      if (res.ok) {
        fetchData();
        return true;
      } else {
        const data = await res.json();
        alert(data.error || 'Failed to create project');
      }
    } catch (e) {
      alert('Failed to create project');
    }
    return false;
  };

  const handleDeleteProject = async (slug) => {
    if (!window.confirm(`Delete project "${slug}" and ALL its updates? This is irreversible.`))
      return;
    try {
      const token = localStorage.getItem('otaship_admin_token');
      const res = await fetch(`${API_BASE}/api/admin/projects/${slug}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) fetchData();
      else alert('Failed to delete project');
    } catch (e) {
      alert('Failed to delete project');
    }
  };

  return (
    <Layout>
      <ProjectsView
        projects={projects}
        onCreate={handleCreateProject}
        onDelete={handleDeleteProject}
        apiBase={API_BASE}
      />
    </Layout>
  );
}
