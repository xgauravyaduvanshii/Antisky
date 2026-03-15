'use client';
import { useState } from 'react';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Call password reset API
    setSent(true);
  };

  return (
    <div className="auth-page">
      <div className="auth-bg" />
      <div className="auth-card fade-in">
        <div className="auth-logo">
          <div className="auth-logo-icon">A</div>
          <span className="auth-logo-text">Antisky</span>
        </div>

        {!sent ? (
          <>
            <h1 className="auth-title">Reset your password</h1>
            <p className="auth-subtitle">Enter your email and we&apos;ll send a reset link</p>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label className="form-label">Email</label>
                <input className="form-input" type="email" placeholder="you@example.com" value={email} onChange={e => setEmail(e.target.value)} required />
              </div>
              <button className="btn btn-primary" type="submit">Send Reset Link</button>
            </form>
          </>
        ) : (
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 48, marginBottom: 16 }}>📧</div>
            <h1 className="auth-title">Check your email</h1>
            <p className="auth-subtitle">We sent a password reset link to <strong style={{ color: 'var(--accent)' }}>{email}</strong></p>
          </div>
        )}

        <div className="auth-footer">
          <a href="/login">← Back to sign in</a>
        </div>
      </div>
    </div>
  );
}
