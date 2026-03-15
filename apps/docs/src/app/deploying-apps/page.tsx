import { Box, Code, Terminal, ServerCog, Copy } from 'lucide-react'

export default function DeployingApps() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-purple-500/10 text-purple-400 text-sm font-medium mb-6 border border-purple-500/20">
        <Box size={14} />
        Deployment Workflows
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        Deploying Applications
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        Antisky's Build Orchestrator natively supports over 30 frameworks including Next.js, Node.js, Go, Python, and PHP. We automatically detect your framework from the codebase and configure the optimal build environment.
      </p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Method 1: Git Integration (Recommended)</h2>
          <p className="text-gray-400 mb-6">
            The easiest way to deploy is directly from your GitHub, GitLab, or Bitbucket repository.
          </p>
          
          <div className="grid gap-4">
            <div className="p-4 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-2">
                <div className="w-6 h-6 rounded-full bg-purple-500/20 text-purple-400 flex items-center justify-center text-xs font-bold">1</div>
                <h3 className="text-white font-medium">Connect Provider</h3>
              </div>
              <p className="text-sm text-gray-400 pl-9">Navigate to your User Dashboard &rarr; <strong>Projects</strong> &rarr; <strong>New Project</strong>.</p>
            </div>
            
            <div className="p-4 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-2">
                <div className="w-6 h-6 rounded-full bg-purple-500/20 text-purple-400 flex items-center justify-center text-xs font-bold">2</div>
                <h3 className="text-white font-medium">Select Repository</h3>
              </div>
              <p className="text-sm text-gray-400 pl-9">Choose the repository and branch you wish to deploy. Antisky will scan the root for a <code>package.json</code> or <code>go.mod</code>.</p>
            </div>
            
            <div className="p-4 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-2">
                <div className="w-6 h-6 rounded-full bg-purple-500/20 text-purple-400 flex items-center justify-center text-xs font-bold">3</div>
                <h3 className="text-white font-medium">Configure Build Settings</h3>
              </div>
              <p className="text-sm text-gray-400 pl-9">Override the default build command (e.g., <code>npm run build</code>) or Output directory (e.g., <code>.next</code> or <code>build</code>).</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <Terminal className="text-purple-400" /> Method 2: Global CLI
          </h2>
          <p className="text-gray-400 mb-6">
            For advanced developers, you can deploy applications directly from your terminal using the Antisky CLI.
          </p>

          <div className="space-y-4">
            <div className="relative group">
              <div className="absolute inset-0 bg-gradient-to-r from-purple-500/20 to-indigo-500/20 rounded-xl blur " />
              <div className="relative bg-[#0d1117] border border-white/10 rounded-xl p-4 overflow-hidden">
                <div className="flex items-center justify-between mb-3 text-xs text-gray-500 font-medium">
                  <span className="flex items-center gap-2"><Terminal size={14} /> terminal</span>
                  <button className="hover:text-white transition-colors" title="Copy code"><Copy size={14} /></button>
                </div>
                <pre className="text-sm font-mono text-gray-300 block whitespace-pre-wrap">
                  <span className="text-purple-400">npm</span> install -g antisky-cli{'\n'}
                  <span className="text-gray-500"># Login to your cluster</span>{'\n'}
                  <span className="text-purple-400">antisky</span> login{'\n'}
                  <span className="text-gray-500"># Deploy the current directory</span>{'\n'}
                  <span className="text-purple-400">antisky</span> deploy --prod
                </pre>
              </div>
            </div>
          </div>
        </section>

        <div className="p-6 rounded-2xl bg-gradient-to-br from-indigo-500/10 to-purple-500/10 border border-indigo-500/20">
          <h3 className="text-lg font-bold text-white mb-2 flex items-center gap-2">
            <ServerCog size={20} className="text-indigo-400" /> Custom Dockerfiles
          </h3>
          <p className="text-gray-300">
            If your codebase contains a <code>Dockerfile</code> in the root directory, Antisky will automatically bypass the auto-framework detection and build your raw container natively, handling all layer caching and multi-stage builds.
          </p>
        </div>
      </div>
    </div>
  )
}
