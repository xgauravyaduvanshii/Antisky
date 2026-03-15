'use client';

import { useState, useEffect } from 'react';

interface Server {
  id: string; name: string; hostname: string; ip_address: string; port: number;
  region: string; status: string; os_info?: string; docker_version?: string;
  cpu_cores?: number; ram_mb?: number; disk_gb?: number;
  metrics?: { cpu_percent: number; ram_used_mb: number; ram_total_mb: number; disk_used_gb: number; disk_total_gb: number; active_containers: number };
}

interface User {
  id: string; email: string; name: string; role: string; is_banned: boolean;
  login_count: number; created_at: string;
}

interface PlatformStats {
  total_users: number; total_servers: number; online_servers: number;
  total_projects: number; total_deployments: number; active_builds: number;
}

export default function AdminDashboard() {
  const [page, setPage] = useState('dashboard');
  const [servers, setServers] = useState<Server[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<PlatformStats>({ total_users: 0, total_servers: 0, online_servers: 0, total_projects: 0, total_deployments: 0, active_builds: 0 });
  const [selectedServer, setSelectedServer] = useState<Server | null>(null);
  const [userSearch, setUserSearch] = useState('');

  useEffect(() => {
    // Demo data — in production, fetched from server-manager API
    setStats({ total_users: 2847, total_servers: 12, online_servers: 10, total_projects: 856, total_deployments: 14283, active_builds: 3 });
    setServers([
      { id: '1', name: 'us-east-1a-worker-01', hostname: 'ip-10-0-1-42', ip_address: '52.23.148.92', port: 8090, region: 'us-east-1', status: 'online', os_info: 'Ubuntu 22.04', docker_version: '24.0.7', cpu_cores: 4, ram_mb: 8192, disk_gb: 100, metrics: { cpu_percent: 42.5, ram_used_mb: 5120, ram_total_mb: 8192, disk_used_gb: 38, disk_total_gb: 100, active_containers: 8 }},
      { id: '2', name: 'us-east-1b-worker-02', hostname: 'ip-10-0-2-18', ip_address: '54.89.12.31', port: 8090, region: 'us-east-1', status: 'online', os_info: 'Ubuntu 22.04', docker_version: '24.0.7', cpu_cores: 8, ram_mb: 16384, disk_gb: 200, metrics: { cpu_percent: 71.2, ram_used_mb: 12288, ram_total_mb: 16384, disk_used_gb: 92, disk_total_gb: 200, active_containers: 15 }},
      { id: '3', name: 'eu-west-1a-edge-01', hostname: 'ip-10-1-1-55', ip_address: '34.249.88.12', port: 8090, region: 'eu-west-1', status: 'online', os_info: 'Ubuntu 22.04', docker_version: '24.0.7', cpu_cores: 2, ram_mb: 4096, disk_gb: 50, metrics: { cpu_percent: 18.7, ram_used_mb: 1024, ram_total_mb: 4096, disk_used_gb: 12, disk_total_gb: 50, active_containers: 4 }},
      { id: '4', name: 'ap-south-1a-builder-01', hostname: 'ip-10-2-1-33', ip_address: '13.235.44.78', port: 8090, region: 'ap-south-1', status: 'offline', os_info: 'Ubuntu 22.04', cpu_cores: 16, ram_mb: 32768, disk_gb: 500, metrics: { cpu_percent: 0, ram_used_mb: 0, ram_total_mb: 32768, disk_used_gb: 128, disk_total_gb: 500, active_containers: 0 }},
    ]);
    setUsers([
      { id: 'u1', email: 'john@example.com', name: 'John Developer', role: 'user', is_banned: false, login_count: 142, created_at: '2026-01-15T10:00:00Z' },
      { id: 'u2', email: 'jane@startup.io', name: 'Jane Smith', role: 'user', is_banned: false, login_count: 89, created_at: '2026-02-01T14:00:00Z' },
      { id: 'u3', email: 'admin@antisky.app', name: 'Platform Admin', role: 'super_admin', is_banned: false, login_count: 512, created_at: '2025-12-01T09:00:00Z' },
      { id: 'u4', email: 'spammer@bad.com', name: 'Bad Actor', role: 'user', is_banned: true, login_count: 3, created_at: '2026-03-10T16:00:00Z' },
    ]);
  }, []);

  const getBarColor = (pct: number) => pct > 80 ? 'red' : pct > 50 ? 'yellow' : 'green';

  const filteredUsers = users.filter(u =>
    userSearch === '' || u.email.toLowerCase().includes(userSearch.toLowerCase()) || u.name.toLowerCase().includes(userSearch.toLowerCase())
  );

  return (
    <div className="admin-layout">
      {/* Sidebar */}
      <aside className="admin-sidebar">
        <div className="admin-sidebar-logo">
          <div className="admin-sidebar-logo-icon">A</div>
          <span className="admin-sidebar-logo-text">Antisky Admin</span>
        </div>
        <nav className="admin-nav">
          <div className="admin-nav-section">
            <div className="admin-nav-title">Platform</div>
            {[
              { id: 'dashboard', icon: '📊', label: 'Dashboard' },
              { id: 'servers', icon: '🖥️', label: 'Servers', count: servers.length },
              { id: 'users', icon: '👥', label: 'Users', count: stats.total_users },
              { id: 'deployments', icon: '🚀', label: 'Deployments', count: stats.total_deployments },
            ].map(item => (
              <div key={item.id} className={`admin-nav-link ${page === item.id ? 'active' : ''}`} onClick={() => { setPage(item.id); setSelectedServer(null); }}>
                <span style={{ fontSize: 16 }}>{item.icon}</span>
                {item.label}
                {item.count !== undefined && <span className="count">{item.count}</span>}
              </div>
            ))}
          </div>
          <div className="admin-nav-section">
            <div className="admin-nav-title">Infrastructure</div>
            {[
              { id: 'terminal', icon: '⌨️', label: 'Terminal' },
              { id: 'billing', icon: '💳', label: 'Billing' },
              { id: 'logs', icon: '📋', label: 'Logs' },
              { id: 'cluster', icon: '⚙️', label: 'Cluster Config' },
            ].map(item => (
              <div key={item.id} className={`admin-nav-link ${page === item.id ? 'active' : ''}`} onClick={() => setPage(item.id)}>
                <span style={{ fontSize: 16 }}>{item.icon}</span>
                {item.label}
              </div>
            ))}
          </div>
        </nav>
      </aside>

      {/* Main */}
      <main className="admin-main">
        <header className="admin-header">
          <h2 style={{ fontSize: 15, fontWeight: 600 }}>
            {page === 'dashboard' && '🏠 Platform Overview'}
            {page === 'servers' && (selectedServer ? `🖥️ ${selectedServer.name}` : '🖥️ Server Management')}
            {page === 'users' && '👥 User Management'}
            {page === 'deployments' && '🚀 All Deployments'}
            {page === 'terminal' && '⌨️ Remote Terminal'}
            {page === 'billing' && '💳 Billing & Revenue'}
            {page === 'logs' && '📋 System Logs'}
            {page === 'cluster' && '⚙️ Cluster Configuration'}
          </h2>
          <div style={{ display: 'flex', gap: 8 }}>
            <span style={{ fontSize: 12, color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 4 }}>
              <span className="server-status-dot" style={{ background: 'var(--color-success)', width: 6, height: 6, borderRadius: '50%', display: 'inline-block' }} />
              {stats.online_servers}/{stats.total_servers} servers online
            </span>
          </div>
        </header>

        <div className="admin-content">
          {/* === DASHBOARD === */}
          {page === 'dashboard' && (
            <>
              <div className="stats-row">
                {[
                  { label: 'Total Users', value: stats.total_users.toLocaleString(), color: 'admin' },
                  { label: 'Servers Online', value: `${stats.online_servers}/${stats.total_servers}`, color: 'admin' },
                  { label: 'Total Projects', value: stats.total_projects.toLocaleString(), color: 'admin' },
                  { label: 'Deployments', value: stats.total_deployments.toLocaleString(), color: 'admin' },
                  { label: 'Active Builds', value: stats.active_builds.toString(), color: 'admin' },
                ].map((s, i) => (
                  <div key={i} className="stat-card">
                    <div className="label">{s.label}</div>
                    <div className={`value ${s.color}`}>{s.value}</div>
                  </div>
                ))}
              </div>

              <div className="page-header">
                <div><h3 className="page-title">Server Fleet</h3><p className="page-subtitle">Real-time server status and metrics</p></div>
              </div>
              <div className="server-grid">
                {servers.map(sv => (
                  <div key={sv.id} className="server-card" onClick={() => { setSelectedServer(sv); setPage('servers'); }}>
                    <div className="server-card-header">
                      <strong style={{ fontSize: 14 }}>{sv.name}</strong>
                      <span className={`server-status ${sv.status}`}>
                        <span className="server-status-dot" />
                        {sv.status}
                      </span>
                    </div>
                    <div className="server-info">
                      <div className="server-info-item"><span className="label">IP</span><br /><span className="val">{sv.ip_address}</span></div>
                      <div className="server-info-item"><span className="label">Region</span><br /><span className="val">{sv.region}</span></div>
                      <div className="server-info-item"><span className="label">Cores</span><br /><span className="val">{sv.cpu_cores}</span></div>
                      <div className="server-info-item"><span className="label">RAM</span><br /><span className="val">{sv.ram_mb ? `${(sv.ram_mb/1024).toFixed(0)}GB` : '—'}</span></div>
                    </div>
                    {sv.metrics && (
                      <>
                        <div className="metric-bar">
                          <div className="bar-header"><span>CPU</span><span>{sv.metrics.cpu_percent.toFixed(1)}%</span></div>
                          <div className="bar-track"><div className={`bar-fill ${getBarColor(sv.metrics.cpu_percent)}`} style={{ width: `${sv.metrics.cpu_percent}%` }} /></div>
                        </div>
                        <div className="metric-bar">
                          <div className="bar-header"><span>RAM</span><span>{((sv.metrics.ram_used_mb/sv.metrics.ram_total_mb)*100).toFixed(0)}%</span></div>
                          <div className="bar-track"><div className={`bar-fill ${getBarColor((sv.metrics.ram_used_mb/sv.metrics.ram_total_mb)*100)}`} style={{ width: `${(sv.metrics.ram_used_mb/sv.metrics.ram_total_mb)*100}%` }} /></div>
                        </div>
                        <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 8 }}>
                          🐳 {sv.metrics.active_containers} containers
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            </>
          )}

          {/* === SERVERS === */}
          {page === 'servers' && !selectedServer && (
            <>
              <div className="page-header">
                <div><h3 className="page-title">All Servers</h3><p className="page-subtitle">Manage your server fleet</p></div>
                <button className="btn btn-admin">+ Add Server</button>
              </div>
              <div className="server-grid">
                {servers.map(sv => (
                  <div key={sv.id} className="server-card" onClick={() => setSelectedServer(sv)}>
                    <div className="server-card-header">
                      <strong>{sv.name}</strong>
                      <span className={`server-status ${sv.status}`}><span className="server-status-dot" />{sv.status}</span>
                    </div>
                    <div className="server-info">
                      <div className="server-info-item"><span className="label">IP</span><br /><span className="val">{sv.ip_address}</span></div>
                      <div className="server-info-item"><span className="label">Region</span><br /><span className="val">{sv.region}</span></div>
                    </div>
                    {sv.metrics && (
                      <div className="metric-bar">
                        <div className="bar-header"><span>CPU</span><span>{sv.metrics.cpu_percent.toFixed(1)}%</span></div>
                        <div className="bar-track"><div className={`bar-fill ${getBarColor(sv.metrics.cpu_percent)}`} style={{ width: `${sv.metrics.cpu_percent}%` }} /></div>
                      </div>
                    )}
                    <div style={{ display: 'flex', gap: 6, marginTop: 12 }}>
                      <button className="btn btn-ghost btn-sm" onClick={(e) => { e.stopPropagation(); }}>Terminal</button>
                      <button className="btn btn-ghost btn-sm">Restart</button>
                      <button className="btn btn-ghost btn-sm" style={{ color: 'var(--color-danger)' }}>Decommission</button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

          {page === 'servers' && selectedServer && (
            <>
              <div className="page-header">
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <button className="btn btn-ghost btn-sm" onClick={() => setSelectedServer(null)}>← Back</button>
                    <h3 className="page-title">{selectedServer.name}</h3>
                    <span className={`server-status ${selectedServer.status}`}><span className="server-status-dot" />{selectedServer.status}</span>
                  </div>
                  <p className="page-subtitle">{selectedServer.ip_address} · {selectedServer.region} · {selectedServer.os_info}</p>
                </div>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button className="btn btn-ghost">Send Command</button>
                  <button className="btn btn-ghost">View Logs</button>
                  <button className="btn btn-admin">Open Terminal</button>
                </div>
              </div>

              {selectedServer.metrics && (
                <div className="stats-row" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
                  <div className="stat-card">
                    <div className="label">CPU Usage</div>
                    <div className="value admin">{selectedServer.metrics.cpu_percent.toFixed(1)}%</div>
                    <div className="metric-bar" style={{ marginTop: 8 }}>
                      <div className="bar-track"><div className={`bar-fill ${getBarColor(selectedServer.metrics.cpu_percent)}`} style={{ width: `${selectedServer.metrics.cpu_percent}%` }} /></div>
                    </div>
                  </div>
                  <div className="stat-card">
                    <div className="label">RAM Usage</div>
                    <div className="value admin">{(selectedServer.metrics.ram_used_mb/1024).toFixed(1)} GB</div>
                    <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>of {(selectedServer.metrics.ram_total_mb/1024).toFixed(0)} GB</div>
                  </div>
                  <div className="stat-card">
                    <div className="label">Disk Usage</div>
                    <div className="value admin">{selectedServer.metrics.disk_used_gb.toFixed(0)} GB</div>
                    <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>of {selectedServer.metrics.disk_total_gb.toFixed(0)} GB</div>
                  </div>
                  <div className="stat-card">
                    <div className="label">Containers</div>
                    <div className="value admin">{selectedServer.metrics.active_containers}</div>
                    <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>running</div>
                  </div>
                </div>
              )}

              <h4 style={{ fontSize: 16, fontWeight: 600, margin: '20px 0 12px' }}>Terminal</h4>
              <div className="terminal-container" style={{ padding: 16, fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: '#50fa7b', lineHeight: 1.6 }}>
                <div>$ docker ps</div>
                <div style={{ color: '#f8f8f2' }}>CONTAINER ID   IMAGE                STATUS</div>
                <div style={{ color: '#f8f8f2' }}>a1b2c3d4e5f6   antisky-agent        Up 2 hours</div>
                <div style={{ color: '#f8f8f2' }}>f6e5d4c3b2a1   antisky-terminal     Up 2 hours</div>
                <div style={{ color: '#f8f8f2' }}>1a2b3c4d5e6f   user-app-nextjs      Up 45 minutes</div>
                <div>$ <span className="pulse">█</span></div>
              </div>
            </>
          )}

          {/* === USERS === */}
          {page === 'users' && (
            <>
              <div className="page-header">
                <div><h3 className="page-title">All Users</h3><p className="page-subtitle">{stats.total_users.toLocaleString()} registered users</p></div>
                <input className="search-input" placeholder="Search users..." value={userSearch} onChange={e => setUserSearch(e.target.value)} />
              </div>
              <table className="user-table">
                <thead>
                  <tr><th>User</th><th>Role</th><th>Status</th><th>Logins</th><th>Joined</th><th>Actions</th></tr>
                </thead>
                <tbody>
                  {filteredUsers.map(u => (
                    <tr key={u.id}>
                      <td>
                        <div style={{ fontWeight: 600 }}>{u.name}</div>
                        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{u.email}</div>
                      </td>
                      <td><span style={{ fontSize: 12, padding: '2px 8px', borderRadius: 12, background: u.role === 'super_admin' ? 'var(--color-admin-glow)' : 'var(--bg-elevated)', color: u.role === 'super_admin' ? 'var(--color-admin)' : 'var(--text-secondary)' }}>{u.role}</span></td>
                      <td>{u.is_banned ? <span style={{ color: 'var(--color-danger)' }}>🚫 Banned</span> : <span style={{ color: 'var(--color-success)' }}>✓ Active</span>}</td>
                      <td>{u.login_count}</td>
                      <td style={{ fontSize: 13, color: 'var(--text-muted)' }}>{new Date(u.created_at).toLocaleDateString()}</td>
                      <td>
                        <div style={{ display: 'flex', gap: 4 }}>
                          <button className="btn btn-ghost btn-sm">👤 Login As</button>
                          {!u.is_banned ? (
                            <button className="btn btn-ghost btn-sm" style={{ color: 'var(--color-danger)' }}>Ban</button>
                          ) : (
                            <button className="btn btn-ghost btn-sm" style={{ color: 'var(--color-success)' }}>Unban</button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          {/* === TERMINAL === */}
          {page === 'terminal' && (
            <>
              <div className="page-header">
                <div><h3 className="page-title">Remote Terminal</h3><p className="page-subtitle">Connect to any server in the fleet</p></div>
                <select className="search-input" style={{ width: 200 }}>
                  {servers.filter(s => s.status === 'online').map(s => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="terminal-container" style={{ height: 500, padding: 16, fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: '#50fa7b', lineHeight: 1.6 }}>
                <div style={{ color: '#6272a4' }}>{'# Connected to us-east-1a-worker-01 (52.23.148.92)'}</div>
                <div style={{ color: '#6272a4' }}>{'# Type commands below. Ctrl+C to disconnect.'}</div>
                <div>&nbsp;</div>
                <div><span style={{ color: '#ff79c6' }}>root@worker-01</span>:<span style={{ color: '#8be9fd' }}>~</span>$ uptime</div>
                <div style={{ color: '#f8f8f2' }}> 14:03:35 up 47 days, 3:22, 1 user, load average: 0.42, 0.38, 0.35</div>
                <div>&nbsp;</div>
                <div><span style={{ color: '#ff79c6' }}>root@worker-01</span>:<span style={{ color: '#8be9fd' }}>~</span>$ free -h</div>
                <div style={{ color: '#f8f8f2' }}>              total        used        free</div>
                <div style={{ color: '#f8f8f2' }}>Mem:          8.0Gi       5.0Gi       3.0Gi</div>
                <div style={{ color: '#f8f8f2' }}>Swap:         2.0Gi       0.0Bi       2.0Gi</div>
                <div>&nbsp;</div>
                <div><span style={{ color: '#ff79c6' }}>root@worker-01</span>:<span style={{ color: '#8be9fd' }}>~</span>$ <span className="pulse">█</span></div>
              </div>
            </>
          )}

          {/* === BILLING === */}
          {page === 'billing' && (
            <>
              <div className="page-header"><div><h3 className="page-title">Billing & Revenue</h3></div></div>
              <div className="stats-row" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
                <div className="stat-card"><div className="label">MRR</div><div className="value admin">$14,280</div></div>
                <div className="stat-card"><div className="label">Active Subscriptions</div><div className="value admin">712</div></div>
                <div className="stat-card"><div className="label">Free Users</div><div className="value admin">2,135</div></div>
                <div className="stat-card"><div className="label">Churn Rate</div><div className="value admin">2.1%</div></div>
              </div>
              <div style={{ background: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-lg)', padding: 20 }}>
                <h4 style={{ fontSize: 15, fontWeight: 600, marginBottom: 12 }}>Plan Distribution</h4>
                {[
                  { name: 'Free', count: 2135, pct: 75, color: 'var(--text-muted)' },
                  { name: 'Hobby ($5/mo)', count: 423, pct: 15, color: 'var(--color-info)' },
                  { name: 'Pro ($20/mo)', count: 256, pct: 9, color: 'var(--color-accent)' },
                  { name: 'Enterprise', count: 33, pct: 1, color: 'var(--color-admin)' },
                ].map(plan => (
                  <div key={plan.name} style={{ marginBottom: 12 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, marginBottom: 4 }}>
                      <span>{plan.name}</span>
                      <span style={{ color: 'var(--text-muted)' }}>{plan.count} users ({plan.pct}%)</span>
                    </div>
                    <div className="bar-track" style={{ height: 8 }}>
                      <div style={{ height: '100%', width: `${plan.pct}%`, background: plan.color, borderRadius: 4 }} />
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

          {page === 'cluster' && (
            <>
              <div className="page-header"><div><h3 className="page-title">Cluster Configuration</h3></div></div>
              <div style={{ background: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-lg)', padding: 20 }}>
                {[
                  { key: 'platform.name', value: 'Antisky', desc: 'Platform name' },
                  { key: 'platform.version', value: '1.0.0', desc: 'Platform version' },
                  { key: 'cluster.max_servers', value: '100', desc: 'Maximum servers' },
                  { key: 'build.max_concurrent', value: '10', desc: 'Max concurrent builds' },
                  { key: 'build.timeout_minutes', value: '30', desc: 'Build timeout' },
                  { key: 'deployment.auto_ssl', value: 'true', desc: 'Auto SSL provisioning' },
                ].map(c => (
                  <div key={c.key} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 0', borderBottom: '1px solid var(--border-color)' }}>
                    <div>
                      <div style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, fontWeight: 600 }}>{c.key}</div>
                      <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{c.desc}</div>
                    </div>
                    <input className="search-input" style={{ width: 200, textAlign: 'right' }} defaultValue={c.value} />
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
