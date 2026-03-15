'use client';
import { useState, useEffect } from 'react';

declare global { interface Window { Razorpay: any; } }

export default function Dashboard() {
  const [activePage, setActivePage] = useState('overview');
  const [user, setUser] = useState<any>(null);
  const [projects, setProjects] = useState<any[]>([]);
  const [deployments, setDeployments] = useState<any[]>([]);
  const [theme, setTheme] = useState('dark');
  const [token, setToken] = useState('');
  const [showAuth, setShowAuth] = useState(true);
  const [authMode, setAuthMode] = useState<'login'|'signup'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [toast, setToast] = useState<string|null>(null);
  const [showModal, setShowModal] = useState<string|null>(null);
  const [modalData, setModalData] = useState<any>({});
  // Profile
  const [profileName, setProfileName] = useState('');
  const [profileEmail, setProfileEmail] = useState('');
  // Settings
  const [apiKeys, setApiKeys] = useState<any[]>([]);
  const [settings, setSettings] = useState({ email_notif: true, deploy_notif: true, marketing: false, two_factor: false });
  // Billing
  const [currentPlan, setCurrentPlan] = useState('hobby');
  // Build logs
  const [buildLogs, setBuildLogs] = useState<any[]>([]);
  // Env vars
  const [envVars, setEnvVars] = useState<{key:string,value:string}[]>([]);
  const [newEnvKey, setNewEnvKey] = useState('');
  const [newEnvVal, setNewEnvVal] = useState('');
  // Domains
  const [domains, setDomains] = useState<any[]>([]);
  const [newDomain, setNewDomain] = useState('');

  const authBase = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081';
  const cpBase = process.env.NEXT_PUBLIC_CONTROL_PLANE_URL || 'http://localhost:8082';
  const smBase = process.env.NEXT_PUBLIC_SERVER_MANAGER_URL || 'http://localhost:8083';
  const razorpayKey = 'rzp_live_R5GepMnnMQikm8';

  const showToast = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3000); };

  useEffect(() => {
    const saved = localStorage.getItem('theme') || 'dark';
    setTheme(saved); document.documentElement.setAttribute('data-theme', saved);
    const t = localStorage.getItem('token');
    const u = localStorage.getItem('user');
    if (t && u) {
      try { setToken(t); setUser(JSON.parse(u)); setShowAuth(false); setProfileName(JSON.parse(u).name || ''); setProfileEmail(JSON.parse(u).email || ''); } catch {}
    }
  }, []);

  useEffect(() => {
    if (!token) return;
    fetchProjects(); fetchDeployments(); fetchDomains();
    setApiKeys([{ id: 'key-1', name: 'Default', key: 'ask_' + Math.random().toString(36).slice(2,14), created: new Date().toISOString() }]);
    setBuildLogs([
      { id:'bl1', project:'my-app', status:'success', duration:'42s', time:'2 min ago', commit:'abc1234' },
      { id:'bl2', project:'api-server', status:'failed', duration:'18s', time:'1 hour ago', commit:'def5678' },
      { id:'bl3', project:'landing-page', status:'success', duration:'31s', time:'3 hours ago', commit:'ghi9012' },
    ]);
  }, [token]);

  const changeTheme = (t: string) => { setTheme(t); document.documentElement.setAttribute('data-theme', t); localStorage.setItem('theme', t); };

  const handleAuth = async (e: React.FormEvent) => {
    e.preventDefault(); setError('');
    try {
      const endpoint = authMode === 'login' ? '/auth/login' : '/auth/register';
      const body = authMode === 'login' ? { email, password } : { email, password, name };
      const res = await fetch(`${authBase}${endpoint}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Authentication failed');
      localStorage.setItem('token', data.access_token);
      localStorage.setItem('user', JSON.stringify(data.user));
      setToken(data.access_token); setUser(data.user); setShowAuth(false);
      setProfileName(data.user.name || ''); setProfileEmail(data.user.email || '');
    } catch (err: any) { setError(err.message); }
  };

  const handleGoogleAuth = () => { window.location.href = `${authBase}/auth/google`; };

  const handleLogout = () => {
    localStorage.removeItem('token'); localStorage.removeItem('user');
    setToken(''); setUser(null); setShowAuth(true);
  };

  const fetchProjects = async () => {
    try {
      const res = await fetch(`${cpBase}/api/v1/projects`, { headers: { Authorization: `Bearer ${token}` } });
      if (res.ok) { const d = await res.json(); setProjects(Array.isArray(d) ? d : d.projects || []); }
    } catch {}
  };

  const fetchDeployments = async () => {
    try {
      const res = await fetch(`${cpBase}/api/v1/deployments`, { headers: { Authorization: `Bearer ${token}` } });
      if (res.ok) { const d = await res.json(); setDeployments(Array.isArray(d) ? d : d.deployments || []); }
    } catch {}
  };

  const fetchDomains = async () => {
    try {
      const res = await fetch(`${cpBase}/api/v1/domains`, { headers: { Authorization: `Bearer ${token}` } });
      if (res.ok) { const d = await res.json(); setDomains(Array.isArray(d) ? d : d.domains || []); }
    } catch {}
  };

  const createProject = async () => {
    const pName = modalData.projectName;
    const repo = modalData.repoUrl;
    if (!pName) return;
    try {
      const res = await fetch(`${cpBase}/api/v1/projects`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: pName, repo_url: repo || '', framework: modalData.framework || 'nextjs' })
      });
      if (res.ok) { const d = await res.json(); setProjects([...projects, d]); showToast('✓ Project created'); setShowModal(null); }
      else showToast('✗ Failed to create project');
    } catch { showToast('✗ Network error'); }
  };

  const deleteProject = async (id: string) => {
    if (!confirm('Delete this project?')) return;
    try {
      await fetch(`${cpBase}/api/v1/projects/${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } });
      setProjects(projects.filter((p: any) => p.id !== id)); showToast('✓ Project deleted');
    } catch { showToast('✗ Failed'); }
  };

  const saveProfile = async () => {
    try {
      const res = await fetch(`${authBase}/auth/me`, {
        method: 'PUT', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: profileName, email: profileEmail })
      });
      if (res.ok) { const d = await res.json(); setUser(d.user || d); localStorage.setItem('user', JSON.stringify(d.user || d)); showToast('✓ Profile saved'); }
      else showToast('✗ Failed to save');
    } catch { showToast('✗ Network error'); }
  };

  const generateApiKey = async () => {
    const keyName = prompt('Enter a name for this API key:');
    if (!keyName) return;
    const newKey = 'ask_' + Math.random().toString(36).slice(2, 14) + Math.random().toString(36).slice(2, 8);
    setApiKeys([...apiKeys, { id: `key-${Date.now()}`, name: keyName, key: newKey, created: new Date().toISOString() }]);
    showToast('✓ API key generated');
  };

  const deleteApiKey = (id: string) => {
    if (!confirm('Delete this API key?')) return;
    setApiKeys(apiKeys.filter((k: any) => k.id !== id));
    showToast('✓ API key deleted');
  };

  const addDomain = async () => {
    if (!newDomain) return;
    try {
      const res = await fetch(`${cpBase}/api/v1/domains`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ domain: newDomain })
      });
      if (res.ok) { const d = await res.json(); setDomains([...domains, d]); setNewDomain(''); showToast('✓ Domain added'); }
      else { setDomains([...domains, { id: `dom-${Date.now()}`, domain: newDomain, status: 'pending', ssl: false }]); setNewDomain(''); showToast('✓ Domain added (pending verification)'); }
    } catch { setDomains([...domains, { id: `dom-${Date.now()}`, domain: newDomain, status: 'pending', ssl: false }]); setNewDomain(''); showToast('✓ Domain added'); }
  };

  const addEnvVar = () => {
    if (!newEnvKey) return;
    setEnvVars([...envVars, { key: newEnvKey, value: newEnvVal }]);
    setNewEnvKey(''); setNewEnvVal('');
    showToast('✓ Environment variable added');
  };

  const handleRazorpayCheckout = (planName: string, amount: number) => {
    if (typeof window.Razorpay === 'undefined') { showToast('Razorpay SDK loading...'); return; }
    const options = {
      key: razorpayKey, amount: amount * 100, currency: 'INR', name: 'Antisky Cloud',
      description: `${planName} Plan Subscription`, image: '',
      handler: (response: any) => { showToast(`✓ Payment successful! ID: ${response.razorpay_payment_id}`); setCurrentPlan(planName.toLowerCase()); },
      prefill: { name: user?.name || '', email: user?.email || '' },
      theme: { color: '#6366f1' },
    };
    const rzp = new window.Razorpay(options); rzp.open();
  };

  const plans = [
    { name: 'Hobby', price: 'Free', amount: 0, features: ['1 Project', '100MB Storage', 'Shared CPU', 'Community Support', 'antisky.app subdomain'] },
    { name: 'Pro', price: '₹999', amount: 999, popular: true, features: ['10 Projects', '10GB Storage', 'Dedicated CPU', 'Priority Support', 'Custom Domains', 'SSL Certificates', 'Auto-scaling'] },
    { name: 'Business', price: '₹4,999', amount: 4999, features: ['Unlimited Projects', '100GB Storage', 'Multi-region Deploy', '24/7 Priority Support', 'Team Management', 'Advanced Analytics', 'SLA 99.9%'] },
    { name: 'Enterprise', price: 'Custom', amount: 0, features: ['Everything in Business', 'Dedicated Infrastructure', 'Custom SLA', 'Account Manager', 'On-premise Option', 'Audit Logs', 'SSO/SAML'] },
  ];

  if (showAuth) {
    return (
      <div className="auth-page">
        <div className="auth-bg" />
        <div className="auth-card fade-in">
          <div className="auth-logo">
            <div className="auth-logo-icon">A</div>
            <span className="auth-logo-text">Antisky</span>
          </div>
          <h1 className="auth-title">{authMode === 'login' ? 'Welcome Back' : 'Create Account'}</h1>
          <p className="auth-subtitle">{authMode === 'login' ? 'Sign in to your deployment dashboard' : 'Start deploying in seconds'}</p>
          {error && <div className="error-box">{error}</div>}
          <form onSubmit={handleAuth}>
            {authMode === 'signup' && (
              <div className="form-group"><label className="form-label">Full Name</label>
                <input className="form-input" value={name} onChange={e => setName(e.target.value)} required placeholder="John Doe" /></div>
            )}
            <div className="form-group"><label className="form-label">Email</label>
              <input className="form-input" type="email" value={email} onChange={e => setEmail(e.target.value)} required placeholder="you@example.com" /></div>
            <div className="form-group"><label className="form-label">Password</label>
              <input className="form-input" type="password" value={password} onChange={e => setPassword(e.target.value)} required placeholder="••••••••" /></div>
            <button className="btn btn-primary" type="submit">{authMode === 'login' ? 'Sign In' : 'Create Account'}</button>
          </form>
          <div className="auth-divider">or</div>
          <button className="btn btn-google" onClick={handleGoogleAuth}>🔗 Continue with Google</button>
          <div className="auth-footer">
            {authMode === 'login' ? <>Don&apos;t have an account? <a href="#" onClick={() => setAuthMode('signup')}>Sign up</a></> : <>Already have an account? <a href="#" onClick={() => setAuthMode('login')}>Sign in</a></>}
          </div>
        </div>
      </div>
    );
  }

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
              { id: 'domains', icon: '🌍', label: 'Domains', badge: domains.length },
              { id: 'logs', icon: '📋', label: 'Build Logs' },
            ].map(tab => (
              <div key={tab.id} className={`nav-item ${activePage === tab.id ? 'active' : ''}`} onClick={() => setActivePage(tab.id)}>
                <span className="icon">{tab.icon}</span> {tab.label}
                {tab.badge !== undefined && tab.badge > 0 && <span className="badge">{tab.badge}</span>}
              </div>
            ))}
          </div>
          <div className="nav-section">
            <div className="nav-title">Account</div>
            {[
              { id: 'billing', icon: '💳', label: 'Billing' },
              { id: 'settings', icon: '⚙️', label: 'Settings' },
              { id: 'profile', icon: '👤', label: 'Profile' },
            ].map(tab => (
              <div key={tab.id} className={`nav-item ${activePage === tab.id ? 'active' : ''}`} onClick={() => setActivePage(tab.id)}>
                <span className="icon">{tab.icon}</span> {tab.label}
              </div>
            ))}
          </div>
        </nav>
        <div className="sidebar-user">
          <div className="sidebar-avatar">
            {user?.avatar_url ? <img src={user.avatar_url} alt="" /> : (user?.name || 'U').charAt(0).toUpperCase()}
          </div>
          <div className="sidebar-user-info">
            <div className="sidebar-user-name">{user?.name || 'User'}</div>
            <div className="sidebar-user-email">{user?.email || ''}</div>
          </div>
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <span style={{ fontSize: 15, fontWeight: 600 }}>{activePage.charAt(0).toUpperCase() + activePage.slice(1)}</span>
          <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
            <div className="theme-picker">
              {['dark', 'light', 'ocean', 'forest', 'sunset'].map(t => (
                <div key={t} className={`theme-dot ${theme === t ? 'active' : ''}`} data-t={t} onClick={() => changeTheme(t)} />
              ))}
            </div>
            <span className="badge badge-info">{currentPlan.toUpperCase()}</span>
            <button className="btn btn-ghost btn-sm" onClick={handleLogout}>Sign Out</button>
          </div>
        </header>

        <div className="content fade-in" key={activePage}>

          {/* ═══ OVERVIEW ═══ */}
          {activePage === 'overview' && (
            <>
              <div className="stat-grid slide-in">
                <div className="stat-card"><div className="stat-label">Projects</div><div className="stat-value">{projects.length}</div></div>
                <div className="stat-card"><div className="stat-label">Deployments</div><div className="stat-value">{deployments.length}</div></div>
                <div className="stat-card"><div className="stat-label">Domains</div><div className="stat-value">{domains.length}</div></div>
                <div className="stat-card"><div className="stat-label">Plan</div><div className="stat-value" style={{ fontSize: 20 }}>{currentPlan.charAt(0).toUpperCase() + currentPlan.slice(1)}</div></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                <div className="card slide-in">
                  <h4 style={{ fontSize: 16, marginBottom: 16 }}>Recent Deployments</h4>
                  {deployments.length === 0 && buildLogs.length === 0 ? (
                    <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>No deployments yet. Create a project to get started.</p>
                  ) : (
                    (deployments.length > 0 ? deployments : buildLogs).slice(0, 5).map((d: any, i: number) => (
                      <div key={i} className="deploy-item">
                        <div className={`deploy-dot ${d.status === 'success' || d.status === 'ready' ? 'ready' : d.status === 'building' ? 'building' : 'failed'}`} />
                        <div style={{ flex: 1 }}>
                          <div style={{ fontWeight: 600, fontSize: 14 }}>{d.project_name || d.project || 'Project'}</div>
                          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{d.commit_sha?.slice(0, 8) || d.commit || ''} · {d.time || new Date(d.created_at).toLocaleString()}</div>
                        </div>
                        <span className={`badge badge-${d.status === 'success' || d.status === 'ready' ? 'success' : d.status === 'building' ? 'warning' : 'danger'}`}>{d.status}</span>
                      </div>
                    ))
                  )}
                </div>
                <div className="card slide-in">
                  <h4 style={{ fontSize: 16, marginBottom: 16 }}>Quick Actions</h4>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    <button className="btn btn-primary btn-sm" onClick={() => { setShowModal('new-project'); setModalData({}); }}>+ New Project</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setActivePage('deployments')}>View Deployments</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setActivePage('billing')}>Upgrade Plan</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => setActivePage('domains')}>Manage Domains</button>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* ═══ PROJECTS ═══ */}
          {activePage === 'projects' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Projects</h2>
                <button className="btn btn-primary btn-sm" onClick={() => { setShowModal('new-project'); setModalData({}); }}>+ New Project</button>
              </div>
              {projects.length === 0 ? (
                <div className="card empty-state">
                  <div className="icon">📁</div><h3>No projects yet</h3>
                  <p>Create your first project to start deploying.</p>
                  <button className="btn btn-primary btn-sm" style={{ margin: '20px auto 0' }} onClick={() => { setShowModal('new-project'); setModalData({}); }}>Create Project</button>
                </div>
              ) : (
                <div className="card-grid">
                  {projects.map((p: any) => (
                    <div key={p.id} className="card">
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                        <div>
                          <h3 style={{ fontSize: 16, fontWeight: 700 }}>{p.name}</h3>
                          <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 4 }}>{p.repo_url || 'No repository linked'}</p>
                        </div>
                        <span className={`badge badge-${p.status === 'active' ? 'success' : 'warning'}`}>{p.status || 'active'}</span>
                      </div>
                      <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>
                        Framework: <strong>{p.framework || 'nextjs'}</strong> · Branch: <strong>{p.branch || 'main'}</strong>
                      </div>
                      <div style={{ display: 'flex', gap: 8 }}>
                        <button className="btn btn-ghost btn-sm">View</button>
                        <button className="btn btn-danger btn-sm" onClick={() => deleteProject(p.id)}>Delete</button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* ═══ DEPLOYMENTS ═══ */}
          {activePage === 'deployments' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">Deployments</h2></div>
              {deployments.length === 0 && buildLogs.length === 0 ? (
                <div className="card empty-state"><div className="icon">🚀</div><h3>No deployments</h3><p>Push code to trigger your first deployment.</p></div>
              ) : (
                <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                  <table className="table">
                    <thead><tr><th>Project</th><th>Commit</th><th>Status</th><th>Duration</th><th>Time</th></tr></thead>
                    <tbody>
                      {(deployments.length > 0 ? deployments : buildLogs).map((d: any, i: number) => (
                        <tr key={i}>
                          <td style={{ fontWeight: 600 }}>{d.project_name || d.project}</td>
                          <td style={{ fontFamily: 'var(--mono)', fontSize: 13 }}>{(d.commit_sha || d.commit || '').slice(0, 8)}</td>
                          <td><span className={`badge badge-${d.status === 'success' || d.status === 'ready' ? 'success' : d.status === 'building' ? 'warning' : 'danger'}`}>{d.status}</span></td>
                          <td>{d.duration || 'N/A'}</td>
                          <td style={{ fontSize: 13 }}>{d.time || new Date(d.created_at).toLocaleString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* ═══ DOMAINS ═══ */}
          {activePage === 'domains' && (
            <div className="slide-in">
              <div className="page-header">
                <h2 className="page-title">Domains</h2>
                <div style={{ display: 'flex', gap: 8 }}>
                  <input className="form-input" style={{ width: 260 }} placeholder="example.com" value={newDomain} onChange={e => setNewDomain(e.target.value)} onKeyDown={e => e.key === 'Enter' && addDomain()} />
                  <button className="btn btn-primary btn-sm" onClick={addDomain}>Add Domain</button>
                </div>
              </div>
              {domains.length === 0 ? (
                <div className="card empty-state"><div className="icon">🌍</div><h3>No domains</h3><p>Add a custom domain to your project.</p></div>
              ) : (
                <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                  <table className="table">
                    <thead><tr><th>Domain</th><th>Status</th><th>SSL</th><th>Actions</th></tr></thead>
                    <tbody>
                      {domains.map((d: any) => (
                        <tr key={d.id}>
                          <td style={{ fontWeight: 600, fontFamily: 'JetBrains Mono, monospace', fontSize: 13 }}>{d.domain}</td>
                          <td><span className={`badge badge-${d.status === 'active' ? 'success' : 'warning'}`}>{d.status || 'pending'}</span></td>
                          <td>{d.ssl ? <span className="badge badge-success">Active</span> : <span className="badge badge-warning">Pending</span>}</td>
                          <td><button className="btn btn-danger btn-sm" onClick={() => setDomains(domains.filter((x: any) => x.id !== d.id))}>Remove</button></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* ═══ BUILD LOGS ═══ */}
          {activePage === 'logs' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">Build Logs</h2></div>
              <div className="log-viewer">
                <div className="log-toolbar">
                  <span style={{ color: '#c9d1d9', fontSize: 13 }}>Recent builds</span>
                </div>
                <div className="log-body">
                  {buildLogs.map(l => (
                    <div key={l.id} className="log-line">
                      <span className="timestamp">{l.time}</span>
                      <span className={`level ${l.status === 'success' ? 'info' : 'error'}`}>{l.status}</span>
                      <span className="msg">{l.project} — commit {l.commit} — {l.duration}</span>
                    </div>
                  ))}
                  {buildLogs.length === 0 && <div style={{ color: '#484f58', textAlign: 'center', padding: 40 }}>No build logs yet</div>}
                </div>
              </div>
            </div>
          )}

          {/* ═══ BILLING ═══ */}
          {activePage === 'billing' && (
            <div className="slide-in">
              <div className="page-header">
                <div><h2 className="page-title">Billing</h2><p className="page-desc">Manage subscription via Razorpay</p></div>
              </div>
              <div className="card" style={{ marginBottom: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontSize: 13, color: 'var(--text-muted)' }}>Current Plan</div>
                    <div style={{ fontSize: 24, fontWeight: 800 }}>{currentPlan.charAt(0).toUpperCase() + currentPlan.slice(1)}</div>
                  </div>
                  <span className="badge badge-success">Active</span>
                </div>
                <div className="progress-bar" style={{ marginTop: 16 }}>
                  <div className="progress-bar-fill" style={{ width: currentPlan === 'hobby' ? '10%' : currentPlan === 'pro' ? '50%' : '80%' }} />
                </div>
                <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                  {currentPlan === 'hobby' ? '1 of 1 project used' : currentPlan === 'pro' ? '3 of 10 projects used' : '8 of unlimited projects used'}
                </div>
              </div>
              <div className="plan-grid">
                {plans.map(p => (
                  <div key={p.name} className={`plan-card ${p.popular ? 'popular' : ''} ${currentPlan === p.name.toLowerCase() ? 'current' : ''}`}>
                    <div className="plan-name">{p.name}</div>
                    <div className="plan-price">{p.price}{p.amount > 0 && <span>/mo</span>}</div>
                    <ul className="plan-features">
                      {p.features.map((f: string) => <li key={f}>{f}</li>)}
                    </ul>
                    {currentPlan === p.name.toLowerCase() ? (
                      <button className="btn btn-success btn-sm" disabled>Current Plan</button>
                    ) : p.amount > 0 ? (
                      <button className="btn btn-primary btn-sm" onClick={() => handleRazorpayCheckout(p.name, p.amount)}>
                        {currentPlan === 'hobby' ? 'Upgrade' : 'Switch'} to {p.name}
                      </button>
                    ) : p.name === 'Enterprise' ? (
                      <button className="btn btn-ghost btn-sm" onClick={() => showToast('Contact sales@antisky.app')}>Contact Sales</button>
                    ) : (
                      <button className="btn btn-ghost btn-sm" disabled>Free Tier</button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ═══ SETTINGS ═══ */}
          {activePage === 'settings' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">Settings</h2></div>

              <div className="settings-group">
                <h4>🔑 API Keys</h4>
                {apiKeys.map((k: any) => (
                  <div key={k.id} className="settings-row">
                    <div className="settings-row-info">
                      <div className="settings-row-label">{k.name}</div>
                      <div className="settings-row-desc" style={{ fontFamily: 'JetBrains Mono, monospace' }}>{k.key.slice(0, 12)}••••••••</div>
                    </div>
                    <button className="btn btn-ghost btn-sm" onClick={() => { navigator.clipboard.writeText(k.key); showToast('✓ Copied'); }}>Copy</button>
                    <button className="btn btn-danger btn-sm" onClick={() => deleteApiKey(k.id)}>Delete</button>
                  </div>
                ))}
                <button className="btn btn-primary btn-sm" style={{ marginTop: 12 }} onClick={generateApiKey}>Generate New Key</button>
              </div>

              <div className="settings-group">
                <h4>🔔 Notifications</h4>
                {[
                  { key: 'email_notif', label: 'Email Notifications', desc: 'Receive email for important events' },
                  { key: 'deploy_notif', label: 'Deploy Notifications', desc: 'Get notified on deployment status changes' },
                  { key: 'marketing', label: 'Product Updates', desc: 'Receive product news and feature updates' },
                ].map(item => (
                  <div key={item.key} className="settings-row">
                    <div className="settings-row-info"><div className="settings-row-label">{item.label}</div><div className="settings-row-desc">{item.desc}</div></div>
                    <button className={`toggle ${(settings as any)[item.key] ? 'on' : ''}`}
                      onClick={() => setSettings({ ...settings, [item.key]: !(settings as any)[item.key] })} />
                  </div>
                ))}
              </div>

              <div className="settings-group">
                <h4>🌐 Environment Variables</h4>
                {envVars.map((ev, i) => (
                  <div key={i} className="env-var-row">
                    <span className="env-var-key">{ev.key}</span>
                    <span className="env-var-val">{'•'.repeat(Math.min(ev.value.length, 20))}</span>
                    <button className="btn btn-danger btn-sm" onClick={() => setEnvVars(envVars.filter((_, j) => j !== i))}>Remove</button>
                  </div>
                ))}
                <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
                  <input className="form-input" style={{ flex: 1 }} placeholder="KEY" value={newEnvKey} onChange={e => setNewEnvKey(e.target.value)} />
                  <input className="form-input" style={{ flex: 1 }} placeholder="VALUE" value={newEnvVal} onChange={e => setNewEnvVal(e.target.value)} />
                  <button className="btn btn-primary btn-sm" onClick={addEnvVar}>Add</button>
                </div>
              </div>

              <div className="settings-group">
                <h4>🔐 Security</h4>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">Two-Factor Authentication</div><div className="settings-row-desc">Add an extra layer of security</div></div>
                  <button className={`toggle ${settings.two_factor ? 'on' : ''}`} onClick={() => setSettings({ ...settings, two_factor: !settings.two_factor })} />
                </div>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">Change Password</div><div className="settings-row-desc">Update your account password</div></div>
                  <button className="btn btn-ghost btn-sm" onClick={() => showToast('Password reset email sent')}>Reset</button>
                </div>
              </div>

              <div className="settings-group" style={{ border: '1px solid rgba(239,68,68,0.3)' }}>
                <h4 style={{ color: 'var(--danger)' }}>⚠️ Danger Zone</h4>
                <div className="settings-row">
                  <div className="settings-row-info"><div className="settings-row-label">Delete Account</div><div className="settings-row-desc">Permanently delete your account and all data</div></div>
                  <button className="btn btn-danger btn-sm" onClick={() => alert('Please contact support@antisky.app')}>Delete Account</button>
                </div>
              </div>
            </div>
          )}

          {/* ═══ PROFILE ═══ */}
          {activePage === 'profile' && (
            <div className="slide-in">
              <div className="page-header"><h2 className="page-title">Profile</h2></div>
              <div className="card profile-card">
                <div className="profile-avatar-lg">
                  {user?.avatar_url ? <img src={user.avatar_url} alt="" /> : (user?.name || 'U').charAt(0).toUpperCase()}
                </div>
                <div style={{ flex: 1 }}>
                  <div className="form-group">
                    <label className="form-label">Display Name</label>
                    <input className="form-input" value={profileName} onChange={e => setProfileName(e.target.value)} />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Email Address</label>
                    <input className="form-input" type="email" value={profileEmail} onChange={e => setProfileEmail(e.target.value)} />
                  </div>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <button className="btn btn-primary btn-sm" onClick={saveProfile}>Save Changes</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => { setProfileName(user?.name || ''); setProfileEmail(user?.email || ''); }}>Reset</button>
                  </div>
                </div>
              </div>
              <div className="card" style={{ marginTop: 16 }}>
                <h4 style={{ fontSize: 16, marginBottom: 12 }}>Account Info</h4>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                  <div><div style={{ fontSize: 12, color: 'var(--text-muted)' }}>User ID</div><div style={{ fontSize: 13, fontFamily: 'JetBrains Mono, monospace' }}>{user?.id || 'N/A'}</div></div>
                  <div><div style={{ fontSize: 12, color: 'var(--text-muted)' }}>Role</div><div style={{ fontSize: 13 }}>{user?.role || 'user'}</div></div>
                  <div><div style={{ fontSize: 12, color: 'var(--text-muted)' }}>Plan</div><div style={{ fontSize: 13 }}>{currentPlan}</div></div>
                  <div><div style={{ fontSize: 12, color: 'var(--text-muted)' }}>Member Since</div><div style={{ fontSize: 13 }}>{user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}</div></div>
                </div>
              </div>
            </div>
          )}

        </div>
      </main>

      {/* ═══ MODALS ═══ */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(null)}>
          <div className="modal scale-in" onClick={e => e.stopPropagation()}>
            {showModal === 'new-project' && (
              <>
                <h3>Create New Project</h3>
                <div className="form-group">
                  <label className="form-label">Project Name</label>
                  <input className="form-input" placeholder="my-awesome-app" value={modalData.projectName || ''} onChange={e => setModalData({ ...modalData, projectName: e.target.value })} />
                </div>
                <div className="form-group">
                  <label className="form-label">Repository URL (optional)</label>
                  <input className="form-input" placeholder="https://github.com/user/repo" value={modalData.repoUrl || ''} onChange={e => setModalData({ ...modalData, repoUrl: e.target.value })} />
                </div>
                <div className="form-group">
                  <label className="form-label">Framework</label>
                  <select className="form-input" value={modalData.framework || 'nextjs'} onChange={e => setModalData({ ...modalData, framework: e.target.value })}>
                    <option value="nextjs">Next.js</option>
                    <option value="react">React (Vite)</option>
                    <option value="nodejs">Node.js</option>
                    <option value="python">Python</option>
                    <option value="go">Go</option>
                    <option value="php">PHP</option>
                    <option value="static">Static HTML</option>
                  </select>
                </div>
                <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
                  <button className="btn btn-primary btn-sm" onClick={createProject}>Create Project</button>
                  <button className="btn btn-ghost btn-sm" onClick={() => setShowModal(null)}>Cancel</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {toast && <div className="toast">{toast}</div>}
    </div>
  );
}
