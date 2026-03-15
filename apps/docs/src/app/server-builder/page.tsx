import { Server, Terminal, Settings, ArrowRight, ShieldCheck, Box, Network } from 'lucide-react'

export default function ServerBuilder() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/10 text-cyan-400 text-sm font-medium mb-6 border border-cyan-500/20">
        <Network size={14} />
        Orchestration Internals
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        Builder Node Architecture
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        Antisky decouples the primary Control Plane (your Admin Dashboard) from the physical machines executing the code. 
        When you provision a new server, it transforms into an autonomous <strong>Builder Node</strong>.
      </p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-2">
            <Terminal className="text-cyan-400" /> Phase 1: The Upload
          </h2>
          <div className="p-6 rounded-2xl bg-white/5 border border-white/10 relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-32 h-32 bg-cyan-500/10 rounded-full blur-3xl -mr-10 -mt-10 transition-transform group-hover:scale-150 duration-700" />
            <p className="text-gray-300 relative z-10">
              When an administrator executes the <code>curl -sSL https://get.antisky.app | bash</code> installation string, a payload is pulled directly from the root <code>builder/</code> directory of the cluster instance.
            </p>
            <div className="mt-4 flex items-center gap-4 text-sm font-mono text-gray-500 relative z-10">
              <span className="flex items-center gap-1 text-cyan-400"><Box size={14}/> install.sh</span>
              <ArrowRight size={14} />
              <span className="flex items-center gap-1 text-cyan-400"><Box size={14}/> start-server.sh</span>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-2">
            <Settings className="text-cyan-400" /> Phase 2: Dependency Initialization
          </h2>
          <div className="space-y-4">
            <p className="text-gray-400">
              The <code>install.sh</code> script executes with root privileges to automatically prepare the Ubuntu 22.04+ headless server for container orchestration.
            </p>
            <ul className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
              <li className="flex items-center gap-3 p-3 rounded-lg bg-[#0d1117] border border-white/5 text-gray-300 text-sm">
                <span className="w-2 h-2 rounded-full bg-blue-500" /> Docker Engine & Compose
              </li>
              <li className="flex items-center gap-3 p-3 rounded-lg bg-[#0d1117] border border-white/5 text-gray-300 text-sm">
                <span className="w-2 h-2 rounded-full bg-emerald-500" /> Node.js 20 LTS
              </li>
              <li className="flex items-center gap-3 p-3 rounded-lg bg-[#0d1117] border border-white/5 text-gray-300 text-sm">
                <span className="w-2 h-2 rounded-full bg-cyan-500" /> Golang 1.22
              </li>
              <li className="flex items-center gap-3 p-3 rounded-lg bg-[#0d1117] border border-white/5 text-gray-300 text-sm">
                <span className="w-2 h-2 rounded-full bg-amber-500" /> UFW Firewall (Ports 80, 443, 8090)
              </li>
            </ul>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-2">
            <ShieldCheck className="text-cyan-400" /> Phase 3: Registration & Handshake
          </h2>
          <div className="p-6 rounded-2xl bg-[#0a0a0f] border border-white/5">
            <p className="text-gray-300 leading-relaxed mb-6">
              Post-installation, <code>register-server.sh</code> generates a unique RSA-level <strong>Server Identity Key</strong> and <strong>Server ID</strong>.
              It then dials the primary Control Plane using your injected <code>--secret</code> key.
            </p>
            
            <div className="relative p-5 rounded-xl bg-[#0d1117] border border-white/10 font-mono text-xs overflow-x-auto text-gray-400">
              <div className="text-emerald-400 mb-2"># payload.json sent to Admin API</div>
              {`{
  "server_id": "8f7d9a2b4...",
  "hostname": "ip-172-31-...",
  "ip_address": "8.8.8.8",
  "key": "[REDACTED]",
  "resources": { "cpu_cores": 4, "ram_gb": 8 }
}`}
            </div>
            <p className="text-sm text-gray-500 mt-4 italic">
              Once verified, the server appears "Online" inside the Global Admin Dashboard.
            </p>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-2">
            <Server className="text-cyan-400" /> Phase 4: Deploying
          </h2>
          <div className="p-6 rounded-2xl bg-cyan-500/5 border border-cyan-500/20">
            <p className="text-gray-300 leading-relaxed mb-4">
              When a user triggers a Github deployment, the Control Plane selects the most optimal Builder Node (lowest CPU/Memory utilization).
            </p>
            <p className="text-gray-300 leading-relaxed">
              The Control Plane sends a webhook to port <code>8090</code> of the Builder. The builder then fetches the source code, matches the framework, boots an isolated Docker build container, pushes the finalized image to the private registry, and routes external traffic to it using Traefik.
            </p>
          </div>
        </section>

      </div>
    </div>
  )
}
