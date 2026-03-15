import { Globe, ShieldCheck, CheckSquare, Copy, Link as LinkIcon } from 'lucide-react'

export default function CustomDomains() {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 max-w-3xl">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 text-sm font-medium mb-6 border border-emerald-500/20">
        <Globe size={14} />
        Networking
      </div>
      <h1 className="text-4xl font-bold tracking-tight mb-4 text-white">
        Custom Domains & SSL
      </h1>
      <p className="text-lg text-muted-foreground mb-12 leading-relaxed">
        Map your custom domains to any project flawlessly. Antisky utilizes Let's Encrypt to instantly auto-provision and automatically renew SSL certificates across the globe.
      </p>

      <div className="space-y-12">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Adding a Domain</h2>
          <p className="text-gray-400 mb-6">
            Domains are attached on a per-project basis. Head over to <strong>Dashboard &rarr; Projects &rarr; Settings &rarr; Domains</strong>.
          </p>
          
          <div className="bg-white/5 border border-white/10 rounded-xl p-6">
            <ol className="space-y-4 list-decimal list-inside text-gray-300 marker:text-emerald-500 marker:font-bold">
              <li>Enter your domain (e.g., <code>api.example.com</code> or <code>example.com</code>).</li>
              <li>Click <strong>Add</strong>.</li>
              <li>Antisky will generate the DNS propagation record required for you to update at your registrar (GoDaddy, Namecheap, Cloudflare).</li>
            </ol>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <LinkIcon className="text-emerald-400" /> DNS Configuration
          </h2>
          <p className="text-gray-400 mb-6">
            You must configure your DNS provider to point to your cluster's load balancers. Depending on whether you are using an apex domain or a subdomain, configure the records accordingly.
          </p>

          <h3 className="text-lg font-bold text-white mb-3 mt-8">Subdomains (CNAME Record)</h3>
          <div className="relative group">
            <div className="absolute inset-0 bg-gradient-to-r from-emerald-500/10 to-teal-500/10 rounded-xl blur " />
            <div className="relative bg-[#0d1117] border border-white/10 rounded-xl overflow-hidden">
              <table className="w-full text-sm text-left text-gray-300">
                <thead className="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/10">
                  <tr>
                    <th className="px-6 py-3">Type</th>
                    <th className="px-6 py-3">Name</th>
                    <th className="px-6 py-3">Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-white/5">
                    <td className="px-6 py-4 font-mono text-emerald-400">CNAME</td>
                    <td className="px-6 py-4">api</td>
                    <td className="px-6 py-4">cname.antisky.app</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <h3 className="text-lg font-bold text-white mb-3 mt-8">Apex / Root Domains (A Record)</h3>
          <div className="relative group">
            <div className="absolute inset-0 bg-gradient-to-r from-emerald-500/10 to-teal-500/10 rounded-xl blur " />
            <div className="relative bg-[#0d1117] border border-white/10 rounded-xl overflow-hidden">
              <table className="w-full text-sm text-left text-gray-300">
                <thead className="text-xs text-gray-400 uppercase bg-white/5 border-b border-white/10">
                  <tr>
                    <th className="px-6 py-3">Type</th>
                    <th className="px-6 py-3">Name</th>
                    <th className="px-6 py-3">Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-white/5">
                    <td className="px-6 py-4 font-mono text-emerald-400">A</td>
                    <td className="px-6 py-4">@</td>
                    <td className="px-6 py-4">76.223.12.9<br/><span className="text-xs text-gray-500">(Your cluster's IP)</span></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section className="p-6 rounded-2xl bg-emerald-500/5 border border-emerald-500/20 flex flex-col items-center text-center">
          <div className="w-16 h-16 rounded-full bg-emerald-500/20 flex items-center justify-center mb-4">
            <ShieldCheck size={32} className="text-emerald-400" />
          </div>
          <h3 className="text-xl font-bold text-white mb-2">Automated Edge SSL</h3>
          <p className="text-gray-400 max-w-lg mx-auto">
            Once your DNS records propagate (usually within 2 minutes), Antisky Control Plane runs a background job intercepting ACME challenges to inject valid, auto-renewing TLS certificates terminating at the edge layer. <strong>Zero manual configuration needed.</strong>
          </p>
        </section>
      </div>
    </div>
  )
}
