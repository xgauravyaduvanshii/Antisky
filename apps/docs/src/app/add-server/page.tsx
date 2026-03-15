import { Server, Terminal, CheckCircle2, Copy } from 'lucide-react'

export default function AddServer() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 text-sm font-medium mb-6 border border-emerald-500/20">
        <Server size={14} />
        Core Infrastructure
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        Adding Bare-Metal Servers
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        Antisky works by clustering your existing Linux machines into a globally distributed edge network.
        You can attach any cloud provider (AWS EC2, DigitalOcean, Hetzner) or on-premise hardware.
      </p>

      <div className="space-y-12">
        <Step 
          number={1} 
          title="Prepare an Ubuntu 22.04+ Server"
          content={
            <div className="text-gray-400 space-y-4">
              <p>Spin up a fresh VPS or EC2 instance. We heavily recommend Ubuntu 22.04 LTS or newer.</p>
              <ul className="space-y-2 list-none p-0">
                <li className="flex items-center gap-2"><CheckCircle2 size={16} className="text-emerald-500" /> Minimum 1GB RAM</li>
                <li className="flex items-center gap-2"><CheckCircle2 size={16} className="text-emerald-500" /> Root or sudo privileges</li>
                <li className="flex items-center gap-2"><CheckCircle2 size={16} className="text-emerald-500" /> Port 80 and 443 open for ingress</li>
              </ul>
            </div>
          }
        />

        <Step 
          number={2} 
          title="Retrieve the Cluster Secret"
          content={
            <div className="text-gray-400">
              <p className="mb-4">Navigate to your Antisky Admin Dashboard &rarr; <strong>Settings</strong>.</p>
              <p>Under the "Security" section, copy your <code>Cluster Secret Key</code>. This authenticates the node with the Control Plane.</p>
            </div>
          }
        />

        <Step 
          number={3} 
          title="Run the Provisioning Script"
          content={
            <div className="space-y-4">
              <p className="text-gray-400">SSH into your newly created server and run the global install shell script. Replace the secret flag with your actual key.</p>
              
              <div className="relative group">
                <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/20 to-purple-500/20 rounded-xl blur " />
                <div className="relative bg-[#0d1117] border border-white/10 rounded-xl p-4 overflow-hidden">
                  <div className="flex items-center justify-between mb-3 text-xs text-gray-500 font-medium">
                    <span className="flex items-center gap-2"><Terminal size={14} /> bash</span>
                    <button className="hover:text-white transition-colors" title="Copy code"><Copy size={14} /></button>
                  </div>
                  <code className="text-sm font-mono text-emerald-400 block whitespace-pre-wrap">
                    curl -sSL https://get.antisky.app | bash -s -- --secret antisky-cluster-secret-2026
                  </code>
                </div>
              </div>

              <div className="p-4 rounded-lg bg-blue-500/10 border border-blue-500/20 mt-6">
                <h4 className="text-blue-400 font-medium mb-1">What does this script do?</h4>
                <p className="text-sm text-blue-200/70">
                  It installs Docker, configures the Docker daemon for standard JSON logging, downloads the Antisky Node Agent binary, registers the machine IP with your cluster, and sets up a systemd service to ensure the agent runs on boot.
                </p>
              </div>
            </div>
          }
        />
      </div>
    </div>
  )
}

function Step({ number, title, content }: { number: number, title: string, content: React.ReactNode }) {
  return (
    <div className="relative pl-12">
      <div className="absolute left-0 top-0.5 w-8 h-8 rounded-full bg-white/5 border border-white/10 flex items-center justify-center font-bold text-white shadow-lg">
        {number}
      </div>
      <div className="absolute left-4 top-10 bottom-[-3rem] w-px bg-white/5 last:bg-transparent" />
      <h3 className="text-xl font-bold text-white mb-3">{title}</h3>
      {content}
    </div>
  )
}
