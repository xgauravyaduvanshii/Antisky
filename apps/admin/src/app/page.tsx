'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [servers, setServers] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [theme, setTheme] = useState('dark');
  const [authenticated, setAuthenticated] = useState(false);
  const [adminKey, setAdminKey] = useState('');
  const [error, setError] = useState('');
  
  const router = useRouter();

  useEffect(() => {
    // Check local auth for admin key
    const currentKey = localStorage.getItem('admin_key');
    if (currentKey) {
      setAdminKey(currentKey);
      setAuthenticated(true);
      fetchData(currentKey);
    }
  }, []);

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    // Verify admin key via a test API call
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    fetch(`${apiBase}/api/v1/admin/stats`, {
      headers: { 'X-Cluster-Secret': adminKey }
    })
    .then(res => {
      if (res.ok) {
        localStorage.setItem('admin_key', adminKey);
        setAuthenticated(true);
        fetchData(adminKey);
      } else {
        setError('Invalid Cluster Secret / Admin Key');
      }
    })
    .catch(() => setError('Connection failed to Server Manager'));
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_key');
    setAuthenticated(false);
    setAdminKey('');
  };

  const fetchData = async (key: string) => {
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    try {
      const statsRes = await fetch(`${apiBase}/api/v1/admin/stats`, { headers: { 'X-Cluster-Secret': key }});
      if (statsRes.ok) setStats(await statsRes.json());

      const serversRes = await fetch(`${apiBase}/api/v1/admin/servers`, { headers: { 'X-Cluster-Secret': key }});
      if (serversRes.ok) setServers((await serversRes.json()).data || []);

      const usersRes = await fetch(`${apiBase}/api/v1/admin/users`, { headers: { 'X-Cluster-Secret': key }});
      if (usersRes.ok) setUsers((await usersRes.json()).data || []);
    } catch(e) {
      console.error(e);
    }
  };

  const executeServerCommand = async (serverId: string) => {
    const cmd = prompt('Enter bash command to execute on server:');
    if (!cmd) return;
    
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    try {
      const res = await fetch(`${apiBase}/api/v1/admin/servers/${serverId}/command`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Cluster-Secret': adminKey },
        body: JSON.stringify({ command: cmd })
      });
      const data = await res.json();
      alert(`Output:\n\n${data.output || data.error}`);
    } catch(e) {
      alert('Command failed');
    }
  };

  const banUser = async (userId: string, isBanned: boolean) => {
    if(!confirm(`Are you sure you want to ${isBanned ? 'unban' : 'ban'} this user?`)) return;
    
    // Optimistic UI update
    setUsers(users.map(u => u.id === userId ? { ...u, is_banned: !isBanned } : u));
    
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    await fetch(`${apiBase}/api/v1/admin/users/${userId}/${isBanned ? 'unban' : 'ban'}`, {
      method: 'POST',
      headers: { 'X-Cluster-Secret': adminKey }
    });
  };

  if (!authenticated) {
    return (
      <div className="auth-page">
        <div className="auth-bg" />
        <div className="auth-card fade-in">
          <div className="auth-logo">
            <div className="auth-logo-icon" style={{ background: '#000' }}>🔒</div>
            <span className="auth-logo-text" style={{ color: '#fff', WebkitTextFillColor: '#fff'}}>Admin Access</span>
          </div>
          <h1 className="auth-title">System Control</h1>
          <p className="auth-subtitle">Enter cluster secret to access infrastructure control</p>
          
          {error && <div className="error-box">{error}</div>}

          <form onSubmit={handleLogin}>
            <div className="form-group">
              <label className="form-label">Cluster Secret Key</label>
              <input 
                className="form-input" 
                type="password" 
                value={adminKey} 
                onChange={e => setAdminKey(e.target.value)} 
                placeholder="antisky-cluster-secret-2026"
                required 
              />
            </div>
            <button className="btn btn-primary" type="submit">Authenticate</button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">
            <div className="sidebar-logo-icon" style={{ background: 'var(--danger)' }}>👑</div>
            <span className="sidebar-logo-text" style={{ color: 'var(--danger)', WebkitTextFillColor: 'var(--danger)'}}>Antisky Admin</span>
          </div>
        </div>
        
        <nav className="sidebar-nav">
          <div className="nav-section">
            <div className="nav-title">Platform</div>
            {[
              { id: 'dashboard', icon: '📊', label: 'Dashboard' },
              { id: 'servers', icon: '🖥️', label: 'Servers Fleet', badge: servers.length },
              { id: 'users', icon: '👥', label: 'Users', badge: users.length },
              { id: 'deployments', icon: '🚀', label: 'All Deployments' }
            ].map(tab => (
              <div key={tab.id} className={`nav-item ${activeTab === tab.id ? 'active' : ''}`} onClick={() => setActiveTab(tab.id)}>
                <span className="icon">{tab.icon}</span> {tab.label}
                {tab.badge !== undefined && tab.badge > 0 && <span className="badge">{tab.badge}</span>}
              </div>
            ))}
          </div>

          <div className="nav-section">
            <div className="nav-title">Infrastructure</div>
            {[
              { id: 'terminal', icon: '⌨️', label: 'Global Terminal' },
              { id: 'billing', icon: '💳', label: 'Stripe Billing' },
              { id: 'logs', icon: '📋', label: 'System Logs' },
            ].map(tab => (
              <div key={tab.id} className={`nav-item ${activeTab === tab.id ? 'active' : ''}`} onClick={() => setActiveTab(tab.id)}>
                <span className="icon">{tab.icon}</span> {tab.label}
              </div>
            ))}
          </div>
        </nav>
      </aside>

      <main className="main">
        <header className="topbar">
          <span style={{ fontSize: 15, fontWeight: 600 }}>System Control Plane</span>
          <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
            <span className="badge badge-success">● {servers.filter(s => s.status === 'online').length} / {servers.length} servers online</span>
            <button className="btn btn-ghost btn-sm" onClick={handleLogout}>Lock Session</button>
          </div>
        </header>

        <div className="content fade-in" key={activeTab}>
          
          {/* DASHBOARD */}
          {activeTab === 'dashboard' && (
            <>
              <div className="stat-grid slide-in">
                <div className="stat-card">
                  <div className="stat-label">Total Users</div>
                  <div className="stat-value">{stats?.total_users || users.length || 0}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Servers (Online / Total)</div>
                  <div className="stat-value">{stats?.online_servers || servers.filter(s=>s.status==='online').length || 0} / {stats?.total_servers || servers.length || 0}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Total Projects</div>
                  <div className="stat-value">{stats?.total_projects || 0}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Deployments</div>
                  <div className="stat-value">{stats?.total_deployments || 0}</div>
                </div>
              </div>

              <div className="page-header slide-in" style={{ marginTop: 24 }}>
                <h3 className="page-title">Quick Actions</h3>
              </div>
              
              <div className="card-grid">
                <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  <h4 style={{ fontSize: 16 }}>Provision Server</h4>
                  <p style={{ fontSize: 13, color: 'var(--text-muted)' }}>Generate a curl script to auto-provision a new Ubuntu VPS to the Antisky cluster.</p>
                  <button className="btn btn-primary" onClick={() => alert(`curl -sSL https://get.antisky.app | bash -s -- --secret ${adminKey}`)}>Get Install Script</button>
                </div>
                <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  <h4 style={{ fontSize: 16 }}>Flush Cache</h4>
                  <p style={{ fontSize: 13, color: 'var(--text-muted)' }}>Clear the global Redis cache across all routing edges.</p>
                  <button className="btn btn-warning">Flush Redis</button>
                </div>
              </div>
            </>
          )}

          {/* SERVERS */}
          {activeTab === 'servers' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Server Fleet</h2>
                <button className="btn btn-primary btn-sm">+ Add Server</button>
              </div>

              {servers.length === 0 ? (
                <div className="card empty-state">
                  <div className="icon">🖥️</div>
                  <h3>No servers registered</h3>
                  <p>Run the provision script on a new VPS to see it appear here automatically.</p>
                </div>
              ) : (
                <table className="table card" style={{ padding: 0, overflow: 'hidden' }}>
                  <thead>
                    <tr>
                      <th>Hostname</th>
                      <th>IP Address</th>
                      <th>Region</th>
                      <th>Resources</th>
                      <th>Status</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {servers.map(s => (
                      <tr key={s.id}>
                        <td style={{ fontWeight: 600 }}>{s.hostname}</td>
                        <td style={{ fontFamily: 'var(--mono)', fontSize: 13 }}>{s.ip_address}</td>
                        <td>{s.region}</td>
                        <td style={{ fontSize: 13, color: 'var(--text-muted)' }}>{s.cpu_cores} cores / {s.ram_gb}GB memory</td>
                        <td><span className={`badge badge-${s.status === 'online' ? 'success' : 'danger'}`}>● {s.status}</span></td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            <button className="btn btn-ghost btn-sm" onClick={() => executeServerCommand(s.id)}>Terminal</button>
                            <button className="btn btn-danger btn-sm">Drain</button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* USERS */}
          {activeTab === 'users' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">User Management</h2>
                <input type="text" className="search" placeholder="Search by email..." />
              </div>

              {users.length === 0 ? (
                <div className="card empty-state">
                  <div className="icon">👥</div>
                  <h3>No users yet</h3>
                </div>
              ) : (
                <table className="table card" style={{ padding: 0, overflow: 'hidden' }}>
                  <thead>
                    <tr>
                      <th>User</th>
                      <th>Email</th>
                      <th>Role</th>
                      <th>Joined</th>
                      <th>Status</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u: any) => (
                      <tr key={u.id}>
                        <td style={{ fontWeight: 600, display: 'flex', alignItems: 'center', gap: 10 }}>
                          <img src={u.avatar_url || `https://ui-avatars.com/api/?name=${u.name}&background=random`} style={{ width: 32, height: 32, borderRadius: '50%' }} alt=""/>
                          {u.name}
                        </td>
                        <td style={{ color: 'var(--text-muted)' }}>{u.email}</td>
                        <td><span className={`badge badge-${u.role === 'admin' ? 'info' : 'warning'}`}>{u.role || 'user'}</span></td>
                        <td style={{ fontSize: 13 }}>{new Date(u.created_at).toLocaleDateString()}</td>
                        <td>
                          {u.is_banned ? <span className="badge badge-danger">Banned</span> : <span className="badge badge-success">Active</span>}
                        </td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            <button className="btn btn-ghost btn-sm">Impersonate</button>
                            <button className={`btn btn-sm ${u.is_banned ? 'btn-success' : 'btn-danger'}`} onClick={() => banUser(u.id, u.is_banned)}>
                              {u.is_banned ? 'Unban' : 'Ban'}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* OTHERS */}
          {['deployments', 'terminal', 'billing', 'logs'].includes(activeTab) && (
            <div className="card empty-state slide-in">
              <div className="icon">🚧</div>
              <h3>Under Construction</h3>
              <p>The {activeTab} view is fully functional via API but the UI implementation is pending in the next sprint.</p>
            </div>
          )}

        </div>
      </main>
    </div>
  );
}
