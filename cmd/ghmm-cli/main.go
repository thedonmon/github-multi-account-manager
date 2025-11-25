package main

import (
	"fmt"
	"os"

	"github.com/donbowman/github-multi-account-manager/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	cfg, err := config.New("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "add":
		if len(os.Args) < 6 {
			fmt.Println("Usage: ghmm-cli add <name> <username> <email> <directory>")
			os.Exit(1)
		}
		addAccount(cfg, os.Args[2], os.Args[3], os.Args[4], os.Args[5])

	case "list":
		listAccounts(cfg)

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghmm-cli remove <name>")
			os.Exit(1)
		}
		removeAccount(cfg, os.Args[2])

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GitHub Multi-Account Manager CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  ghmm-cli add <name> <username> <email> <directory>")
	fmt.Println("  ghmm-cli list")
	fmt.Println("  ghmm-cli remove <name>")
	fmt.Println("\nExamples:")
	fmt.Println("  ghmm-cli add work john-work john@company.com ~/code/work")
	fmt.Println("  ghmm-cli add personal john john@gmail.com ~/code/personal")
	fmt.Println("  ghmm-cli list")
	fmt.Println("  ghmm-cli remove work")
}

func addAccount(cfg *config.Config, name, username, email, directory string) {
	if err := cfg.AddAccount(name, username, email, directory); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Account '%s' added successfully!\n", name)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Run 'ghmm' to launch the TUI")
	fmt.Println("  2. Generate SSH keys for your accounts")
	fmt.Println("  3. Add SSH keys to GitHub")
}

func listAccounts(cfg *config.Config) {
	accounts := cfg.ListAccounts()
	defaultAcc := cfg.GetDefaultAccount()

	if len(accounts) == 0 {
		fmt.Println("No accounts configured yet.")
		fmt.Println("\nRun 'ghmm-cli add <name> <username> <email> <directory>' to add one!")
		return
	}

	fmt.Println("\n📦 GitHub Accounts:\n")

	for _, acc := range accounts {
		isDefault := acc.Name == defaultAcc
		prefix := "  "
		if isDefault {
			prefix = "⭐"
		}

		fmt.Printf("%s %s\n", prefix, acc.Name)
		fmt.Printf("     Username:  %s\n", acc.Username)
		fmt.Printf("     Email:     %s\n", acc.Email)
		fmt.Printf("     Directory: %s\n", acc.Directory)
		fmt.Printf("     SSH Key:   %s\n", acc.SSHKeyPath)
		fmt.Println()
	}
}

func removeAccount(cfg *config.Config, name string) {
	if err := cfg.RemoveAccount(name); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Account '%s' removed\n", name)
}
