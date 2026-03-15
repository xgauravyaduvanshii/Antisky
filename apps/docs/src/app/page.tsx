import { Rocket, Box, Server, Globe, Shield, Code, ChevronRight } from 'lucide-react'
import Link from 'next/link'

export default function Home() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700">
      
      {/* Hero Header */}
      <div className="mb-12 border-b border-white/5 pb-10">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 text-indigo-400 text-sm font-medium mb-6 border border-indigo-500/20">
          <Rocket size={14} />
          Welcome to Antisky
        </div>
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight mb-4 text-white">
          The Ultimate World-Class
          <br className="hidden md:block" />
          <span className="bg-gradient-to-r from-indigo-400 to-purple-400 bg-clip-text text-transparent">
            Distributed Hosting Platform
          </span>
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl leading-relaxed">
          Deploy websites, APIs, and full-stack applications effortlessly at infinite scale. 
          Think Vercel meets Heroku, built exclusively for unlimited bare-metal server fleets.
        </p>
      </div>

      {/* Feature Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-16">
        <FeatureCard 
          icon={<Server className="text-blue-400" />}
          title="Distributed Fleet"
          description="Attach unlimited physical servers across global regions. Scale horizontally forever."
        />
        <FeatureCard 
          icon={<Box className="text-purple-400" />}
          title="Multi-Framework"
          description="Next.js, Node, Go, Python, PHP, Ruby, Rust, Java. Static rendering out-of-the-box."
        />
        <FeatureCard 
          icon={<Globe className="text-emerald-400" />}
          title="Instant Deployments"
          description="Push code. We handle the containerization, build, and global content distribution."
        />
        <FeatureCard 
          icon={<Shield className="text-rose-400" />}
          title="Automated SSL"
          description="Bring your custom domains. We auto-provision wildcards and renew them indefinitely."
        />
      </div>

      {/* Quick Start Module */}
      <h2 className="text-2xl font-bold mb-6 text-white flex items-center gap-2">
        <Code className="text-indigo-400" />
        Quick Start Guides
      </h2>
      <div className="space-y-3">
        <GuideRow href="/add-server" title="Add your first server to the fleet" />
        <GuideRow href="/deploying-apps" title="Deploy a Next.js application" />
        <GuideRow href="/custom-domains" title="Connect a custom domain" />
        <GuideRow href="/vscode-extension" title="Install the VSCode Extension" />
      </div>

    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
  return (
    <div className="p-6 rounded-2xl bg-[#0a0a0f] border border-white/5 hover:border-indigo-500/30 transition-colors group">
      <div className="w-10 h-10 rounded-lg bg-white/5 flex items-center justify-center mb-4 group-hover:bg-indigo-500/10 transition-colors">
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
      <p className="text-sm text-muted-foreground leading-relaxed">{description}</p>
    </div>
  )
}

function GuideRow({ href, title }: { href: string, title: string }) {
  return (
    <Link href={href} className="flex items-center justify-between p-4 rounded-xl bg-white/5 border border-transparent hover:border-white/10 hover:bg-white/[0.07] transition-all group">
      <span className="font-medium text-gray-200 group-hover:text-white transition-colors">{title}</span>
      <ChevronRight size={18} className="text-gray-500 group-hover:text-indigo-400 transition-colors transform group-hover:translate-x-1" />
    </Link>
  )
}
