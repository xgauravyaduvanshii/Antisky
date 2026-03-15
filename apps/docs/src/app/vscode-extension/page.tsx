import { Code, Settings, Terminal, Copy } from 'lucide-react'

export default function VSCodeExtension() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 text-blue-400 text-sm font-medium mb-6 border border-blue-500/20">
        <Code size={14} />
        Developer Experience
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        VSCode Extension
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        Integrate the Antisky platform directly into your editor. Track live deployments, tail production logs, and stream container metrics without ever leaving Visual Studio Code.
      </p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Installation</h2>
          <p className="text-gray-400 mb-4">
            You can install the official Antisky VSCode extension directly from the Visual Studio Marketplace.
          </p>
          <div className="bg-[#0d1117] border border-white/10 rounded-xl p-4 flex items-center justify-between">
            <code className="text-sm font-mono text-gray-300">
              ext install antisky.antisky-vscode
            </code>
            <button className="text-gray-500 hover:text-white transition-colors"><Copy size={16} /></button>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <Settings className="text-blue-400" /> Configurations
          </h2>
          <p className="text-gray-400 mb-4">
            To link the extension to your platform cluster, you must generate a Personal API Key from the Antisky User Dashboard, then update your <code>settings.json</code> workflow.
          </p>

          <div className="relative group mt-6">
            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-indigo-500/10 rounded-xl blur " />
            <div className="relative bg-[#0d1117] border border-white/10 rounded-xl p-4 overflow-hidden">
              <div className="flex items-center justify-between mb-3 text-xs text-gray-500 font-medium">
                <span className="flex items-center gap-2"><Terminal size={14} /> settings.json</span>
                <button className="hover:text-white transition-colors"><Copy size={14} /></button>
              </div>
              <pre className="text-sm font-mono text-blue-300 overflow-x-auto">
{`{
  "antisky.clusterUrl": "https://api.antisky.app",
  "antisky.apiKey": "asky_live_9f8d7c6b5a4...",
  "antisky.defaultProject": "my-nextjs-app",
  "antisky.telemetry": true
}`}
              </pre>
            </div>
          </div>
        </section>

        <section className="p-6 rounded-2xl bg-white/5 border border-white/10">
          <h3 className="text-lg font-bold text-white mb-2">Available Commands</h3>
          <ul className="space-y-3 text-gray-400">
            <li><kbd className="bg-white/10 px-2 py-1 rounded text-xs text-gray-200 uppercase">CMD + SHIFT + P</kbd> <code>Antisky: Deploy Project</code></li>
            <li><kbd className="bg-white/10 px-2 py-1 rounded text-xs text-gray-200 uppercase">CMD + SHIFT + P</kbd> <code>Antisky: Tail Build Logs</code></li>
            <li><kbd className="bg-white/10 px-2 py-1 rounded text-xs text-gray-200 uppercase">CMD + SHIFT + P</kbd> <code>Antisky: Open Dashboard</code></li>
          </ul>
        </section>
      </div>
    </div>
  )
}
