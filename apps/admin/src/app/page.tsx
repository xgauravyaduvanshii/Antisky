'use client';
import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';

// --- Types ---
interface LogEntry { id: string; timestamp: string; level: 'info'|'warn'|'error'|'debug'; service: string; message: string; }
interface Notification { id: string; type: 'deploy'|'alert'|'user'|'billing'|'system'; title: string; desc: string; time: string; read: boolean; }
interface ClusterSetting { key: string; value: string; description: string; enabled: boolean; }

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
  // New states
  const [logFilter, setLogFilter] = useState('all');
  const [logSearch, setLogSearch] = useState('');
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [analyticsRange, setAnalyticsRange] = useState('7d');
  const [clusterSettings, setClusterSettings] = useState<ClusterSetting[]>([]);
  const [showModal, setShowModal] = useState<string|null>(null);
  const [modalData, setModalData] = useState<any>({});
  const [billingTab, setBillingTab] = useState('overview');
  const [toast, setToast] = useState<string|null>(null);

  const router = useRouter();
  const serverManagerBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
  const controlPlaneBase = process.env.NEXT_PUBLIC_CONTROL_PLANE_URL || 'http://localhost:8082';
  const authBase = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081';
  const billingBase = 'http://localhost:8087';
  const razorpayKeyId = 'rzp_live_R5GepMnnMQikm8';

  const showToast = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3000); };
  const pickArray = (data: any, key: string) => {
    if (Array.isArray(data?.[key])) return data[key];
    if (Array.isArray(data?.data)) return data.data;
    if (Array.isArray(data)) return data;
    return [];
  };

  useEffect(() => {
    const days = ['dark','ocean','forest','sunset','light','dark','ocean'];
    const todayTheme = days[new Date().getDay()];
    setTheme(todayTheme);
    document.documentElement.setAttribute('data-theme', todayTheme);
    const currentKey = localStorage.getItem('admin_key');
    const storedToken = localStorage.getItem('admin_token');
    if (currentKey) {
      setAdminKey(currentKey);
      if (storedToken) setToken(storedToken);
      setAuthenticated(true);
      fetchData(currentKey);
    }
  }, []);

  // Generate mock logs for demo
  useEffect(() => {
    if (!authenticated) return;
    const services = ['auth','control-plane','build-orchestrator','server-manager','billing'];
    const msgs = [
      'Request processed successfully','Health check passed','Connection pool refreshed',
      'Cache miss for key user:session','JWT token generated','Build queued for project',
      'Deployment completed','Webhook received from GitHub','Rate limit approaching',
      'Database query slow (>2s)','TLS certificate renewed','Worker process spawned',
      'Memory usage at 72%','Stale server detected','Subscription webhook processed'
    ];
    const levels: ('info'|'warn'|'error'|'debug')[] = ['info','info','info','info','debug','info','warn','error','debug','info'];
    const generated: LogEntry[] = [];
    for (let i = 0; i < 50; i++) {
      const d = new Date(Date.now() - Math.random() * 3600000);
      generated.push({
        id: `log-${i}`, timestamp: d.toISOString(),
        level: levels[Math.floor(Math.random() * levels.length)],
        service: services[Math.floor(Math.random() * services.length)],
        message: msgs[Math.floor(Math.random() * msgs.length)]
      });
    }
    generated.sort((a,b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
    setLogs(generated);

    setNotifications([
      { id:'n1', type:'deploy', title:'Deployment Completed', desc:'project-alpha deployed to production successfully', time:'2 min ago', read:false },
      { id:'n2', type:'alert', title:'High Memory Usage', desc:'server-eu-west-01 memory at 89%', time:'15 min ago', read:false },
      { id:'n3', type:'user', title:'New User Registered', desc:'john@example.com signed up', time:'1 hour ago', read:true },
      { id:'n4', type:'billing', title:'Payment Received', desc:'₹4,999 from Org: Acme Corp', time:'3 hours ago', read:true },
      { id:'n5', type:'system', title:'SSL Certificate Renewed', desc:'*.antisky.app certificate auto-renewed', time:'6 hours ago', read:true },
      { id:'n6', type:'deploy', title:'Build Failed', desc:'project-beta build #42 failed on step 3', time:'8 hours ago', read:true },
    ]);

    setClusterSettings([
      { key:'install_script_url', value:'https://get.antisky.app', description:'Base URL for the server provisioning script', enabled:true },
      { key:'maintenance_mode', value:'false', description:'Enable platform-wide maintenance mode', enabled:false },
      { key:'auto_scaling', value:'true', description:'Automatically scale servers based on load', enabled:true },
      { key:'build_concurrency', value:'5', description:'Maximum concurrent builds per server', enabled:true },
      { key:'ssl_auto_provision', value:'true', description:'Auto-provision SSL certificates for custom domains', enabled:true },
      { key:'webhook_retries', value:'3', description:'Number of webhook delivery retries', enabled:true },
      { key:'log_retention_days', value:'30', description:'System log retention period in days', enabled:true },
    ]);
  }, [authenticated]);

  const changeTheme = (t: string) => { setTheme(t); document.documentElement.setAttribute('data-theme', t); };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault(); setError('');
    try {
      const authRes = await fetch(`${authBase}/auth/login`, { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({email,password}) });
      const authData = await authRes.json();
      if (!authRes.ok) throw new Error(authData.error || 'Invalid email or password');
      if (authData.user.role !== 'admin' && authData.user.role !== 'super_admin') throw new Error('Unauthorized: No admin privileges');
      const statsRes = await fetch(`${serverManagerBase}/api/v1/admin/stats`, { headers:{'X-Cluster-Secret':adminKey} });
      if (!statsRes.ok) throw new Error('Invalid Cluster Secret Key');
      localStorage.setItem('admin_key', adminKey);
      localStorage.setItem('admin_token', authData.access_token);
      setToken(authData.access_token);
      setAuthenticated(true);
      fetchData(adminKey);
    } catch (err: any) { setError(err.message || 'Login failed'); }
  };

  const handleLogout = () => { localStorage.removeItem('admin_key'); localStorage.removeItem('admin_token'); setAuthenticated(false); setAdminKey(''); };

  const fetchData = async (key: string) => {
    try {
      const statsRes = await fetch(`${serverManagerBase}/api/v1/admin/stats`, { headers:{'X-Cluster-Secret':key} });
      if (statsRes.ok) { const d = await statsRes.json(); setStats(d && typeof d === 'object' ? d : null); }
      const serversRes = await fetch(`${serverManagerBase}/api/v1/admin/servers`, { headers:{'X-Cluster-Secret':key} });
      if (serversRes.ok) { const d = await serversRes.json(); setServers(pickArray(d,'servers')); }
      const usersRes = await fetch(`${serverManagerBase}/api/v1/admin/users`, { headers:{'X-Cluster-Secret':key} });
      if (usersRes.ok) { const d = await usersRes.json(); setUsers(pickArray(d,'users')); }
      try { const dRes = await fetch(`${controlPlaneBase}/api/v1/admin/deployments`, { headers:{'X-Cluster-Secret':key} }); if (dRes.ok) { const d = await dRes.json(); setDeployments(pickArray(d,'deployments')); } } catch {}
    } catch(e) { console.error(e); }
  };

  const executeServerCommand = async (serverId: string) => {
    const cmd = prompt('Enter bash command:'); if (!cmd) return;
    try {
      const res = await fetch(`${serverManagerBase}/api/v1/admin/servers/${serverId}/command`, { method:'POST', headers:{'Content-Type':'application/json','X-Cluster-Secret':adminKey}, body:JSON.stringify({command:cmd}) });
      const data = await res.json(); alert(`Output:\n\n${data.output || data.error}`);
    } catch { alert('Command failed'); }
  };

  const flushCache = async () => {
    if(!confirm('Flush Redis cache?')) return;
    try { const res = await fetch(`${serverManagerBase}/api/v1/admin/cache/flush`, { method:'POST', headers:{'X-Cluster-Secret':adminKey} }); if(res.ok) showToast('✓ Cache flushed'); else showToast('✗ Failed'); } catch { showToast('✗ Network error'); }
  };

  const drainServer = async (serverId: string) => {
    if(!confirm('Drain this server?')) return;
    try { const res = await fetch(`${serverManagerBase}/api/v1/admin/servers/${serverId}/drain`, { method:'POST', headers:{'X-Cluster-Secret':adminKey} }); if(res.ok) { setServers(servers.map(s => s.id === serverId ? {...s, status:'draining'} : s)); showToast('✓ Server draining'); } else showToast('✗ Failed'); } catch { showToast('✗ Network error'); }
  };

  const banUser = async (userId: string, isBanned: boolean) => {
    if(!confirm(`${isBanned ? 'Unban' : 'Ban'} this user?`)) return;
    setUsers(users.map(u => u.id === userId ? {...u, is_banned: !isBanned} : u));
    await fetch(`${serverManagerBase}/api/v1/admin/users/${userId}/${isBanned ? 'unban' : 'ban'}`, { method:'POST', headers:{'X-Cluster-Secret':adminKey} });
    showToast(isBanned ? '✓ User unbanned' : '✓ User banned');
  };

  const deleteUser = async (userId: string, userName: string) => {
    if(!confirm(`PERMANENTLY delete ${userName}?`)) return;
    try { const res = await fetch(`${serverManagerBase}/api/v1/admin/users/${userId}`, { method:'DELETE', headers:{'X-Cluster-Secret':adminKey} }); if(res.ok) { setUsers(users.filter(u => u.id !== userId)); showToast('✓ User deleted'); } else showToast('✗ Failed'); } catch { showToast('✗ Network error'); }
  };

  const impersonateUser = async (targetUserId: string) => {
    try {
      const res = await fetch(`${authBase}/auth/impersonate`, { method:'POST', headers:{'Content-Type':'application/json','Authorization':`Bearer ${token}`}, body:JSON.stringify({target_user_id:targetUserId}) });
      const data = await res.json();
      if (res.ok && data.access_token) { localStorage.setItem('token', data.access_token); if(data.user) localStorage.setItem('user', JSON.stringify(data.user)); window.open('http://localhost:3000','_blank'); showToast('✓ Impersonating user'); }
      else alert(data.error?.message || data.error || 'Failed to impersonate');
    } catch (err) { console.error(err); alert('Error connecting to auth service'); }
  };

  const runGlobalTerminal = async () => {
    if (!terminalCmd.trim()) return;
    if (!terminalServerId) { showToast('Select a server first'); return; }
    setTerminalOutput(prev => [...prev, `$ ${terminalCmd}`]);
    try {
      const res = await fetch(`${serverManagerBase}/api/v1/admin/servers/${terminalServerId}/command`, { method:'POST', headers:{'Content-Type':'application/json','X-Cluster-Secret':adminKey}, body:JSON.stringify({command:terminalCmd}) });
      const data = await res.json(); setTerminalOutput(prev => [...prev, data.output || data.error || JSON.stringify(data)]);
    } catch { setTerminalOutput(prev => [...prev, 'Error: Connection failed']); }
    setTerminalCmd('');
  };

  const addServer = async () => {
    const installUrl = clusterSettings.find(s => s.key === 'install_script_url')?.value || 'https://get.antisky.app';
    const script = `curl -sSL ${installUrl} | bash -s -- --secret ${adminKey}`;
    setShowModal('add-server');
    setModalData({ script });
  };

  const updateClusterSetting = (key: string, value: string | boolean) => {
    setClusterSettings(prev => prev.map(s => s.key === key ? { ...s, ...(typeof value === 'boolean' ? { enabled: value, value: String(value) } : { value }) } : s));
    showToast(`✓ Setting ${key} updated`);
    // In production: PUT to /api/v1/admin/cluster/{key}
  };

  const filteredUsers = users.filter(u => !userSearch || u.name?.toLowerCase().includes(userSearch.toLowerCase()) || u.email?.toLowerCase().includes(userSearch.toLowerCase()));
  const filteredLogs = logs.filter(l => (logFilter === 'all' || l.level === logFilter) && (!logSearch || l.message.toLowerCase().includes(logSearch.toLowerCase()) || l.service.toLowerCase().includes(logSearch.toLowerCase())));

  // Analytics data
  const chartData = [65,42,78,95,53,87,72,60,88,45,92,70];
  const chartLabels = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  const maxChart = Math.max(...chartData);

  // Plans data
  const plans = [
    { name:'Hobby', price:'Free', priceNum:0, features:['1 Project','100MB Storage','Shared CPU','Community Support','antisky.app subdomain'] },
    { name:'Pro', price:'₹999', priceNum:999, popular:true, features:['10 Projects','10GB Storage','Dedicated CPU','Priority Support','Custom Domains','SSL Certificates','Auto-scaling'] },
    { name:'Business', price:'₹4,999', priceNum:4999, features:['Unlimited Projects','100GB Storage','Multi-region','24/7 Support','Team Management','Advanced Analytics','SLA 99.9%'] },
    { name:'Enterprise', price:'Custom', priceNum:0, features:['Unlimited Everything','Dedicated Infrastructure','Custom SLA','Dedicated Account Manager','On-premise Option','Audit Logs'] },
  ];

  if (!authenticated) {
    return (
      <div className="auth-page">
        <div className="auth-bg" />
        <div className="auth-card fade-in">
          <div className="auth-logo">
            <div className="auth-logo-icon" style={{ background:'#000' }}>🔒</div>
            <span className="auth-logo-text" style={{ color:'#fff', WebkitTextFillColor:'#fff'}}>Admin Access</span>
          </div>
          <h1 className="auth-title">System Control</h1>
          <p className="auth-subtitle">Enter credentials to access infrastructure control</p>
          {error && <div className="error-box">{error}</div>}
          <form onSubmit={handleLogin}>
            <div className="form-group" style={{ marginBottom: '12px' }}>
              <label className="form-label">Admin Email</label>
              <input className="form-input" type="email" value={email} onChange={e => setEmail(e.target.value)} required />
            </div>
            <div className="form-group" style={{ marginBottom: '12px' }}>
              <label className="form-label">Password</label>
              <input className="form-input" type="password" value={password} onChange={e => setPassword(e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Cluster Secret Key</label>
              <input className="form-input" type="password" value={adminKey} onChange={e => setAdminKey(e.target.value)} required />
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
            <div className="sidebar-logo-icon" style={{ background:'var(--danger)' }}>👑</div>
            <span className="sidebar-logo-text" style={{ color:'var(--danger)', WebkitTextFillColor:'var(--danger)'}}>Antisky Admin</span>
          </div>
        </div>
        <nav className="sidebar-nav">
          <div className="nav-section">
            <div className="nav-title">Platform</div>
            {[
              { id:'dashboard', icon:'📊', label:'Dashboard' },
              { id:'servers', icon:'🖥️', label:'Servers Fleet', badge:servers.length },
              { id:'users', icon:'👥', label:'Users', badge:users.length },
              { id:'deployments', icon:'🚀', label:'All Deployments' },
              { id:'analytics', icon:'📈', label:'Analytics' },
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
              { id:'terminal', icon:'⌨️', label:'Global Terminal' },
              { id:'billing', icon:'💳', label:'Razorpay Billing' },
              { id:'logs', icon:'📋', label:'System Logs' },
              { id:'notifications', icon:'🔔', label:'Notifications', badge: notifications.filter(n=>!n.read).length },
              { id:'settings', icon:'⚙️', label:'Settings' },
            ].map(tab => (
              <div key={tab.id} className={`nav-item ${activeTab === tab.id ? 'active' : ''}`} onClick={() => setActiveTab(tab.id)}>
                <span className="icon">{tab.icon}</span> {tab.label}
                {tab.badge !== undefined && tab.badge > 0 && <span className="badge" style={{ background:'var(--danger)', color:'#fff' }}>{tab.badge}</span>}
              </div>
            ))}
          </div>
        </nav>
      </aside>

      <main className="main">
        <header className="topbar">
          <span style={{ fontSize:15, fontWeight:600 }}>System Control Plane</span>
          <div style={{ display:'flex', gap:12, alignItems:'center' }}>
            <div className="theme-picker">
              {['dark','light','ocean','forest','sunset'].map(t => (
                <div key={t} className={`theme-dot ${theme===t?'active':''}`} data-t={t} title={t} onClick={() => changeTheme(t)} />
              ))}
            </div>
            <span className="badge badge-success">● {servers.filter(s => s.status==='online').length} / {servers.length} online</span>
            <button className="btn btn-ghost btn-sm" onClick={() => fetchData(adminKey)}>↻ Refresh</button>
            <button className="btn btn-ghost btn-sm" onClick={handleLogout}>Lock Session</button>
          </div>
        </header>

        <div className="content fade-in" key={activeTab}>

          {/* ═══ DASHBOARD ═══ */}
          {activeTab === 'dashboard' && (
            <>
              <div className="stat-grid slide-in">
                <div className="stat-card"><div className="stat-label">Total Users</div><div className="stat-value">{stats?.total_users || users.length || 0}</div></div>
                <div className="stat-card"><div className="stat-label">Servers Online</div><div className="stat-value">{stats?.online_servers || servers.filter(s=>s.status==='online').length} / {stats?.total_servers || servers.length || 0}</div></div>
                <div className="stat-card"><div className="stat-label">Total Projects</div><div className="stat-value">{stats?.total_projects || 0}</div></div>
                <div className="stat-card"><div className="stat-label">Deployments</div><div className="stat-value">{stats?.total_deployments || deployments.length}</div></div>
                <div className="stat-card"><div className="stat-label">Revenue (MTD)</div><div className="stat-value" style={{fontSize:20}}>₹{stats?.monthly_revenue || '0'}</div></div>
              </div>
              <div style={{ display:'grid', gridTemplateColumns:'2fr 1fr', gap:16 }}>
                <div className="card slide-in">
                  <h4 style={{ fontSize:16, marginBottom:16 }}>Deployment Activity (7 days)</h4>
                  <div className="chart-container">
                    {[28,42,35,67,55,48,62].map((v,i) => (
                      <div key={i} className="chart-bar" style={{ height:`${(v/67)*100}%` }}>
                        <div className="chart-tooltip">{v} deploys</div>
                        <div className="chart-bar-label">{['Mon','Tue','Wed','Thu','Fri','Sat','Sun'][i]}</div>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="card slide-in">
                  <h4 style={{ fontSize:16, marginBottom:12 }}>Quick Actions</h4>
                  <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
                    <button className="btn btn-primary btn-sm" onClick={addServer}>+ Provision Server</button>
                    <button className="btn btn-warning btn-sm" onClick={flushCache}>Flush Cache</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setActiveTab('logs')}>View Logs</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setActiveTab('analytics')}>View Analytics</button>
                  </div>
                </div>
              </div>
              <div className="card slide-in" style={{ marginTop:16 }}>
                <h4 style={{ fontSize:16, marginBottom:12 }}>Recent Activity</h4>
                {notifications.slice(0,5).map(n => (
                  <div key={n.id} className="activity-item">
                    <div className="activity-dot" style={{ background: n.type==='alert'?'var(--warning)':n.type==='deploy'?'var(--success)':'var(--accent)' }} />
                    <div style={{flex:1}}><strong>{n.title}</strong> — {n.desc}</div>
                    <span style={{ fontSize:11, color:'var(--text-muted)', flexShrink:0 }}>{n.time}</span>
                  </div>
                ))}
              </div>
            </>
          )}

          {/* ═══ SERVERS ═══ */}
          {activeTab === 'servers' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Server Fleet</h2>
                <button className="btn btn-primary btn-sm" onClick={addServer}>+ Add Server</button>
              </div>
              {servers.length === 0 ? (
                <div className="card empty-state"><div className="icon">🖥️</div><h3>No servers registered</h3><p>Run the provision script on a new VPS to see it here.</p><button className="btn btn-primary btn-sm" style={{margin:'20px auto 0'}} onClick={addServer}>Get Install Script</button></div>
              ) : (
                <table className="table card" style={{ padding:0, overflow:'hidden' }}>
                  <thead><tr><th>Hostname</th><th>IP Address</th><th>Region</th><th>Resources</th><th>Status</th><th>Actions</th></tr></thead>
                  <tbody>
                    {servers.map(s => (
                      <tr key={s.id}>
                        <td style={{ fontWeight:600 }}>{s.hostname}</td>
                        <td style={{ fontFamily:'var(--mono)', fontSize:13 }}>{s.ip_address}</td>
                        <td>{s.region}</td>
                        <td style={{ fontSize:13, color:'var(--text-muted)' }}>{s.cpu_cores} cores / {s.ram_gb}GB</td>
                        <td><span className={`badge badge-${s.status==='online'?'success':s.status==='draining'?'warning':'danger'}`}>● {s.status}</span></td>
                        <td>
                          <div style={{ display:'flex', gap:8 }}>
                            <button className="btn btn-ghost btn-sm" onClick={() => executeServerCommand(s.id)}>Terminal</button>
                            <button className="btn btn-warning btn-sm" onClick={() => drainServer(s.id)}>Drain</button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* ═══ USERS ═══ */}
          {activeTab === 'users' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">User Management</h2>
                <input type="text" className="search" placeholder="Search users..." value={userSearch} onChange={e => setUserSearch(e.target.value)} />
              </div>
              {users.length === 0 ? (
                <div className="card empty-state"><div className="icon">👥</div><h3>No users yet</h3></div>
              ) : (
                <table className="table card" style={{ padding:0, overflow:'hidden' }}>
                  <thead><tr><th>User</th><th>Email</th><th>Role</th><th>Joined</th><th>Status</th><th>Actions</th></tr></thead>
                  <tbody>
                    {filteredUsers.map((u: any) => (
                      <tr key={u.id}>
                        <td style={{ fontWeight:600, display:'flex', alignItems:'center', gap:10 }}>
                          <img src={u.avatar_url || `https://ui-avatars.com/api/?name=${u.name}&background=random`} style={{ width:32, height:32, borderRadius:'50%' }} alt=""/>
                          {u.name}
                        </td>
                        <td style={{ color:'var(--text-muted)' }}>{u.email}</td>
                        <td><span className={`badge badge-${u.role==='admin'?'info':'warning'}`}>{u.role||'user'}</span></td>
                        <td style={{ fontSize:13 }}>{new Date(u.created_at).toLocaleDateString()}</td>
                        <td>{u.is_banned ? <span className="badge badge-danger">Banned</span> : <span className="badge badge-success">Active</span>}</td>
                        <td>
                          <div style={{ display:'flex', gap:8 }}>
                            <button className="btn btn-ghost btn-sm" onClick={() => impersonateUser(u.id)}>Impersonate</button>
                            <button className={`btn btn-sm ${u.is_banned?'btn-success':'btn-danger'}`} onClick={() => banUser(u.id, u.is_banned)}>{u.is_banned?'Unban':'Ban'}</button>
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

          {/* ═══ DEPLOYMENTS ═══ */}
          {activeTab === 'deployments' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">All Deployments</h2><button className="btn btn-ghost btn-sm" onClick={() => fetchData(adminKey)}>↻ Refresh</button></div>
              {deployments.length === 0 ? (
                <div className="card empty-state"><div className="icon">🚀</div><h3>No deployments tracked</h3><p>Deployments appear here once users push code.</p></div>
              ) : (
                <table className="table card" style={{ padding:0, overflow:'hidden' }}>
                  <thead><tr><th>Project</th><th>Branch</th><th>Commit</th><th>Status</th><th>Created</th></tr></thead>
                  <tbody>{deployments.map((d: any) => (
                    <tr key={d.id}>
                      <td style={{ fontWeight:600 }}>{d.project_name||d.project_id}</td>
                      <td>{d.branch||'main'}</td>
                      <td style={{ fontFamily:'var(--mono)', fontSize:13 }}>{(d.commit_sha||'n/a').slice(0,8)}</td>
                      <td><span className={`badge badge-${d.status==='ready'?'success':d.status==='building'?'warning':'danger'}`}>{d.status}</span></td>
                      <td style={{ fontSize:13 }}>{new Date(d.created_at).toLocaleString()}</td>
                    </tr>
                  ))}</tbody>
                </table>
              )}
            </div>
          )}

          {/* ═══ ANALYTICS ═══ */}
          {activeTab === 'analytics' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Platform Analytics</h2>
                <div className="tab-pills">
                  {['7d','30d','90d','1y'].map(r => (
                    <button key={r} className={`tab-pill ${analyticsRange===r?'active':''}`} onClick={() => setAnalyticsRange(r)}>{r}</button>
                  ))}
                </div>
              </div>
              <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16, marginBottom:16 }}>
                <div className="card">
                  <h4 style={{ fontSize:15, marginBottom:8 }}>Deployments Over Time</h4>
                  <div className="chart-container">
                    {chartData.map((v,i) => (
                      <div key={i} className="chart-bar" style={{ height:`${(v/maxChart)*100}%` }}>
                        <div className="chart-tooltip">{v} deploys</div>
                        <div className="chart-bar-label">{chartLabels[i]}</div>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="card">
                  <h4 style={{ fontSize:15, marginBottom:8 }}>Revenue (₹)</h4>
                  <div className="chart-container">
                    {[12000,18000,15000,22000,28000,25000,31000,27000,35000,30000,38000,42000].map((v,i) => (
                      <div key={i} className="chart-bar" style={{ height:`${(v/42000)*100}%`, background:'linear-gradient(135deg, var(--success), #14b8a6)' }}>
                        <div className="chart-tooltip">₹{v.toLocaleString()}</div>
                        <div className="chart-bar-label">{chartLabels[i]}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr 1fr', gap:16 }}>
                <div className="card">
                  <h4 style={{ fontSize:15, marginBottom:12 }}>Server Utilization</h4>
                  {[{l:'CPU Average',v:68},{l:'Memory',v:72},{l:'Disk I/O',v:45},{l:'Network',v:38}].map(m => (
                    <div key={m.l} className="mini-stat">
                      <div className="mini-stat-info">
                        <div className="mini-stat-label">{m.l}</div>
                        <div className="mini-stat-value">{m.v}%</div>
                        <div className="progress-bar"><div className="progress-bar-fill" style={{ width:`${m.v}%` }} /></div>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="card">
                  <h4 style={{ fontSize:15, marginBottom:12 }}>Top Frameworks</h4>
                  {[{n:'Next.js',c:45},{n:'Node.js',c:28},{n:'Python',c:15},{n:'Go',c:8},{n:'PHP',c:4}].map(f => (
                    <div key={f.n} className="mini-stat">
                      <div className="mini-stat-info">
                        <div style={{display:'flex',justifyContent:'space-between'}}><span className="mini-stat-label">{f.n}</span><span style={{fontSize:13,fontWeight:600}}>{f.c}%</span></div>
                        <div className="progress-bar"><div className="progress-bar-fill" style={{ width:`${f.c}%` }} /></div>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="card">
                  <h4 style={{ fontSize:15, marginBottom:12 }}>Key Metrics</h4>
                  {[{i:'⚡',l:'Avg Build Time',v:'42s'},{i:'🌍',l:'Uptime',v:'99.97%'},{i:'📦',l:'Total Builds',v:'1,247'},{i:'🔄',l:'Daily Active Users',v:'156'}].map(m => (
                    <div key={m.l} className="mini-stat">
                      <div className="mini-stat-icon" style={{background:'var(--accent-glow)'}}>{m.i}</div>
                      <div className="mini-stat-info"><div className="mini-stat-label">{m.l}</div><div className="mini-stat-value">{m.v}</div></div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* ═══ TERMINAL ═══ */}
          {activeTab === 'terminal' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Global Terminal</h2>
                <select className="form-input" style={{ width:260 }} value={terminalServerId} onChange={e => setTerminalServerId(e.target.value)}>
                  <option value="">Select a server…</option>
                  {servers.map(s => <option key={s.id} value={s.id}>{s.hostname} ({s.ip_address})</option>)}
                </select>
              </div>
              <div className="card" style={{ background:'#0d1117', padding:0, overflow:'hidden', fontFamily:'var(--mono)', fontSize:13 }}>
                <div style={{ padding:16, minHeight:300, maxHeight:400, overflowY:'auto' }}>
                  {terminalOutput.length === 0 && <div style={{ color:'#6e7681' }}>Terminal output will appear here…</div>}
                  {terminalOutput.map((line, i) => (
                    <div key={i} style={{ color:line.startsWith('$')?'#58a6ff':line.startsWith('Error')?'#f85149':'#c9d1d9', marginBottom:4, whiteSpace:'pre-wrap' }}>{line}</div>
                  ))}
                </div>
                <div style={{ display:'flex', borderTop:'1px solid #21262d', alignItems:'center' }}>
                  <span style={{ padding:'10px 12px', color:'#58a6ff' }}>$</span>
                  <input className="form-input" style={{ background:'transparent', border:'none', color:'#c9d1d9', flex:1, outline:'none', fontSize:13 }}
                    value={terminalCmd} onChange={e => setTerminalCmd(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && runGlobalTerminal()}
                    placeholder="Enter command…" />
                  <button className="btn btn-primary btn-sm" style={{ margin:8 }} onClick={runGlobalTerminal}>Run</button>
                </div>
              </div>
            </div>
          )}

          {/* ═══ BILLING ═══ */}
          {activeTab === 'billing' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Razorpay Billing</h2>
                <div className="tab-pills">
                  {['overview','plans','transactions'].map(t => (
                    <button key={t} className={`tab-pill ${billingTab===t?'active':''}`} onClick={() => setBillingTab(t)}>{t.charAt(0).toUpperCase()+t.slice(1)}</button>
                  ))}
                </div>
              </div>

              {billingTab === 'overview' && (
                <>
                  <div className="stat-grid">
                    <div className="stat-card"><div className="stat-label">Monthly Revenue</div><div className="stat-value">₹42,500</div></div>
                    <div className="stat-card"><div className="stat-label">Active Subscriptions</div><div className="stat-value">{stats?.active_subscriptions || 12}</div></div>
                    <div className="stat-card"><div className="stat-label">Free Users</div><div className="stat-value">{stats?.free_users || users.length}</div></div>
                    <div className="stat-card"><div className="stat-label">Churn Rate</div><div className="stat-value" style={{fontSize:20}}>2.3%</div></div>
                  </div>
                  <div className="card">
                    <h4 style={{ fontSize:16, marginBottom:12 }}>Revenue Trend (12 months)</h4>
                    <div className="chart-container">
                      {[12000,18000,15000,22000,28000,25000,31000,27000,35000,30000,38000,42000].map((v,i) => (
                        <div key={i} className="chart-bar" style={{ height:`${(v/42000)*100}%`, background:'linear-gradient(135deg, #22c55e, #14b8a6)' }}>
                          <div className="chart-tooltip">₹{v.toLocaleString()}</div>
                          <div className="chart-bar-label">{chartLabels[i]}</div>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}

              {billingTab === 'plans' && (
                <div className="plan-grid">
                  {plans.map(p => (
                    <div key={p.name} className={`plan-card ${p.popular?'popular':''}`}>
                      <div className="plan-name">{p.name}</div>
                      <div className="plan-price">{p.price}{p.priceNum > 0 && <span>/mo</span>}</div>
                      <ul className="plan-features">
                        {p.features.map(f => <li key={f}>{f}</li>)}
                      </ul>
                      <button className="btn btn-primary btn-sm">{p.name==='Enterprise'?'Contact Sales':'Edit Plan'}</button>
                    </div>
                  ))}
                </div>
              )}

              {billingTab === 'transactions' && (
                <div className="card" style={{ padding:0, overflow:'hidden' }}>
                  <table className="table">
                    <thead><tr><th>Organization</th><th>Plan</th><th>Amount</th><th>Status</th><th>Date</th></tr></thead>
                    <tbody>
                      {[
                        {org:'Acme Corp',plan:'Business',amount:'₹4,999',status:'paid',date:'2026-03-15'},
                        {org:'Startup Inc',plan:'Pro',amount:'₹999',status:'paid',date:'2026-03-14'},
                        {org:'DevTeam',plan:'Pro',amount:'₹999',status:'paid',date:'2026-03-13'},
                        {org:'CloudOps',plan:'Business',amount:'₹4,999',status:'pending',date:'2026-03-12'},
                      ].map((t,i) => (
                        <tr key={i}>
                          <td style={{fontWeight:600}}>{t.org}</td>
                          <td><span className="badge badge-info">{t.plan}</span></td>
                          <td style={{fontWeight:600}}>{t.amount}</td>
                          <td><span className={`badge badge-${t.status==='paid'?'success':'warning'}`}>{t.status}</span></td>
                          <td style={{fontSize:13}}>{t.date}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* ═══ LOGS ═══ */}
          {activeTab === 'logs' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">System Logs</h2>
                <div className="log-live"><div className="log-live-dot" /> Live</div>
              </div>
              <div className="log-viewer">
                <div className="log-toolbar">
                  <select className="log-filter" value={logFilter} onChange={e => setLogFilter(e.target.value)}>
                    <option value="all">All Levels</option>
                    <option value="info">Info</option>
                    <option value="warn">Warning</option>
                    <option value="error">Error</option>
                    <option value="debug">Debug</option>
                  </select>
                  <input className="log-filter" style={{ flex:1, minWidth:200 }} placeholder="Search logs..." value={logSearch} onChange={e => setLogSearch(e.target.value)} />
                  <button className="btn btn-ghost btn-sm" style={{ color:'#c9d1d9', borderColor:'#30363d' }} onClick={() => setLogs([])}>Clear</button>
                </div>
                <div className="log-body">
                  {filteredLogs.map(l => (
                    <div key={l.id} className="log-line">
                      <span className="timestamp">{new Date(l.timestamp).toLocaleTimeString()}</span>
                      <span className={`level ${l.level}`}>{l.level}</span>
                      <span style={{ color:'#7ee787', fontWeight:500 }}>[{l.service}]</span>
                      <span className="msg">{l.message}</span>
                    </div>
                  ))}
                  {filteredLogs.length === 0 && <div style={{ color:'#484f58', textAlign:'center', padding:40 }}>No logs matching filter</div>}
                </div>
              </div>
            </div>
          )}

          {/* ═══ NOTIFICATIONS ═══ */}
          {activeTab === 'notifications' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Notifications</h2>
                <button className="btn btn-ghost btn-sm" onClick={() => { setNotifications(prev => prev.map(n => ({...n, read:true}))); showToast('✓ All marked as read'); }}>Mark All Read</button>
              </div>
              <div className="card">
                {notifications.map(n => (
                  <div key={n.id} className={`notif-item ${!n.read?'notif-unread':''}`}>
                    <div className="notif-icon" style={{ background:
                      n.type==='deploy'?'rgba(16,185,129,0.15)':
                      n.type==='alert'?'rgba(245,158,11,0.15)':
                      n.type==='billing'?'rgba(99,102,241,0.15)':
                      n.type==='user'?'rgba(56,189,248,0.15)':'rgba(128,128,128,0.15)' }}>
                      {n.type==='deploy'?'🚀':n.type==='alert'?'⚠️':n.type==='billing'?'💳':n.type==='user'?'👤':'🔧'}
                    </div>
                    <div className="notif-content">
                      <div className="notif-title">{n.title}</div>
                      <div className="notif-desc">{n.desc}</div>
                    </div>
                    <div className="notif-time">{n.time}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ═══ SETTINGS ═══ */}
          {activeTab === 'settings' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">Cluster Settings</h2></div>

              <div className="settings-group">
                <h4>⚙️ Platform Configuration</h4>
                {clusterSettings.map(s => (
                  <div key={s.key} className="settings-row">
                    <div className="settings-row-info">
                      <div className="settings-row-label">{s.key.replace(/_/g,' ').replace(/\b\w/g, c => c.toUpperCase())}</div>
                      <div className="settings-row-desc">{s.description}</div>
                    </div>
                    {s.value === 'true' || s.value === 'false' ? (
                      <button className={`toggle ${s.enabled?'on':''}`} onClick={() => updateClusterSetting(s.key, !s.enabled)} />
                    ) : (
                      <input className="form-input" style={{ width:100, textAlign:'center' }} value={s.value} onChange={e => updateClusterSetting(s.key, e.target.value)} />
                    )}
                  </div>
                ))}
              </div>

              <div className="settings-group">
                <h4>🔐 Security</h4>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">Cluster Secret</div><div className="settings-row-desc">Current authentication key for admin operations</div></div>
                  <code style={{ fontSize:12, color:'var(--text-muted)', background:'var(--bg-tertiary)', padding:'4px 10px', borderRadius:6 }}>{adminKey.slice(0,8)}••••••••</code>
                </div>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">JWT Expiry</div><div className="settings-row-desc">Access token lifetime</div></div>
                  <span style={{ fontSize:14, fontWeight:600 }}>15 minutes</span>
                </div>
              </div>

              <div className="settings-group" style={{ border:'1px solid rgba(239,68,68,0.3)' }}>
                <h4 style={{ color:'var(--danger)' }}>⚠️ Danger Zone</h4>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">Reset All Data</div><div className="settings-row-desc">Permanently delete all platform data. This cannot be undone.</div></div>
                  <button className="btn btn-danger btn-sm" onClick={() => alert('Feature disabled in production')}>Reset Platform</button>
                </div>
              </div>
            </div>
          )}

        </div>
      </main>

      {/* ═══ MODAL ═══ */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(null)}>
          <div className="modal scale-in" onClick={e => e.stopPropagation()}>
            {showModal === 'add-server' && (
              <>
                <h3>Provision New Server</h3>
                <p style={{ fontSize:14, color:'var(--text-muted)', marginBottom:16 }}>Run this command on a fresh Ubuntu 22.04+ VPS to auto-join the cluster:</p>
                <div style={{ background:'#0d1117', borderRadius:8, padding:16, fontFamily:'JetBrains Mono, monospace', fontSize:12, color:'#c9d1d9', overflowX:'auto', marginBottom:16 }}>
                  {modalData.script}
                </div>
                <div style={{ display:'flex', gap:8 }}>
                  <button className="btn btn-primary btn-sm" onClick={() => { navigator.clipboard.writeText(modalData.script); showToast('✓ Copied to clipboard'); }}>Copy Script</button>
                  <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(null)}>Close</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* ═══ TOAST ═══ */}
      {toast && <div className="toast">{toast}</div>}
    </div>
  );
}
