'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [servers, setServers] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [deployments, setDeployments] = useState<any[]>([]);
  const [theme, setTheme] = useState('dark');
  const [authenticated, setAuthenticated] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [adminKey, setAdminKey] = useState('');
  const [token, setToken] = useState('');
  const [error, setError] = useState('');
  const [userSearch, setUserSearch] = useState('');
  const [terminalCmd, setTerminalCmd] = useState('');
  const [terminalOutput, setTerminalOutput] = useState<string[]>([]);
  const [terminalServerId, setTerminalServerId] = useState('');
  
  const router = useRouter();

  useEffect(() => {
    // Daily rotating theme
    const days = ['dark', 'ocean', 'forest', 'sunset', 'light', 'dark', 'ocean'];
    const todayTheme = days[new Date().getDay()];
    setTheme(todayTheme);
    document.documentElement.setAttribute('data-theme', todayTheme);

    // Check local auth for admin key
    const currentKey = localStorage.getItem('admin_key');
    const storedToken = localStorage.getItem('admin_token');
    if (currentKey) {
      setAdminKey(currentKey);
      if (storedToken) setToken(storedToken);
      setAuthenticated(true);
      fetchData(currentKey);
    }
  }, []);

  const changeTheme = (t: string) => {
    setTheme(t);
    document.documentElement.setAttribute('data-theme', t);
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      // 1. Authenticate with Auth Service
      const authRes = await fetch(`${process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081'}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });
      const authData = await authRes.json();
      
      if (!authRes.ok) {
        throw new Error(authData.error || 'Invalid email or password');
      }

      // 2. Verify Admin Role
      if (authData.user.role !== 'admin' && authData.user.role !== 'super_admin') {
        throw new Error('Unauthorized: This account does not have admin privileges');
      }

      // 3. Verify Cluster Secret
      const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
      const statsRes = await fetch(`${apiBase}/api/v1/admin/stats`, {
        headers: { 'X-Cluster-Secret': adminKey }
      });

      if (!statsRes.ok) {
        throw new Error('Invalid Cluster Secret Key');
      }

      // Success
      localStorage.setItem('admin_key', adminKey);
      localStorage.setItem('admin_token', authData.access_token);
      setToken(authData.access_token);
      setAuthenticated(true);
      fetchData(adminKey);

    } catch (err: any) {
      setError(err.message || 'Login failed');
    }
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
      if (serversRes.ok) {
        const sData = await serversRes.json();
        setServers(sData.servers || sData.data || sData || []);
      }

      const usersRes = await fetch(`${apiBase}/api/v1/admin/users`, { headers: { 'X-Cluster-Secret': key }});
      if (usersRes.ok) {
        const uData = await usersRes.json();
        setUsers(uData.users || uData.data || uData || []);
      }

      // Fetch deployments from control-plane
      try {
        const cpBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082';
        const dRes = await fetch(`${cpBase}/api/v1/admin/deployments`, { headers: { 'X-Cluster-Secret': key }});
        if (dRes.ok) {
          const dData = await dRes.json();
          setDeployments(dData.deployments || dData.data || dData || []);
        }
      } catch(e) { /* control-plane may not have this endpoint yet */ }
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

  const flushCache = async () => {
    if(!confirm('Are you sure you want to flush the entire Redis cache? This may cause a temporary spike in database load.')) return;
    try {
      const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
      const res = await fetch(`${apiBase}/api/v1/admin/cache/flush`, {
        method: 'POST',
        headers: { 'X-Cluster-Secret': adminKey }
      });
      if (res.ok) alert('Cache flushed successfully');
      else alert('Failed to flush cache');
    } catch(e) { alert('Network error'); }
  };

  const drainServer = async (serverId: string) => {
    if(!confirm('Are you sure you want to drain this server? No new deployments will be scheduled here.')) return;
    try {
      const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
      const res = await fetch(`${apiBase}/api/v1/admin/servers/${serverId}/drain`, {
        method: 'POST',
        headers: { 'X-Cluster-Secret': adminKey }
      });
      if (res.ok) {
        setServers(servers.map(s => s.id === serverId ? { ...s, status: 'draining' } : s));
      } else {
        alert('Failed to drain server');
      }
    } catch(e) { alert('Network error'); }
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

  const deleteUser = async (userId: string, userName: string) => {
    if(!confirm(`Are you sure you want to PERMANENTLY delete ${userName}? This cannot be undone.`)) return;
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    try {
      const res = await fetch(`${apiBase}/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: { 'X-Cluster-Secret': adminKey }
      });
      if (res.ok) {
        setUsers(users.filter(u => u.id !== userId));
      } else {
        alert('Failed to delete user');
      }
    } catch(e) { alert('Network error'); }
  };

  const impersonateUser = async (targetUserId: string) => {
    try {
      const authBase = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081';
      const res = await fetch(`${authBase}/auth/impersonate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ target_user_id: targetUserId })
      });
      const data = await res.json();
      if (res.ok && data.access_token) {
        localStorage.setItem('token', data.access_token);
        if (data.user) localStorage.setItem('user', JSON.stringify(data.user));
        window.open('http://localhost:3000', '_blank');
      } else {
        alert(data.error?.message || data.error || 'Failed to impersonate user');
      }
    } catch (err) {
      console.error(err);
      alert('Error connecting to auth service');
    }
  };

  const runGlobalTerminal = async () => {
    if (!terminalCmd.trim()) return;
    if (!terminalServerId) { alert('Select a server first'); return; }
    setTerminalOutput(prev => [...prev, `$ ${terminalCmd}`]);
    const apiBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
    try {
      const res = await fetch(`${apiBase}/api/v1/admin/servers/${terminalServerId}/command`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Cluster-Secret': adminKey },
        body: JSON.stringify({ command: terminalCmd })
      });
      const data = await res.json();
      setTerminalOutput(prev => [...prev, data.output || data.error || JSON.stringify(data)]);
    } catch(e) {
      setTerminalOutput(prev => [...prev, 'Error: Connection failed']);
    }
    setTerminalCmd('');
  };

  const filteredUsers = users.filter(u =>
    !userSearch || u.name?.toLowerCase().includes(userSearch.toLowerCase()) || u.email?.toLowerCase().includes(userSearch.toLowerCase())
  );

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
            <div className="form-group" style={{ marginBottom: '12px' }}>
              <label className="form-label">Admin Email</label>
              <input 
                className="form-input" 
                type="email" 
                value={email} 
                onChange={e => setEmail(e.target.value)} 
                required 
              />
            </div>
            <div className="form-group" style={{ marginBottom: '12px' }}>
              <label className="form-label">Password</label>
              <input 
                className="form-input" 
                type="password" 
                value={password} 
                onChange={e => setPassword(e.target.value)} 
                required 
              />
            </div>
            <div className="form-group">
              <label className="form-label">Cluster Secret Key</label>
              <input 
                className="form-input" 
                type="password" 
                value={adminKey} 
                onChange={e => setAdminKey(e.target.value)} 
                required 
              />
            </div>
            <button className="btn btn-primary" type="submit" style={{ marginTop: '24px' }}>Authenticate</button>
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
            <div className="theme-picker">
              <div className={`theme-dot ${theme === 'dark' ? 'active' : ''}`} data-t="dark" title="Dark" onClick={() => changeTheme('dark')} />
              <div className={`theme-dot ${theme === 'light' ? 'active' : ''}`} data-t="light" title="Light" onClick={() => changeTheme('light')} />
              <div className={`theme-dot ${theme === 'ocean' ? 'active' : ''}`} data-t="ocean" title="Ocean" onClick={() => changeTheme('ocean')} />
              <div className={`theme-dot ${theme === 'forest' ? 'active' : ''}`} data-t="forest" title="Forest" onClick={() => changeTheme('forest')} />
              <div className={`theme-dot ${theme === 'sunset' ? 'active' : ''}`} data-t="sunset" title="Sunset" onClick={() => changeTheme('sunset')} />
            </div>
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
                  <button className="btn btn-warning" onClick={flushCache}>Flush Redis</button>
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
                            <button className="btn btn-danger btn-sm" onClick={() => drainServer(s.id)}>Drain</button>
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
                <input type="text" className="search" placeholder="Search by email..." value={userSearch} onChange={e => setUserSearch(e.target.value)} />
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
                    {filteredUsers.map((u: any) => (
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
                            <button className="btn btn-ghost btn-sm" onClick={() => impersonateUser(u.id)}>Impersonate</button>
                            <button className={`btn btn-sm ${u.is_banned ? 'btn-success' : 'btn-danger'}`} onClick={() => banUser(u.id, u.is_banned)}>
                              {u.is_banned ? 'Unban' : 'Ban'}
                            </button>
                            <button className="btn btn-danger btn-sm" onClick={() => deleteUser(u.id, u.name)}>Delete</button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* DEPLOYMENTS */}
          {activeTab === 'deployments' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">All Deployments</h2>
              </div>
              {deployments.length === 0 ? (
                <div className="card empty-state">
                  <div className="icon">🚀</div>
                  <h3>No deployments tracked</h3>
                  <p>Deployments will appear here once users push code to their projects.</p>
                </div>
              ) : (
                <table className="table card" style={{ padding: 0, overflow: 'hidden' }}>
                  <thead><tr><th>Project</th><th>Branch</th><th>Commit</th><th>Status</th><th>Created</th></tr></thead>
                  <tbody>
                    {deployments.map((d: any) => (
                      <tr key={d.id}>
                        <td style={{ fontWeight: 600 }}>{d.project_name || d.project_id}</td>
                        <td>{d.branch || 'main'}</td>
                        <td style={{ fontFamily: 'var(--mono)', fontSize: 13 }}>{(d.commit_sha || 'n/a').slice(0,8)}</td>
                        <td><span className={`badge badge-${d.status === 'ready' ? 'success' : d.status === 'building' ? 'warning' : 'danger'}`}>{d.status}</span></td>
                        <td style={{ fontSize: 13 }}>{new Date(d.created_at).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* TERMINAL */}
          {activeTab === 'terminal' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Global Terminal</h2>
                <select className="form-input" style={{ width: 260 }} value={terminalServerId} onChange={e => setTerminalServerId(e.target.value)}>
                  <option value="">Select a server…</option>
                  {servers.map(s => <option key={s.id} value={s.id}>{s.hostname} ({s.ip_address})</option>)}
                </select>
              </div>
              <div className="card" style={{ background: '#0d1117', padding: 0, overflow: 'hidden', fontFamily: 'var(--mono)', fontSize: 13 }}>
                <div style={{ padding: 16, minHeight: 300, maxHeight: 400, overflowY: 'auto' }}>
                  {terminalOutput.length === 0 && <div style={{ color: '#6e7681' }}>Terminal output will appear here…</div>}
                  {terminalOutput.map((line, i) => (
                    <div key={i} style={{ color: line.startsWith('$') ? '#58a6ff' : line.startsWith('Error') ? '#f85149' : '#c9d1d9', marginBottom: 4, whiteSpace: 'pre-wrap' }}>{line}</div>
                  ))}
                </div>
                <div style={{ display: 'flex', borderTop: '1px solid #21262d', alignItems: 'center' }}>
                  <span style={{ padding: '10px 12px', color: '#58a6ff' }}>$</span>
                  <input className="form-input" style={{ background: 'transparent', border: 'none', color: '#c9d1d9', flex: 1, outline: 'none', fontSize: 13 }} 
                    value={terminalCmd} onChange={e => setTerminalCmd(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && runGlobalTerminal()}
                    placeholder="Enter command…" />
                  <button className="btn btn-primary btn-sm" style={{ margin: 8 }} onClick={runGlobalTerminal}>Run</button>
                </div>
              </div>
            </div>
          )}

          {/* BILLING / LOGS */}
          {['billing', 'logs'].includes(activeTab) && (
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
