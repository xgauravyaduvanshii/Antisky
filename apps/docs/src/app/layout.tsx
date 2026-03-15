import './globals.css'
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import Link from 'next/link'
import { Server, Code, Globe, Shield, Rocket, Activity, CreditCard, Box, Terminal } from 'lucide-react'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Antisky Documentation',
  description: 'The Ultimate Distributed Hosting Platform',
}

const NAV_GROUPS = [
  {
    title: 'OVERVIEW',
    items: [
      { name: 'Introduction', href: '/', icon: <Rocket size={16} /> },
    ]
  },
  {
    title: 'CORE GUIDES',
    items: [
      { name: 'Adding Servers (EC2/VPS)', href: '/add-server', icon: <Server size={16} /> },
      { name: 'Deploying Applications', href: '/deploying-apps', icon: <Box size={16} /> },
      { name: 'Custom Domains & SSL', href: '/custom-domains', icon: <Globe size={16} /> },
      { name: 'Using API Keys', href: '/api-keys', icon: <Shield size={16} /> },
      { name: 'VSCode Extension', href: '/vscode-extension', icon: <Code size={16} /> },
    ]
  },
  {
    title: 'PLATFORM ARCHITECTURE',
    items: [
      { name: 'Builder Node Orchestration', href: '/server-builder', icon: <Terminal size={16} /> },
      { name: 'System Monitoring', href: '#', icon: <Activity size={16} /> },
      { name: 'Razorpay Billing', href: '#', icon: <CreditCard size={16} /> },
    ]
  }
];

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={`${inter.className} bg-background text-foreground antialiased min-h-screen flex selection:bg-primary/30`}>
        {/* Glow Effect Background */}
        <div className="fixed inset-0 z-0 pointer-events-none overflow-hidden">
          <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-primary/10 blur-[120px]" />
          <div className="absolute bottom-[-10%] right-[-5%] w-[30%] h-[50%] rounded-full bg-indigo-500/5 blur-[100px]" />
        </div>

        {/* Sidebar */}
        <aside className="fixed inset-y-0 left-0 w-72 border-r border-white/5 bg-[#030305]/80 backdrop-blur-xl z-20 flex flex-col">
          <div className="h-16 flex items-center px-6 border-b border-white/5">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-indigo-600 flex items-center justify-center shadow-lg shadow-primary/20">
                <span className="text-white font-bold text-lg leading-none">A</span>
              </div>
              <span className="font-semibold text-lg tracking-tight">Antisky Docs</span>
            </div>
          </div>
          
          <div className="flex-1 overflow-y-auto py-6 px-4 space-y-8 no-scrollbar">
            {NAV_GROUPS.map((group) => (
              <div key={group.title}>
                <h4 className="px-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                  {group.title}
                </h4>
                <nav className="space-y-1">
                  {group.items.map((item) => (
                    <Link
                      key={item.href}
                      href={item.href}
                      className="flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-lg text-gray-300 hover:text-white hover:bg-white/5 transition-colors group"
                    >
                      <span className="text-gray-500 group-hover:text-primary transition-colors">{item.icon}</span>
                      {item.name}
                    </Link>
                  ))}
                </nav>
              </div>
            ))}
          </div>
          
          <div className="p-4 border-t border-white/5">
            <div className="text-xs text-muted-foreground text-center">
              v1.0.0 - MIT License
            </div>
          </div>
        </aside>

        {/* Main Content Area */}
        <main className="flex-1 pl-72 relative z-10 min-h-screen">
          <div className="max-w-4xl mx-auto px-12 py-16">
            {children}
          </div>
        </main>
      </body>
    </html>
  )
}
