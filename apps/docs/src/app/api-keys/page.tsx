import { Shield, Key, Lock, AlertTriangle, Terminal, Copy } from 'lucide-react'

export default function ApiKeys() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-rose-500/10 text-rose-400 text-sm font-medium mb-6 border border-rose-500/20">
        <Shield size={14} />
        Security & Access
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        Personal API Keys
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        API Keys allow you to programmatically trigger builds, retrieve server metrics, or run headless deployments directly against the Antisky Control Plane without needing browser authentication.
      </p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Generating a Key</h2>
          <p className="text-gray-400 mb-6">
            Head to the Antisky User Dashboard &rarr; <strong>Settings</strong> &rarr; <strong>API Keys</strong>.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-5 rounded-2xl bg-[#0a0a0f] border border-white/5 flex flex-col items-center justify-center text-center">
              <div className="w-12 h-12 rounded-full bg-rose-500/10 flex items-center justify-center mb-3">
                <Key size={24} className="text-rose-400" />
              </div>
              <h3 className="text-white font-medium mb-1">Create Key</h3>
              <p className="text-sm text-gray-500">Assign a descriptive name (e.g., GitHub Actions).</p>
            </div>
            
            <div className="p-5 rounded-2xl bg-[#0a0a0f] border border-white/5 flex flex-col items-center justify-center text-center">
              <div className="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center mb-3">
                <AlertTriangle size={24} className="text-amber-400" />
              </div>
              <h3 className="text-white font-medium mb-1">Store Securely</h3>
              <p className="text-sm text-gray-500">Keys are hashed immediately. You can only view it once.</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <Terminal className="text-rose-400" /> Authenticating Requests
          </h2>
          <p className="text-gray-400 mb-6">
            Pass your generated key inside the HTTP <code>Authorization</code> header using the standard Bearer token schema.
          </p>

          <div className="relative group mt-6">
            <div className="absolute inset-0 bg-gradient-to-r from-rose-500/10 to-orange-500/10 rounded-xl blur " />
            <div className="relative bg-[#0d1117] border border-white/10 rounded-xl p-4 overflow-hidden">
              <div className="flex items-center justify-between mb-3 text-xs text-gray-500 font-medium">
                <span className="flex items-center gap-2"><Terminal size={14} /> cURL</span>
                <button className="hover:text-white transition-colors"><Copy size={14} /></button>
              </div>
              <pre className="text-sm font-mono text-gray-300 overflow-x-auto whitespace-pre-wrap">
<span className="text-rose-400">curl</span> -X GET https://api.antisky.app/v1/projects \
  -H "Authorization: Bearer <span className="text-amber-300">asky_live_9f8d7c6b5...</span>"
              </pre>
            </div>
          </div>
        </section>

        <div className="p-6 rounded-2xl bg-white/5 border border-white/10 flex gap-4">
          <Lock size={24} className="text-gray-400 shrink-0" />
          <div>
            <h3 className="text-lg font-bold text-white mb-2">Revoking Access</h3>
            <p className="text-gray-400 text-sm leading-relaxed">
              If an API key is compromised, you can instantly revoke its access from the Dashboard. Outstanding requests signed with a revoked key will immediately drop with a <code>401 Unauthorized</code> response from the edge firewall.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
