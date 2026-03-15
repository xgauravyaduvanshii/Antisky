# 🌌 Antisky Cloud Platform

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Antisky-6366f1?style=for-the-badge&logo=cloud&logoColor=white" />
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/Next.js-15-black?style=for-the-badge&logo=next.js&logoColor=white" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" />
</p>

<p align="center">
  <strong>The Ultimate World-Class Distributed Hosting Platform</strong><br/>
  Deploy websites, APIs, and full-stack applications effortlessly at infinite scale. <br/>Think Vercel meets Heroku, built for unlimited bare-metal server fleets.
</p>

---

## ✨ Unmatched Capabilities

| Feature | Description |
|---------|-------------|
| 🚀 **Instant Git Deploys** | Push code. We handle the containerization, build, and global distribution. |
| 🌍 **Multi-Framework** | Next.js, Node, Go, Python, PHP, Ruby, Rust, Java, Static rendering out-of-the-box. |
| 🖥️ **Distributed Fleet** | Attach unlimited physical servers across global regions. Scale horizontally forever. |
| 🔐 **Enterprise Auth** | Built-in JWT, OAuth (Google/GitHub/GitLab), granular API Keys, and 2FA. |
| 📊 **Command Center** | State-of-the-art Admin Panel for complete platform oversight and user management. |
| ⌨️ **Global Terminal** | Web-based SSH proxy. Access any fleet node securely directly from your browser. |
| 💳 **Razorpay Billing** | Fully automated subscriptions, metered usage tracking, and automated webhooks. |
| 🔧 **Builder System** | Zero-config build orchestra that natively detects and compiles source code. |

---

## 📚 Comprehensive Documentation

Dive deep into the platform mechanics with our detailed technical guides:

- 🏛 **[Architecture Overview](docs/architecture.md)** — High-level system design and component interaction.
- ⚙️ **[Backend Microservices](docs/services.md)** — In-depth look at Auth, Control Plane, Builder, Server Manager, and Billing APIs.
- 💻 **[Frontend Applications](docs/apps.md)** — Breakdown of the User Dashboard and Admin Panel UI/UX logic.
- 🚀 **[Deployment Guide](docs/deployment.md)** — How to host Antisky locally or on production AWS using Terraform.

---

## 🚀 Quick Start (Local Development)

```bash
# 1. Clone the platform
git clone https://github.com/xgauravyaduvanshii/Antisky.git
cd Antisky

# 2. Configure Environment
cp .env.example .env
# Important: Open .env and add your Razorpay Live/Test Keys

# 3. Spin up the distributed cluster (Requires Docker)
docker compose up --build -d
```

### 📍 Access Points

- 🖥️ **User Dashboard:** `http://localhost:3000`
- 🛡️ **Admin Panel:** `http://localhost:3001`
- 🔑 **Auth API:** `http://localhost:8081`
- 🌐 **Control API:** `http://localhost:8082`

---

## 🛠️ State-of-the-Art Tech Stack

- **Backend:** Go 1.22 (Chi router, pgx async drivers, stateless JWT)
- **Frontend:** Next.js 15, React 19, TypeScript
- **Styling:** Highly custom Vanilla CSS Engine (Glassmorphism, Dark/Light/Ocean Themes)
- **Database:** PostgreSQL 16 (22 normalized tables)
- **Cache:** Redis Cloud (Real-time Pub/Sub)
- **Payments:** Razorpay API Integration
- **Infrastructure:** Docker, Docker Compose, AWS Terraform

---

## 📄 License & Open Source

Licensed under the [MIT License](LICENSE). Free for both personal and enterprise commercial use.

---

<p align="center">Built with immense passion ❤️ by <strong>Gaurav Yaduvanshi</strong></p>
