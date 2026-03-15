# 🌌 Antisky

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Antisky-6366f1?style=for-the-badge&logo=cloud&logoColor=white" />
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/Next.js-15-black?style=for-the-badge&logo=next.js&logoColor=white" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" />
</p>

<p align="center">
  <strong>A world-class, distributed hosting platform</strong><br/>
  Deploy websites, APIs, and full-stack applications at scale — like Vercel meets Heroku, built for unlimited server fleets.
</p>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🚀 **Instant Deployments** | Push to Git → automatic builds & deploys |
| 🌍 **Multi-Language** | Node.js, Go, Python, PHP, Ruby, Rust, Java, .NET, Static |
| 🖥️ **Distributed Fleet** | Unlimited server scaling across regions |
| 🔐 **Enterprise Auth** | JWT, OAuth (GitHub/Google/GitLab/Bitbucket), API Keys |
| 📊 **Admin Panel** | Full platform control — users, servers, billing, terminal |
| ⌨️ **Web Terminal** | SSH-in-browser to any server node |
| 💳 **Stripe Billing** | Subscriptions, usage metering, webhooks |
| 🔧 **Builder System** | One-command server provisioning |
| 🧩 **VS Code Extension** | Deploy & manage from your editor |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      FRONTENDS                          │
│  Dashboard (:3000)  │  Admin Panel (:3001)  │  CLI      │
└─────────┬───────────┴───────────┬───────────┴───────────┘
          │                       │
┌─────────▼───────────────────────▼───────────────────────┐
│                   CORE SERVICES (Go)                    │
│  Auth (:8081)  │  Control Plane (:8082)  │  Billing     │
│  Server Manager (:8083)  │  Build Orchestrator          │
└─────────┬───────────────────────┬───────────────────────┘
          │                       │
┌─────────▼───────────┐ ┌────────▼────────────────────────┐
│  PostgreSQL (22 tbl) │ │     Server Fleet                │
│  Redis Cloud         │ │  Agent + Terminal per node      │
└──────────────────────┘ └─────────────────────────────────┘
```

---

## 🚀 Quick Start

```bash
# Clone
git clone https://github.com/xgauravyaduvanshii/Antisky.git
cd Antisky

# Start all services (requires Docker)
docker compose up --build -d

# Access
# Dashboard:  http://localhost:3000
# Admin:      http://localhost:3001
# Auth API:   http://localhost:8081
# API:        http://localhost:8082
```

---

## 📁 Project Structure

```
antisky/
├── apps/
│   ├── dashboard/      # User dashboard (Next.js 15)
│   └── admin/          # Admin panel (Next.js 15)
├── services/
│   ├── auth/           # Authentication (Go) — JWT, OAuth, Sessions
│   ├── control-plane/  # Projects, Deployments, Orgs (Go)
│   ├── build-orchestrator/  # Build queue & language detection (Go)
│   ├── server-manager/     # Fleet management & admin API (Go)
│   └── billing/        # Stripe subscriptions & usage (Go)
├── builder/            # Server node provisioning package
│   ├── install.sh      # One-command server setup
│   ├── start-server.sh # Bootstrap & auto-register
│   └── services/       # Agent + Terminal proxy
├── tools/
│   ├── cli/            # Antisky CLI (Go)
│   └── vscode-extension/  # VS Code extension
├── infra/              # Terraform + Docker templates
└── docker-compose.yml  # Full local development stack
```

---

## 🛠️ Tech Stack

- **Backend:** Go 1.22 (Chi router, pgx, JWT)
- **Frontend:** Next.js 15, React 19, TypeScript
- **Database:** PostgreSQL 16 (22 tables)
- **Cache:** Redis Cloud
- **Payments:** Stripe
- **Container:** Docker, Docker Compose
- **IaC:** Terraform (AWS)
- **Monitoring:** Prometheus + Grafana (planned)

---

## 📄 License

[MIT License](LICENSE) — free for personal and commercial use.

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

<p align="center">Built with ❤️ by <strong>Gaurav Yaduvanshi</strong></p>
