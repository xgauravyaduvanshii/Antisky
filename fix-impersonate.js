const fs = require('fs');
const file = '/home/ubuntu/antisky/apps/admin/src/app/page.tsx';
let data = fs.readFileSync(file, 'utf8');

const target1 = `  if (authChecking) {
    return <div style={{ display: 'grid', placeItems: 'center', height: '100vh', background: 'var(--bg)', color: 'var(--text)' }}>
      <div className="loader"></div>
    </div>;
  }`;

const replacement1 = `  async function impersonateUser(targetUserId: string) {
    try {
      const res = await fetch(\`\${apiBase}/impersonate\`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${token}\`
        },
        body: JSON.stringify({ target_user_id: targetUserId })
      });
      const data = await res.json();
      if (res.ok && data.access_token) {
        // Save user token to localStorage and redirect to user dashboard
        localStorage.setItem('auth_token', data.access_token);
        // Do not overwrite cluster_secret in case admin comes back
        window.location.href = 'http://localhost:3000';
      } else {
        alert(data.error?.message || 'Failed to impersonate user');
      }
    } catch (err) {
      console.error(err);
      alert('Error connecting to auth service');
    }
  }

  if (authChecking) {
    return <div style={{ display: 'grid', placeItems: 'center', height: '100vh', background: 'var(--bg)', color: 'var(--text)' }}>
      <div className="loader"></div>
    </div>;
  }`;

data = data.replace(target1, replacement1);

fs.writeFileSync(file, data);
console.log("Replaced successfully!");
