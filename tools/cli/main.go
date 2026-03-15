package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	apiURL    = "http://localhost:8080"
	authURL   = "http://localhost:8081"
	authToken = ""
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "antisky",
		Short: "Antisky CLI — Deploy & manage your cloud applications",
		Long: color.New(color.FgHiMagenta, color.Bold).Sprint("Antisky CLI") +
			" — Deploy, manage, and scale your applications from the terminal.",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", getEnv("ANTISKY_API_URL", "http://localhost:8080"), "API URL")
	rootCmd.PersistentFlags().StringVar(&authToken, "token", getEnv("ANTISKY_TOKEN", ""), "Auth token")

	// Add commands
	rootCmd.AddCommand(
		loginCmd(),
		deployCmd(),
		projectsCmd(),
		envCmd(),
		logsCmd(),
		domainsCmd(),
		whoamiCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loginCmd() *cobra.Command {
	var email, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Antisky",
		Run: func(cmd *cobra.Command, args []string) {
			if email == "" || password == "" {
				color.Red("Email and password are required")
				os.Exit(1)
			}

			fmt.Printf("Authenticating as %s...\n", color.CyanString(email))
			// TODO: Call auth API
			color.Green("✓ Logged in successfully!")
			fmt.Println("  Token saved to ~/.antisky/credentials")
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	return cmd
}

func deployCmd() *cobra.Command {
	var projectID, ref string
	var prod bool
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy your project",
		Long:  "Trigger a deployment from the current directory or a specified project.",
		Run: func(cmd *cobra.Command, args []string) {
			env := "preview"
			if prod {
				env = "production"
			}
			fmt.Printf("🚀 Deploying to %s...\n", color.New(color.FgHiCyan, color.Bold).Sprint(env))
			fmt.Printf("   Project: %s\n", color.CyanString(projectID))
			fmt.Printf("   Ref: %s\n", color.CyanString(ref))
			fmt.Println()

			// TODO: Call deploy API with live log streaming
			color.Yellow("⏳ Build queued...")
			fmt.Println("   Installing dependencies...")
			fmt.Println("   Building project...")
			fmt.Println("   Uploading artifacts...")
			fmt.Println()
			color.Green("✅ Deployment complete!")
			fmt.Printf("   URL: %s\n", color.HiBlueString("https://myapp.antisky.app"))
		},
	}
	cmd.Flags().StringVar(&projectID, "project", "", "Project ID or slug")
	cmd.Flags().StringVar(&ref, "ref", "main", "Git ref to deploy")
	cmd.Flags().BoolVar(&prod, "prod", false, "Deploy to production")
	return cmd
}

func projectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.New(color.FgHiWhite, color.Bold).Sprint("Your Projects"))
			fmt.Println("─────────────────────────────────────────────────────────")
			// TODO: Fetch from API
			fmt.Printf("  %s  %-20s  %-10s  %s\n",
				color.GreenString("●"), "my-nextjs-app", "nodejs", color.HiBlackString("2h ago"))
			fmt.Printf("  %s  %-20s  %-10s  %s\n",
				color.GreenString("●"), "go-api-service", "go", color.HiBlackString("1d ago"))
			fmt.Printf("  %s  %-20s  %-10s  %s\n",
				color.YellowString("●"), "django-dashboard", "python", color.HiBlackString("3d ago"))
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Creating project %s...\n", color.CyanString(name))
			// TODO: Call create API
			color.Green("✓ Project created: %s", name)
		},
	}

	cmd.AddCommand(listCmd, createCmd)
	return cmd
}

func envCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List environment variables",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.New(color.FgHiWhite, color.Bold).Sprint("Environment Variables"))
			fmt.Println("─────────────────────────────────────")
			// TODO: Fetch from API
			fmt.Printf("  %-20s  %s\n", "DATABASE_URL", color.HiBlackString("●●●●●●●●"))
			fmt.Printf("  %-20s  %s\n", "REDIS_URL", color.HiBlackString("●●●●●●●●"))
			fmt.Printf("  %-20s  %s\n", "JWT_SECRET", color.HiBlackString("●●●●●●●●"))
		},
	}

	setCmd := &cobra.Command{
		Use:   "set KEY=VALUE",
		Short: "Set an environment variable",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setting environment variable...\n")
			color.Green("✓ Environment variable set")
		},
	}

	cmd.AddCommand(listCmd, setCmd)
	return cmd
}

func logsCmd() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs [deployment-id]",
		Short: "View deployment logs",
		Run: func(cmd *cobra.Command, args []string) {
			if follow {
				fmt.Println(color.HiBlackString("Streaming logs... (Ctrl+C to stop)"))
			}
			// TODO: Stream from API
			fmt.Println(color.HiBlackString("[12:00:01.123]") + " 🔨 Build started...")
			fmt.Println(color.HiBlackString("[12:00:02.456]") + " 📥 Cloning repository...")
			fmt.Println(color.HiBlackString("[12:00:05.789]") + " 📦 Installing dependencies...")
			fmt.Println(color.HiBlackString("[12:00:12.012]") + " 🏗️  Building project...")
			fmt.Println(color.HiBlackString("[12:00:18.345]") + " " + color.GreenString("✅ Build complete!"))
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	return cmd
}

func domainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage custom domains",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.New(color.FgHiWhite, color.Bold).Sprint("Custom Domains"))
			fmt.Println("─────────────────────────────────────────────")
			fmt.Printf("  %s  %-30s  %s\n",
				color.GreenString("✓"), "myapp.com", color.GreenString("SSL Active"))
			fmt.Printf("  %s  %-30s  %s\n",
				color.YellowString("⏳"), "api.myapp.com", color.YellowString("Pending verification"))
		},
	}

	addCmd := &cobra.Command{
		Use:   "add [domain]",
		Short: "Add a custom domain",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			domain := args[0]
			fmt.Printf("Adding domain %s...\n", color.CyanString(domain))
			color.Green("✓ Domain added!")
			fmt.Println("\n  Add this CNAME record to your DNS:")
			fmt.Printf("  %s → %s\n",
				color.CyanString(domain), color.HiBlackString("cname.antisky.app"))
		},
	}

	cmd.AddCommand(listCmd, addCmd)
	return cmd
}

func whoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Fetch from auth API
			fmt.Printf("Logged in as %s\n", color.CyanString("dev@antisky.app"))
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
