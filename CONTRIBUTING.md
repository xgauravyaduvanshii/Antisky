# Contributing to Antisky

Thank you for your interest in contributing! 🎉

## Development Setup

```bash
# Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Node.js 20+

# Clone and start
git clone https://github.com/xgauravyaduvanshii/Antisky.git
cd Antisky
docker compose up --build -d
```

## Project Structure

- `services/` — Go microservices (auth, control-plane, billing, etc.)
- `apps/` — Next.js frontends (dashboard, admin)
- `builder/` — Server node provisioning
- `tools/` — CLI and VS Code extension

## How to Contribute

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feat/my-feature`
3. **Commit** your changes: `git commit -m "feat: add my feature"`
4. **Push** to the branch: `git push origin feat/my-feature`
5. **Open** a Pull Request

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation
- `refactor:` — Code refactoring
- `test:` — Tests
- `chore:` — Maintenance

## Code Style

- **Go**: `gofmt` and `golint`
- **TypeScript/React**: ESLint + Prettier
- **CSS**: Vanilla CSS with BEM-like naming

## Reporting Issues

Please use [GitHub Issues](https://github.com/xgauravyaduvanshii/Antisky/issues) with:
- Clear title
- Steps to reproduce
- Expected vs actual behavior
- Screenshots if applicable

## License

By contributing, you agree your code will be licensed under [MIT](LICENSE).
