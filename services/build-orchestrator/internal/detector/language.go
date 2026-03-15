package detector

import (
	"strings"
)

// DetectedProject holds the detected language, framework, and suggested commands
type DetectedProject struct {
	Runtime      string `json:"runtime"`
	Framework    string `json:"framework,omitempty"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
	OutputDir    string `json:"output_dir,omitempty"`
	Dockerfile   bool   `json:"has_dockerfile"`
}

// DetectFromFiles analyzes a list of filenames to determine the project type
func DetectFromFiles(files []string) *DetectedProject {
	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[strings.ToLower(f)] = true
	}

	// Priority 1: Dockerfile
	if fileSet["dockerfile"] {
		return &DetectedProject{
			Runtime:    "docker",
			Dockerfile: true,
		}
	}

	// Priority 2: Node.js (check for framework-specific files)
	if fileSet["package.json"] {
		return detectNodeFramework(files, fileSet)
	}

	// Priority 3: Go
	if fileSet["go.mod"] {
		return &DetectedProject{
			Runtime:      "go",
			BuildCommand: "go build -o app .",
			StartCommand: "./app",
		}
	}

	// Priority 4: Python
	if fileSet["requirements.txt"] || fileSet["pipfile"] || fileSet["pyproject.toml"] {
		return detectPythonFramework(files, fileSet)
	}

	// Priority 5: PHP
	if fileSet["composer.json"] {
		return detectPHPFramework(files, fileSet)
	}

	// Priority 6: Ruby
	if fileSet["gemfile"] {
		return detectRubyFramework(files, fileSet)
	}

	// Priority 7: Rust
	if fileSet["cargo.toml"] {
		return &DetectedProject{
			Runtime:      "rust",
			BuildCommand: "cargo build --release",
			StartCommand: "./target/release/app",
		}
	}

	// Priority 8: Java
	if fileSet["pom.xml"] || fileSet["build.gradle"] {
		return &DetectedProject{
			Runtime:      "java",
			BuildCommand: "mvn package -DskipTests",
			StartCommand: "java -jar target/*.jar",
		}
	}

	// Priority 9: .NET
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f), ".csproj") {
			return &DetectedProject{
				Runtime:      "dotnet",
				BuildCommand: "dotnet publish -c Release -o out",
				StartCommand: "dotnet out/*.dll",
			}
		}
	}

	// Priority 10: Static site (index.html present)
	if fileSet["index.html"] {
		return &DetectedProject{
			Runtime:    "static",
			OutputDir:  ".",
		}
	}

	return &DetectedProject{
		Runtime: "unknown",
	}
}

func detectNodeFramework(files []string, fileSet map[string]bool) *DetectedProject {
	p := &DetectedProject{
		Runtime: "nodejs",
	}

	// Check for Next.js
	if fileSet["next.config.js"] || fileSet["next.config.mjs"] || fileSet["next.config.ts"] {
		p.Framework = "nextjs"
		p.BuildCommand = "npm run build"
		p.StartCommand = "npm start"
		p.OutputDir = ".next"
		return p
	}

	// Check for Nuxt
	if fileSet["nuxt.config.js"] || fileSet["nuxt.config.ts"] {
		p.Framework = "nuxt"
		p.BuildCommand = "npm run build"
		p.StartCommand = "npm start"
		p.OutputDir = ".nuxt"
		return p
	}

	// Check for Remix
	if fileSet["remix.config.js"] || fileSet["remix.config.ts"] {
		p.Framework = "remix"
		p.BuildCommand = "npm run build"
		p.StartCommand = "npm start"
		return p
	}

	// Check for Astro
	if fileSet["astro.config.mjs"] || fileSet["astro.config.ts"] {
		p.Framework = "astro"
		p.BuildCommand = "npm run build"
		p.OutputDir = "dist"
		return p
	}

	// Check for SvelteKit
	if fileSet["svelte.config.js"] {
		p.Framework = "sveltekit"
		p.BuildCommand = "npm run build"
		p.StartCommand = "npm start"
		return p
	}

	// Check for Gatsby
	if fileSet["gatsby-config.js"] || fileSet["gatsby-config.ts"] {
		p.Framework = "gatsby"
		p.BuildCommand = "npm run build"
		p.OutputDir = "public"
		return p
	}

	// Check for Vite (generic SPA)
	if fileSet["vite.config.js"] || fileSet["vite.config.ts"] {
		p.Framework = "vite"
		p.BuildCommand = "npm run build"
		p.OutputDir = "dist"
		return p
	}

	// Check for Express/NestJS (server-side)
	if fileSet["nest-cli.json"] {
		p.Framework = "nestjs"
		p.BuildCommand = "npm run build"
		p.StartCommand = "npm start"
		return p
	}

	// Default Node.js
	p.BuildCommand = "npm run build"
	p.StartCommand = "npm start"
	return p
}

func detectPythonFramework(files []string, fileSet map[string]bool) *DetectedProject {
	p := &DetectedProject{
		Runtime: "python",
	}

	// Check for Django
	if fileSet["manage.py"] {
		p.Framework = "django"
		p.BuildCommand = "pip install -r requirements.txt && python manage.py collectstatic --noinput"
		p.StartCommand = "gunicorn config.wsgi:application"
		return p
	}

	// Check for FastAPI (look for main.py with uvicorn)
	if fileSet["main.py"] {
		p.Framework = "fastapi"
		p.BuildCommand = "pip install -r requirements.txt"
		p.StartCommand = "uvicorn main:app --host 0.0.0.0 --port 8000"
		return p
	}

	// Check for Streamlit
	if fileSet["streamlit_app.py"] {
		p.Framework = "streamlit"
		p.BuildCommand = "pip install -r requirements.txt"
		p.StartCommand = "streamlit run streamlit_app.py --server.port 8000"
		return p
	}

	// Default Python (Flask-like)
	p.BuildCommand = "pip install -r requirements.txt"
	p.StartCommand = "gunicorn app:app --bind 0.0.0.0:8000"
	return p
}

func detectPHPFramework(files []string, fileSet map[string]bool) *DetectedProject {
	p := &DetectedProject{
		Runtime: "php",
	}

	// Check for Laravel
	if fileSet["artisan"] {
		p.Framework = "laravel"
		p.BuildCommand = "composer install --no-dev --optimize-autoloader"
		p.StartCommand = "php artisan serve --host=0.0.0.0 --port=8000"
		return p
	}

	// Check for WordPress
	if fileSet["wp-config.php"] {
		p.Framework = "wordpress"
		p.BuildCommand = "composer install --no-dev"
		p.StartCommand = "php -S 0.0.0.0:8000"
		return p
	}

	// Check for Symfony
	if fileSet["symfony.lock"] {
		p.Framework = "symfony"
		p.BuildCommand = "composer install --no-dev --optimize-autoloader"
		p.StartCommand = "php bin/console server:run 0.0.0.0:8000"
		return p
	}

	// Default PHP
	p.BuildCommand = "composer install --no-dev"
	p.StartCommand = "php -S 0.0.0.0:8000 -t public"
	return p
}

func detectRubyFramework(files []string, fileSet map[string]bool) *DetectedProject {
	p := &DetectedProject{
		Runtime: "ruby",
	}

	// Check for Rails
	if fileSet["config.ru"] && fileSet["rakefile"] {
		p.Framework = "rails"
		p.BuildCommand = "bundle install && bundle exec rake assets:precompile"
		p.StartCommand = "bundle exec puma -C config/puma.rb"
		return p
	}

	// Check for Sinatra
	if fileSet["config.ru"] {
		p.Framework = "sinatra"
		p.BuildCommand = "bundle install"
		p.StartCommand = "bundle exec rackup -p 8000 -o 0.0.0.0"
		return p
	}

	// Default Ruby
	p.BuildCommand = "bundle install"
	p.StartCommand = "bundle exec ruby app.rb"
	return p
}
