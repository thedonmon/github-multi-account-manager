"""CLI interface for ghmm."""

import click
from pathlib import Path

from .config import Config
from .ssh import SSHManager
from .shell import ShellManager
from .gitconfig import GitConfigManager


@click.group()
@click.version_option()
def cli():
    """GitHub Multi-Account Manager - Easily manage multiple GitHub accounts."""
    pass


@cli.command()
def tui():
    """Launch the interactive TUI."""
    from .tui import run_tui
    run_tui()


@cli.command()
@click.option("--name", prompt="Account name", help="Account identifier (e.g., work, personal)")
@click.option("--username", prompt="GitHub username", help="Your GitHub username")
@click.option("--email", prompt="GitHub email", help="Your GitHub email")
@click.option("--directory", prompt="Directory path", help="Directory for this account")
def add_account(name, username, email, directory):
    """Add a new GitHub account."""
    config = Config()
    ssh_manager = SSHManager()
    git_manager = GitConfigManager()
    
    # Add to config
    success = config.add_account(name, username, email, directory)
    if not success:
        click.echo(click.style(f"❌ Account '{name}' already exists!", fg="red"))
        return
    
    account = config.get_account(name)
    
    # Generate SSH key
    click.echo(f"🔑 Generating SSH key for {name}...")
    key_success, key_msg = ssh_manager.generate_key(account['ssh_key_path'], email)
    
    if key_success:
        click.echo(click.style(f"✅ {key_msg}", fg="green"))
        
        # Add to ssh-agent
        ssh_manager.add_to_ssh_agent(account['ssh_key_path'])
        
        # Show public key
        pub_key = ssh_manager.get_public_key(account['ssh_key_path'])
        if pub_key:
            click.echo("\n📋 Your SSH public key (add this to GitHub):")
            click.echo(click.style(pub_key, fg="cyan"))
            click.echo("\n👉 Go to: https://github.com/settings/keys")
    else:
        click.echo(click.style(f"❌ {key_msg}", fg="red"))
    
    click.echo(click.style(f"\n✅ Account '{name}' added successfully!", fg="green"))
    click.echo("💡 Run 'ghmm apply-config' to update all configurations")


@cli.command()
@click.argument("account_name")
def remove_account(account_name):
    """Remove a GitHub account."""
    config = Config()
    git_manager = GitConfigManager()
    
    success = config.remove_account(account_name)
    if success:
        git_manager.remove_account_gitconfig(account_name)
        click.echo(click.style(f"✅ Account '{account_name}' removed", fg="green"))
        click.echo("💡 Run 'ghmm apply-config' to update configurations")
    else:
        click.echo(click.style(f"❌ Account '{account_name}' not found", fg="red"))


@cli.command()
def list_accounts():
    """List all configured accounts."""
    config = Config()
    ssh_manager = SSHManager()
    accounts = config.list_accounts()
    default = config.get_default_account()
    
    if not accounts:
        click.echo("No accounts configured yet.")
        click.echo("Run 'ghmm add-account' to add one!")
        return
    
    click.echo("\n📦 GitHub Accounts:\n")
    
    for account in accounts:
        is_default = account['name'] == default
        key_exists = Path(account['ssh_key_path']).exists()
        
        prefix = "⭐" if is_default else "  "
        status = "✅" if key_exists else "⚠️"
        
        click.echo(f"{prefix} {status} {click.style(account['name'], fg='cyan', bold=True)}")
        click.echo(f"     Username: {account['username']}")
        click.echo(f"     Email:    {account['email']}")
        click.echo(f"     Directory: {account['directory']}")
        click.echo(f"     SSH Key:  {account['ssh_key_path']}")
        click.echo()


@cli.command()
@click.argument("account_name")
def set_default(account_name):
    """Set the default GitHub account."""
    config = Config()
    
    success = config.set_default_account(account_name)
    if success:
        click.echo(click.style(f"✅ Default account set to '{account_name}'", fg="green"))
    else:
        click.echo(click.style(f"❌ Account '{account_name}' not found", fg="red"))


@cli.command()
@click.argument("account_name")
def show_key(account_name):
    """Show the SSH public key for an account."""
    config = Config()
    ssh_manager = SSHManager()
    
    account = config.get_account(account_name)
    if not account:
        click.echo(click.style(f"❌ Account '{account_name}' not found", fg="red"))
        return
    
    pub_key = ssh_manager.get_public_key(account['ssh_key_path'])
    if pub_key:
        click.echo(f"\n🔑 SSH Public Key for {account_name}:")
        click.echo(click.style(pub_key, fg="cyan"))
        click.echo("\n👉 Add this to: https://github.com/settings/keys")
    else:
        click.echo(click.style("❌ SSH key not found", fg="red"))
        click.echo(f"Run 'ghmm generate-key {account_name}' to create one")


@cli.command()
@click.argument("account_name")
def test_connection(account_name):
    """Test SSH connection to GitHub for an account."""
    config = Config()
    ssh_manager = SSHManager()
    
    account = config.get_account(account_name)
    if not account:
        click.echo(click.style(f"❌ Account '{account_name}' not found", fg="red"))
        return
    
    click.echo(f"🔌 Testing connection for {account_name}...")
    success, message = ssh_manager.test_connection(account['host_alias'])
    
    if success:
        click.echo(click.style(f"✅ {message}", fg="green"))
    else:
        click.echo(click.style(f"❌ {message}", fg="red"))


@cli.command()
def apply_config():
    """Apply all configurations (SSH, Git, Shell)."""
    config = Config()
    ssh_manager = SSHManager()
    shell_manager = ShellManager()
    git_manager = GitConfigManager()
    
    accounts = config.list_accounts()
    default = config.get_default_account()
    
    if not accounts:
        click.echo("No accounts configured yet.")
        return
    
    click.echo("⚙️ Applying configurations...\n")
    
    # Update SSH config
    click.echo("1️⃣ Updating SSH config...")
    success, msg = ssh_manager.update_ssh_config(accounts)
    if success:
        click.echo(click.style(f"   ✅ {msg}", fg="green"))
    else:
        click.echo(click.style(f"   ❌ {msg}", fg="red"))
        return
    
    # Update gitconfig
    click.echo("2️⃣ Updating git config...")
    success, msg = git_manager.update_gitconfig(accounts)
    if success:
        click.echo(click.style(f"   ✅ {msg}", fg="green"))
    else:
        click.echo(click.style(f"   ❌ {msg}", fg="red"))
        return
    
    # Create individual gitconfigs
    for account in accounts:
        git_manager.create_account_gitconfig(account)
    
    # Update shell config
    click.echo(f"3️⃣ Updating {shell_manager.shell_type} config...")
    success, msg = shell_manager.update_shell_config(accounts, default)
    if success:
        click.echo(click.style(f"   ✅ {msg}", fg="green"))
    else:
        click.echo(click.style(f"   ❌ {msg}", fg="red"))
        return
    
    click.echo(click.style("\n✨ All configurations applied successfully!", fg="green"))
    click.echo(f"\n🔄 Reload your shell with:")
    click.echo(click.style(f"   {shell_manager.get_reload_command()}", fg="cyan"))


@cli.command()
def info():
    """Show system information."""
    shell_manager = ShellManager()
    config = Config()
    
    click.echo("\n📊 System Information:\n")
    click.echo(f"Shell:        {shell_manager.shell_type}")
    click.echo(f"Config file:  {shell_manager.config_file}")
    click.echo(f"SSH config:   ~/.ssh/config")
    click.echo(f"Git config:   ~/.gitconfig")
    click.echo(f"ghmm config:  {config.config_file}")
    click.echo(f"Accounts:     {len(config.list_accounts())}")
    click.echo()


def main():
    """Main entry point."""
    cli()


if __name__ == "__main__":
    main()
