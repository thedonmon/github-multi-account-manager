package main

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/donbowman/github-multi-account-manager/internal/config"
	"github.com/donbowman/github-multi-account-manager/internal/ssh"
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

	sshMgr, err := ssh.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SSH manager: %v\n", err)
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

	case "generate-key":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghmm-cli generate-key <account-name>")
			os.Exit(1)
		}
		generateKey(cfg, sshMgr, os.Args[2])

	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghmm-cli test <account-name>")
			os.Exit(1)
		}
		testConnection(cfg, sshMgr, os.Args[2])

	case "set-default":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghmm-cli set-default <account-name>")
			os.Exit(1)
		}
		setDefault(cfg, os.Args[2])

	case "setup":
		setupWizard(cfg, sshMgr)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GitHub Multi-Account Manager CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  ghmm-cli setup                                        # Guided setup (recommended)")
	fmt.Println("  ghmm-cli add <name> <username> <email> <directory>   # Add account manually")
	fmt.Println("  ghmm-cli list                                         # List all accounts")
	fmt.Println("  ghmm-cli generate-key <name>                          # Generate SSH key")
	fmt.Println("  ghmm-cli test <name>                                  # Test GitHub connection")
	fmt.Println("  ghmm-cli set-default <name>                           # Set default account")
	fmt.Println("  ghmm-cli remove <name>                                # Remove account")
	fmt.Println("\nQuick Start:")
	fmt.Println("  ghmm-cli setup      # Guided setup - recommended for first-time users!")
	fmt.Println("\nManual Setup:")
	fmt.Println("  ghmm-cli add work john-work john@company.com ~/code/work")
	fmt.Println("  ghmm-cli generate-key work")
	fmt.Println("  ghmm-cli test work")
}

func addAccount(cfg *config.Config, name, username, email, directory string) {
	if err := cfg.AddAccount(name, username, email, directory); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Account '%s' added successfully!\n", name)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Generate SSH key: ghmm-cli generate-key", name)
	fmt.Println("  2. Add the key to GitHub (we'll help!)")
	fmt.Println("  3. Test connection: ghmm-cli test", name)
}

func generateKey(cfg *config.Config, sshMgr *ssh.Manager, name string) {
	accounts := cfg.ListAccounts()
	var account *config.Account

	for _, acc := range accounts {
		if acc.Name == name {
			account = &acc
			break
		}
	}

	if account == nil {
		fmt.Fprintf(os.Stderr, "❌ Account '%s' not found\n", name)
		os.Exit(1)
	}

	fmt.Printf("🔑 Generating SSH key for '%s'...\n", name)

	// Generate the key
	if err := sshMgr.GenerateKey(account.SSHKeyPath, account.Email, "ed25519"); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to generate key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ SSH key generated: %s\n\n", account.SSHKeyPath)

	// Get and copy the public key
	pubKey, err := sshMgr.GetPublicKey(account.SSHKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read public key: %v\n", err)
		os.Exit(1)
	}

	if err := clipboard.WriteAll(pubKey); err != nil {
		fmt.Printf("⚠️  Could not copy to clipboard, but here's your public key:\n\n%s\n\n", pubKey)
	} else {
		fmt.Println("📋 Public key copied to clipboard!")
		fmt.Printf("\n%s\n\n", pubKey)
	}

	fmt.Println("🚀 Next steps:")
	fmt.Println("  1. Go to https://github.com/settings/keys")
	fmt.Println("  2. Click 'New SSH key'")
	fmt.Println("  3. Paste the key (already in clipboard!)")
	fmt.Println("  4. Give it a title like:", name)
	fmt.Println()

	// Ask if they've added it
	fmt.Print("Have you added the key to GitHub? (y/n): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" || response == "yes" {
		fmt.Println("\n🧪 Testing connection...")
		testConnection(cfg, sshMgr, name)
	} else {
		fmt.Println("\n💡 Run 'ghmm-cli test", name, "' when you're ready to test")
	}
}

func testConnection(cfg *config.Config, sshMgr *ssh.Manager, name string) {
	accounts := cfg.ListAccounts()
	var account *config.Account

	for _, acc := range accounts {
		if acc.Name == name {
			account = &acc
			break
		}
	}

	if account == nil {
		fmt.Fprintf(os.Stderr, "❌ Account '%s' not found\n", name)
		os.Exit(1)
	}

	fmt.Printf("🧪 Testing SSH connection for '%s'...\n", name)

	success, message := sshMgr.TestConnection(account.HostAlias)

	if success {
		fmt.Printf("✅ Success! %s\n", message)
		fmt.Println("\n🎉 Your account is ready to use!")
		fmt.Println("\n💡 Next steps:")
		fmt.Println("  • Run 'ghmm' to see all accounts and apply configs")
		fmt.Println("  • Or set this as default: ghmm-cli set-default", name)
	} else {
		fmt.Printf("❌ Connection failed: %s\n", message)
		fmt.Println("\n🔍 Troubleshooting:")
		fmt.Println("  • Make sure you added the SSH key to GitHub")
		fmt.Println("  • Check: https://github.com/settings/keys")
		fmt.Println("  • Verify the key matches:", account.SSHKeyPath+".pub")
		fmt.Println("  • Make sure SSH config is updated: ghmm (TUI) -> press 'a'")
	}
}

func setDefault(cfg *config.Config, name string) {
	if err := cfg.SetDefaultAccount(name); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Default account set to '%s'\n", name)
}

func setupWizard(cfg *config.Config, sshMgr *ssh.Manager) {
	fmt.Println("🚀 GitHub Multi-Account Manager - Setup Wizard")
	fmt.Println()
	fmt.Println("Let's set up your GitHub accounts! This wizard will:")
	fmt.Println("  1. Create an account configuration")
	fmt.Println("  2. Generate an SSH key")
	fmt.Println("  3. Help you add it to GitHub")
	fmt.Println("  4. Test the connection")
	fmt.Println()

	accountNum := 1
	for {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Account #%d\n", accountNum)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		// Get account details
		var name, username, email, directory string

		fmt.Print("Account name (e.g., work, personal): ")
		fmt.Scanln(&name)
		if name == "" {
			fmt.Println("❌ Account name is required")
			continue
		}

		fmt.Print("GitHub username: ")
		fmt.Scanln(&username)
		if username == "" {
			fmt.Println("❌ GitHub username is required")
			continue
		}

		fmt.Print("Email address: ")
		fmt.Scanln(&email)
		if email == "" {
			fmt.Println("❌ Email is required")
			continue
		}

		fmt.Printf("Directory path (default: ~/code/%s): ", name)
		fmt.Scanln(&directory)
		if directory == "" {
			home, _ := os.UserHomeDir()
			directory = fmt.Sprintf("%s/code/%s", home, name)
		}

		fmt.Println()

		// Add the account
		fmt.Printf("✓ Creating account '%s'...\n", name)
		if err := cfg.AddAccount(name, username, email, directory); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			continue
		}

		// Get the account we just added
		accounts := cfg.ListAccounts()
		var account *config.Account
		for _, acc := range accounts {
			if acc.Name == name {
				account = &acc
				break
			}
		}

		// Generate SSH key
		fmt.Printf("✓ Generating SSH key...\n")
		if err := sshMgr.GenerateKey(account.SSHKeyPath, account.Email, "ed25519"); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to generate key: %v\n", err)
			fmt.Println("You can generate it later with: ghmm-cli generate-key", name)
		} else {
			fmt.Printf("✓ SSH key created: %s\n\n", account.SSHKeyPath)

			// Get and copy the public key
			pubKey, err := sshMgr.GetPublicKey(account.SSHKeyPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to read public key: %v\n", err)
			} else {
				if err := clipboard.WriteAll(pubKey); err != nil {
					fmt.Printf("⚠️  Could not copy to clipboard\n\n")
					fmt.Println("Your public key:")
					fmt.Println(pubKey)
				} else {
					fmt.Println("📋 Public key copied to clipboard!")
					fmt.Println()
					fmt.Println("Your public key:")
					fmt.Println(pubKey)
				}

				fmt.Println()
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println("🔑 Next: Add this key to GitHub")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println("  1. Go to: https://github.com/settings/keys")
				fmt.Println("  2. Click 'New SSH key'")
				fmt.Println("  3. Title:", name, "(or any name you like)")
				fmt.Println("  4. Paste the key above (already in clipboard!)")
				fmt.Println("  5. Click 'Add SSH key'")
				fmt.Println()

				// Wait for user to add key
				fmt.Print("Have you added the key to GitHub? (y/n): ")
				var response string
				fmt.Scanln(&response)
				fmt.Println()

				if response == "y" || response == "Y" || response == "yes" {
					// Apply configs first
					fmt.Println("✓ Applying SSH and Git configurations...")
					if err := sshMgr.UpdateSSHConfig(cfg.ListAccounts()); err != nil {
						fmt.Fprintf(os.Stderr, "⚠️  Failed to update SSH config: %v\n", err)
					}

					// Test connection
					fmt.Printf("🧪 Testing connection for '%s'...\n", name)
					success, message := sshMgr.TestConnection(account.HostAlias)

					if success {
						fmt.Printf("✅ Success! %s\n\n", message)
						fmt.Println("🎉 Account '%s' is fully configured and ready!\n", name)
					} else {
						fmt.Printf("❌ Connection failed: %s\n\n", message)
						fmt.Println("🔍 Troubleshooting:")
						fmt.Println("  • Make sure the key was added to the correct GitHub account")
						fmt.Println("  • Try running: ghmm-cli test", name)
						fmt.Println()
					}
				} else {
					fmt.Println("💡 No problem! Test the connection later with:")
					fmt.Println("   ghmm-cli test", name)
					fmt.Println()
				}
			}
		}

		// Ask about another account
		fmt.Print("Add another account? (y/n): ")
		var addAnother string
		fmt.Scanln(&addAnother)
		fmt.Println()

		if addAnother != "y" && addAnother != "Y" && addAnother != "yes" {
			break
		}

		accountNum++
	}

	// Final summary
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Your accounts:")
	listAccounts(cfg)
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("  • Run 'ghmm' to launch the TUI and apply all configurations")
	fmt.Println("  • Or manually apply: ghmm-cli and press 'a'")
	fmt.Println()
	fmt.Println("📚 Learn more: https://github.com/thedonmon/github-multi-account-manager")
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
