'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: string; name: string; email: string; avatar_url?: string; role?: string; created_at?: string;
}

export default function Dashboard() {
  const [user, setUser] = useState<User | null>(null);
  const [page, setPage] = useState('overview');
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    const stored = localStorage.getItem('user');
    if (!token) { router.push('/login'); return; }
    if (stored) {
      setUser(JSON.parse(stored));
    } else {
      fetch(`${process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081'}/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
      }).then(r => r.json()).then(u => { setUser(u); localStorage.setItem('user', JSON.stringify(u)); }).catch(() => router.push('/login'));
    }
  }, [router]);

  const logout = () => { localStorage.clear(); router.push('/login'); };
  const avatar = user?.avatar_url || user?.name?.[0] || 'U';

  const projects = [
    { id: '1', name: 'my-nextjs-app', framework: 'Next.js', domain: 'my-app.antisky.app', status: 'ready', lastDeploy: '2 hours ago', branch: 'main' },
    { id: '2', name: 'api-backend', framework: 'Go', domain: 'api.antisky.app', status: 'ready', lastDeploy: '5 hours ago', branch: 'main' },
    { id: '3', name: 'landing-page', framework: 'Static', domain: 'landing.antisky.app', status: 'building', lastDeploy: 'Just now', branch: 'feat/redesign' },
  ];

  const deployments = [
    { id: 'd1', project: 'my-nextjs-app', status: 'ready', commit: 'fix: auth flow', branch: 'main', duration: '42s', time: '2h ago', url: 'my-app.antisky.app' },
    { id: 'd2', project: 'landing-page', status: 'building', commit: 'feat: hero section', branch: 'feat/redesign', duration: '—', time: 'Just now', url: '—' },
    { id: 'd3', project: 'api-backend', status: 'ready', commit: 'chore: deps update', branch: 'main', duration: '18s', time: '5h ago', url: 'api.antisky.app' },
    { id: 'd4', project: 'my-nextjs-app', status: 'failed', commit: 'test: broken build', branch: 'dev', duration: '12s', time: '1d ago', url: '—' },
  ];

  if (!user) return <div className="auth-page"><div className="auth-bg"/><div style={{ color: 'var(--text-muted)' }}>Loading...</div></div>;

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
                {n.badge && <span className="badge">{n.badge}</span>}
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
          <span style={{ fontSize: 14, fontWeight: 600 }}>
            {page === 'overview' && '📊 Overview'}
            {page === 'projects' && '📁 Projects'}
            {page === 'deployments' && '🚀 Deployments'}
            {page === 'domains' && '🌐 Domains'}
            {page === 'profile' && '👤 Profile'}
            {page === 'settings' && '⚙️ Settings'}
            {page === 'billing' && '💳 Billing'}
          </span>
          <button className="btn btn-ghost btn-sm" onClick={logout}>Sign Out</button>
        </header>

        <div className="content fade-in" key={page}>
          {/* OVERVIEW */}
          {page === 'overview' && (
            <>
              <div className="stat-grid">
                <div className="stat-card"><div className="stat-label">Projects</div><div className="stat-value">{projects.length}</div></div>
                <div className="stat-card"><div className="stat-label">Deployments</div><div className="stat-value">{deployments.length}</div></div>
                <div className="stat-card"><div className="stat-label">Domains</div><div className="stat-value">3</div></div>
                <div className="stat-card"><div className="stat-label">Build Time Avg</div><div className="stat-value">34s</div></div>
              </div>
              <div className="page-header"><div><h2 className="page-title">Recent Deployments</h2></div></div>
              {deployments.map(d => (
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
            </>
          )}

          {/* PROJECTS */}
          {page === 'projects' && (
            <>
              <div className="page-header">
                <div><h2 className="page-title">Your Projects</h2><p className="page-desc">All your deployed projects</p></div>
                <button className="btn btn-primary btn-sm">+ New Project</button>
              </div>
              <div className="card-grid">
                {projects.map(p => (
                  <div key={p.id} className="card" style={{ cursor: 'pointer' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                      <strong style={{ fontSize: 15 }}>{p.name}</strong>
                      <span className={`badge badge-${p.status === 'ready' ? 'success' : 'warning'}`}>{p.status}</span>
                    </div>
                    <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 8 }}>
                      <span style={{ fontFamily: 'var(--mono)' }}>{p.domain}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: 'var(--text-muted)' }}>
                      <span>{p.framework}</span>
                      <span>{p.lastDeploy}</span>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

          {/* DEPLOYMENTS */}
          {page === 'deployments' && (
            <>
              <div className="page-header"><div><h2 className="page-title">All Deployments</h2></div></div>
              <table className="table">
                <thead><tr><th>Status</th><th>Project</th><th>Commit</th><th>Branch</th><th>Duration</th><th>Time</th></tr></thead>
                <tbody>
                  {deployments.map(d => (
                    <tr key={d.id}>
                      <td><span className={`badge badge-${d.status === 'ready' ? 'success' : d.status === 'building' ? 'warning' : 'danger'}`}>{d.status}</span></td>
                      <td style={{ fontWeight: 600 }}>{d.project}</td>
                      <td style={{ fontFamily: 'var(--mono)', fontSize: 13 }}>{d.commit}</td>
                      <td style={{ fontFamily: 'var(--mono)', fontSize: 13 }}>{d.branch}</td>
                      <td>{d.duration}</td>
                      <td style={{ color: 'var(--text-muted)' }}>{d.time}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          {/* DOMAINS */}
          {page === 'domains' && (
            <>
              <div className="page-header">
                <div><h2 className="page-title">Custom Domains</h2></div>
                <button className="btn btn-primary btn-sm">+ Add Domain</button>
              </div>
              <table className="table">
                <thead><tr><th>Domain</th><th>Project</th><th>SSL</th><th>Status</th></tr></thead>
                <tbody>
                  {[
                    { domain: 'my-app.antisky.app', project: 'my-nextjs-app', ssl: true, status: 'active' },
                    { domain: 'api.antisky.app', project: 'api-backend', ssl: true, status: 'active' },
                    { domain: 'mysite.com', project: 'landing-page', ssl: false, status: 'pending DNS' },
                  ].map((d, i) => (
                    <tr key={i}>
                      <td style={{ fontFamily: 'var(--mono)', fontWeight: 600 }}>{d.domain}</td>
                      <td>{d.project}</td>
                      <td>{d.ssl ? <span className="badge badge-success">✓ SSL</span> : <span className="badge badge-warning">Pending</span>}</td>
                      <td><span className={`badge badge-${d.status === 'active' ? 'success' : 'warning'}`}>{d.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          {/* PROFILE */}
          {page === 'profile' && (
            <>
              <div className="page-header"><div><h2 className="page-title">Your Profile</h2></div></div>
              <div className="card profile-card">
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
                      <input className="form-input" defaultValue={user.email} disabled style={{ opacity: 0.6 }} />
                    </div>
                  </div>
                  <button className="btn btn-primary btn-sm" style={{ marginTop: 8 }}>Save Changes</button>
                </div>
              </div>

              <div className="card" style={{ marginTop: 16 }}>
                <h4 style={{ fontSize: 15, fontWeight: 600, marginBottom: 12 }}>Connected Accounts</h4>
                {[
                  { name: 'GitHub', icon: '🐙', connected: false },
                  { name: 'Google', icon: '🔵', connected: true },
                  { name: 'GitLab', icon: '🦊', connected: false },
                ].map(a => (
                  <div key={a.name} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid var(--border)' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span style={{ fontSize: 20 }}>{a.icon}</span>
                      <span style={{ fontWeight: 500 }}>{a.name}</span>
                    </div>
                    {a.connected ? (
                      <span className="badge badge-success">Connected</span>
                    ) : (
                      <button className="btn btn-ghost btn-sm">Connect</button>
                    )}
                  </div>
                ))}
              </div>
            </>
          )}

          {/* SETTINGS */}
          {page === 'settings' && (
            <>
              <div className="page-header"><div><h2 className="page-title">Settings</h2></div></div>
              <div className="card" style={{ marginBottom: 16 }}>
                <h4 style={{ fontSize: 15, fontWeight: 600, marginBottom: 16 }}>General</h4>
                <div className="form-group">
                  <label className="form-label">Organization Name</label>
                  <input className="form-input" defaultValue="Personal" style={{ maxWidth: 400 }} />
                </div>
                <div className="form-group">
                  <label className="form-label">Default Branch</label>
                  <input className="form-input" defaultValue="main" style={{ maxWidth: 400 }} />
                </div>
                <button className="btn btn-primary btn-sm">Save Settings</button>
              </div>
              <div className="card">
                <h4 style={{ fontSize: 15, fontWeight: 600, marginBottom: 16, color: 'var(--danger)' }}>Danger Zone</h4>
                <button className="btn btn-danger btn-sm">Delete Account</button>
              </div>
            </>
          )}

          {/* BILLING */}
          {page === 'billing' && (
            <>
              <div className="page-header"><div><h2 className="page-title">Billing</h2></div></div>
              <div className="stat-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
                <div className="stat-card"><div className="stat-label">Current Plan</div><div className="stat-value" style={{ fontSize: 22 }}>Free</div></div>
                <div className="stat-card"><div className="stat-label">Builds This Month</div><div className="stat-value">24</div></div>
                <div className="stat-card"><div className="stat-label">Bandwidth</div><div className="stat-value">1.2 GB</div></div>
              </div>
              <div className="card-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
                {[
                  { name: 'Free', price: '$0', features: ['3 Projects', '100 Builds/mo', '1 GB Bandwidth', 'Community Support'] },
                  { name: 'Pro', price: '$20', features: ['Unlimited Projects', 'Unlimited Builds', '100 GB Bandwidth', 'Priority Support', 'Custom Domains', 'Team Members'] },
                  { name: 'Enterprise', price: 'Custom', features: ['Everything in Pro', 'Dedicated Servers', 'SLA 99.99%', '24/7 Support', 'SSO/SAML', 'Audit Logs'] },
                ].map(plan => (
                  <div key={plan.name} className="card" style={{ textAlign: 'center' }}>
                    <h3 style={{ fontSize: 18, fontWeight: 700, marginBottom: 4 }}>{plan.name}</h3>
                    <div style={{ fontSize: 28, fontWeight: 800, background: 'var(--gradient)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', margin: '8px 0' }}>{plan.price}<span style={{ fontSize: 14, fontWeight: 400 }}>/mo</span></div>
                    <div style={{ textAlign: 'left', marginTop: 16 }}>
                      {plan.features.map(f => (
                        <div key={f} style={{ fontSize: 13, padding: '6px 0', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: 6 }}>
                          <span style={{ color: 'var(--success)' }}>✓</span> {f}
                        </div>
                      ))}
                    </div>
                    <button className={`btn ${plan.name === 'Free' ? 'btn-ghost' : 'btn-primary'}`} style={{ marginTop: 16 }}>
                      {plan.name === 'Free' ? 'Current Plan' : 'Upgrade'}
                    </button>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  );
}
