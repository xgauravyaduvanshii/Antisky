'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: string; name: string; email: string; avatar_url?: string; role?: string; created_at?: string;
}

export default function Dashboard() {
  const [user, setUser] = useState<User | null>(null);
  const [page, setPage] = useState('overview');
  const [theme, setTheme] = useState('dark');
  const [projects, setProjects] = useState<any[]>([]);
  const [deployments, setDeployments] = useState<any[]>([]);
  const [orgs, setOrgs] = useState<any[]>([]);
  const router = useRouter();

  useEffect(() => {
    // Theme setup based on day of week for rotating themes (bonus feature)
    const days = ['dark', 'ocean', 'forest', 'sunset', 'light', 'dark', 'ocean'];
    const todayTheme = days[new Date().getDay()];
    setTheme(todayTheme);
    document.documentElement.setAttribute('data-theme', todayTheme);

    const token = localStorage.getItem('token');
    const stored = localStorage.getItem('user');
    if (!token) { router.push('/login'); return; }
    
    // Load User
    if (stored) setUser(JSON.parse(stored));
    fetch(`${process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081'}/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
    }).then(r => r.json()).then(u => { setUser(u); localStorage.setItem('user', JSON.stringify(u)); }).catch(() => router.push('/login'));

    // Load Data
    const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082';
    
    fetch(`${apiBase}/api/v1/projects`, { headers: { Authorization: `Bearer ${token}` }}).then(r => r.json()).then(data => setProjects(data || [])).catch(() => {});
    fetch(`${apiBase}/api/v1/orgs`, { headers: { Authorization: `Bearer ${token}` }}).then(r => r.json()).then(data => setOrgs(data || [])).catch(() => {});
  }, [router]);

  const changeTheme = (t: string) => {
    setTheme(t);
    document.documentElement.setAttribute('data-theme', t);
  };

  const logout = () => { localStorage.clear(); router.push('/login'); };
  const avatar = user?.avatar_url || user?.name?.[0] || 'U';

  const mockDeployments = [
    { id: 'd1', project: 'my-nextjs-app', status: 'ready', commit: 'fix: auth flow', branch: 'main', duration: '42s', time: '2h ago' },
    { id: 'd2', project: 'landing-page', status: 'building', commit: 'feat: hero section', branch: 'feat/redesign', duration: '—', time: 'Just now' },
    { id: 'd3', project: 'api-backend', status: 'ready', commit: 'chore: deps update', branch: 'main', duration: '18s', time: '5h ago' },
  ];

  if (!user) return <div className="auth-page"><div className="auth-bg"/><div style={{ color: 'var(--text-muted)' }}>Loading session...</div></div>;

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">
            <div className="sidebar-logo-icon">A</div>
            <span className="sidebar-logo-text">Antisky</span>
          </div>
        </div>
        <nav className="sidebar-nav">
          <div className="nav-section">
            <div className="nav-title">Dashboard</div>
            {[
              { id: 'overview', icon: '📊', label: 'Overview' },
              { id: 'projects', icon: '📁', label: 'Projects', badge: projects.length },
              { id: 'deployments', icon: '🚀', label: 'Deployments' },
              { id: 'domains', icon: '🌐', label: 'Domains' },
            ].map(n => (
              <div key={n.id} className={`nav-item ${page === n.id ? 'active' : ''}`} onClick={() => setPage(n.id)}>
                <span className="icon">{n.icon}</span> {n.label}
                {n.badge !== undefined && n.badge > 0 && <span className="badge">{n.badge}</span>}
              </div>
            ))}
          </div>
          <div className="nav-section">
            <div className="nav-title">Account</div>
            {[
              { id: 'profile', icon: '👤', label: 'Profile' },
              { id: 'settings', icon: '⚙️', label: 'Settings' },
              { id: 'billing', icon: '💳', label: 'Billing' },
            ].map(n => (
              <div key={n.id} className={`nav-item ${page === n.id ? 'active' : ''}`} onClick={() => setPage(n.id)}>
                <span className="icon">{n.icon}</span> {n.label}
              </div>
            ))}
          </div>
        </nav>
        <div className="sidebar-user">
          <div className="sidebar-avatar">
            {user.avatar_url ? <img src={user.avatar_url} alt="" /> : avatar}
          </div>
          <div className="sidebar-user-info">
            <div className="sidebar-user-name">{user.name}</div>
            <div className="sidebar-user-email">{user.email}</div>
          </div>
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <span style={{ fontSize: 14, fontWeight: 600 }}>
              {page === 'overview' && '📊 Overview'}
              {page === 'projects' && '📁 Projects'}
              {page === 'deployments' && '🚀 Deployments'}
              {page === 'domains' && '🌐 Domains'}
              {page === 'profile' && '👤 Profile'}
              {page === 'settings' && '⚙️ Settings'}
              {page === 'billing' && '💳 Billing'}
            </span>
          </div>
          
          <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
            <div className="theme-picker">
              <div className={`theme-dot ${theme === 'dark' ? 'active' : ''}`} data-t="dark" title="Dark" onClick={() => changeTheme('dark')} />
              <div className={`theme-dot ${theme === 'light' ? 'active' : ''}`} data-t="light" title="Light" onClick={() => changeTheme('light')} />
              <div className={`theme-dot ${theme === 'ocean' ? 'active' : ''}`} data-t="ocean" title="Ocean" onClick={() => changeTheme('ocean')} />
              <div className={`theme-dot ${theme === 'forest' ? 'active' : ''}`} data-t="forest" title="Forest" onClick={() => changeTheme('forest')} />
              <div className={`theme-dot ${theme === 'sunset' ? 'active' : ''}`} data-t="sunset" title="Sunset" onClick={() => changeTheme('sunset')} />
            </div>
            <button className="btn btn-ghost btn-sm" onClick={logout}>Sign Out</button>
          </div>
        </header>

        <div className="content fade-in" key={page}>
          
          {/* OVERVIEW */}
          {page === 'overview' && (
            <>
              <div className="stat-grid slide-in">
                <div className="stat-card"><div className="stat-label">Projects</div><div className="stat-value">{projects.length}</div></div>
                <div className="stat-card"><div className="stat-label">Organizations</div><div className="stat-value">{orgs.length}</div></div>
                <div className="stat-card"><div className="stat-label">Deployments</div><div className="stat-value">{mockDeployments.length}</div></div>
                <div className="stat-card"><div className="stat-label">Active Plan</div><div className="stat-value" style={{ fontSize: 20 }}>Hobby</div></div>
              </div>

              <div className="page-header" style={{ marginTop: 32 }}>
                <div><h2 className="page-title">Recent Deployments</h2></div>
              </div>
              
              <div className="card">
                {mockDeployments.map(d => (
                  <div key={d.id} className="deploy-item">
                    <div className={`deploy-dot ${d.status}`} />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 600, fontSize: 14 }}>{d.project}</div>
                      <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{d.commit} · {d.branch}</div>
                    </div>
                    <span className={`badge badge-${d.status === 'ready' ? 'success' : d.status === 'building' ? 'warning' : 'danger'}`}>{d.status}</span>
                    <span style={{ fontSize: 12, color: 'var(--text-muted)', width: 60, textAlign: 'right' }}>{d.time}</span>
                  </div>
                ))}
              </div>
            </>
          )}

          {/* PROJECTS */}
          {page === 'projects' && (
            <>
              <div className="page-header slide-in">
                <div><h2 className="page-title">Your Projects</h2><p className="page-desc">All deployed applications</p></div>
                <button className="btn btn-primary btn-sm">+ New Project</button>
              </div>
              
              {projects.length === 0 ? (
                <div className="card empty-state">
                  <div className="icon">🚀</div>
                  <h3>No projects yet</h3>
                  <p>Deploy your first app in seconds from a GitHub repository.</p>
                  <button className="btn btn-primary btn-sm" style={{ margin: '20px auto 0' }}>Import Repository</button>
                </div>
              ) : (
                <div className="card-grid">
                  {projects.map(p => (
                    <div key={p.id} className="card" style={{ cursor: 'pointer' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                        <strong style={{ fontSize: 16 }}>{p.name}</strong>
                      </div>
                      <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12 }}>
                        {p.description || 'No description'}
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: 'var(--text-muted)', borderTop: '1px solid var(--border)', paddingTop: 12 }}>
                        <span>Framework: Next.js</span>
                        <span>Updated: {new Date(p.updated_at).toLocaleDateString()}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}

          {/* PROFILE */}
          {page === 'profile' && (
            <div className="slide-in">
              <div className="page-header"><div><h2 className="page-title">Your Profile</h2></div></div>
              <div className="card profile-card" style={{ marginBottom: 20 }}>
                <div className="profile-avatar-lg">
                  {user.avatar_url ? <img src={user.avatar_url} alt="" /> : user.name?.[0] || 'U'}
                </div>
                <div style={{ flex: 1 }}>
                  <h3 style={{ fontSize: 20, fontWeight: 700, marginBottom: 4 }}>{user.name}</h3>
                  <p style={{ fontSize: 14, color: 'var(--text-muted)', marginBottom: 20 }}>{user.email}</p>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                    <div className="form-group">
                      <label className="form-label">Display Name</label>
                      <input className="form-input" defaultValue={user.name} />
                    </div>
                    <div className="form-group">
                      <label className="form-label">Email</label>
                      <input className="form-input" defaultValue={user.email} disabled />
                    </div>
                  </div>
                  <button className="btn btn-primary btn-sm" style={{ marginTop: 8 }}>Save Changes</button>
                </div>
              </div>
            </div>
          )}

          {/* SETTINGS / DANGER ZONE */}
          {page === 'settings' && (
            <div className="slide-in">
              <div className="page-header"><div><h2 className="page-title">Settings</h2></div></div>
              <div className="card" style={{ marginBottom: 16 }}>
                <h4 style={{ fontSize: 16, fontWeight: 600, marginBottom: 16 }}>API Keys</h4>
                <p style={{ fontSize: 14, color: 'var(--text-muted)', marginBottom: 16 }}>Generate API keys for programmatic access to the Antisky CLI and endpoints.</p>
                <button className="btn btn-ghost btn-sm">Generate Key</button>
              </div>
              <div className="card" style={{ border: '1px solid rgba(239,68,68,0.3)' }}>
                <h4 style={{ fontSize: 16, fontWeight: 600, marginBottom: 8, color: 'var(--danger)' }}>Danger Zone</h4>
                <p style={{ fontSize: 14, color: 'var(--text-muted)', marginBottom: 16 }}>Once you delete your account, there is no going back. Please be certain.</p>
                <button className="btn btn-danger btn-sm" onClick={() => {
                  if(confirm('Are you sure you want to delete your account?')) {
                     alert('Demo Mode: Account deletion disabled.');
                  }
                }}>Delete Account</button>
              </div>
            </div>
          )}

        </div>
      </main>
    </div>
  );
}
