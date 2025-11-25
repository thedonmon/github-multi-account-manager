"""TUI interface for ghmm using Textual."""

import pyperclip
from textual.app import App, ComposeResult
from textual.containers import Container, Horizontal, Vertical
from textual.widgets import Button, DataTable, Footer, Header, Input, Label, Static
from textual.screen import ModalScreen
from rich.text import Text

from .config import Config
from .ssh import SSHManager
from .shell import ShellManager
from .gitconfig import GitConfigManager


class AddAccountScreen(ModalScreen):
    """Modal screen for adding a new account."""

    def compose(self) -> ComposeResult:
        yield Container(
            Label("Add New GitHub Account", id="dialog-title"),
            Input(placeholder="Account name (e.g., work, personal)", id="input-name"),
            Input(placeholder="GitHub username", id="input-username"),
            Input(placeholder="GitHub email", id="input-email"),
            Input(placeholder="Directory path (e.g., ~/code/work)", id="input-directory"),
            Horizontal(
                Button("Add", variant="success", id="btn-add"),
                Button("Cancel", variant="error", id="btn-cancel"),
                id="dialog-buttons",
            ),
            id="add-account-dialog",
        )

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-cancel":
            self.app.pop_screen()
        elif event.button.id == "btn-add":
            # Get values from inputs
            name = self.query_one("#input-name", Input).value
            username = self.query_one("#input-username", Input).value
            email = self.query_one("#input-email", Input).value
            directory = self.query_one("#input-directory", Input).value

            if not all([name, username, email, directory]):
                self.app.show_message("Error", "All fields are required!")
                return

            # Add the account
            result = self.app.add_account(name, username, email, directory)
            if result:
                self.app.pop_screen()
            else:
                self.app.show_message("Error", f"Account '{name}' already exists!")


class MessageScreen(ModalScreen):
    """Modal screen for displaying messages."""

    def __init__(self, title: str, message: str):
        super().__init__()
        self.title_text = title
        self.message_text = message

    def compose(self) -> ComposeResult:
        yield Container(
            Label(self.title_text, id="message-title"),
            Static(self.message_text, id="message-body"),
            Button("OK", variant="primary", id="btn-ok"),
            id="message-dialog",
        )

    def on_button_pressed(self, event: Button.Pressed) -> None:
        self.app.pop_screen()


class GHMMApp(App):
    """GitHub Multi-Account Manager TUI application."""

    CSS = """
    Screen {
        background: $surface;
    }

    #add-account-dialog, #message-dialog {
        width: 60;
        height: auto;
        border: thick $primary;
        background: $surface;
        padding: 1;
    }

    #dialog-title, #message-title {
        width: 100%;
        content-align: center middle;
        text-style: bold;
        color: $accent;
        margin-bottom: 1;
    }

    #message-body {
        width: 100%;
        margin-bottom: 1;
    }

    Input {
        margin: 1 0;
    }

    #dialog-buttons {
        width: 100%;
        height: auto;
        align: center middle;
    }

    #main-container {
        padding: 1;
    }

    #accounts-table {
        height: 1fr;
        margin: 1 0;
    }

    #actions {
        height: auto;
        margin: 1 0;
    }

    Button {
        margin: 0 1;
    }

    #status-bar {
        dock: bottom;
        height: 3;
        background: $panel;
        padding: 1;
    }
    """

    BINDINGS = [
        ("q", "quit", "Quit"),
        ("a", "add_account", "Add Account"),
        ("d", "delete_account", "Delete Account"),
        ("t", "test_connection", "Test Connection"),
        ("r", "refresh", "Refresh"),
    ]

    def __init__(self):
        super().__init__()
        self.config = Config()
        self.ssh_manager = SSHManager()
        self.shell_manager = ShellManager()
        self.git_manager = GitConfigManager()

    def compose(self) -> ComposeResult:
        yield Header()
        yield Container(
            Label("📦 GitHub Multi-Account Manager", id="title"),
            DataTable(id="accounts-table"),
            Horizontal(
                Button("➕ Add Account", variant="success", id="btn-add"),
                Button("🗑️ Delete Account", variant="error", id="btn-delete"),
                Button("🔑 Copy SSH Key", variant="primary", id="btn-copy-key"),
                Button("✅ Test Connection", variant="primary", id="btn-test"),
                Button("⚙️ Apply Config", variant="warning", id="btn-apply"),
                id="actions",
            ),
            Static("", id="status-bar"),
            id="main-container",
        )
        yield Footer()

    def on_mount(self) -> None:
        """Initialize the table when the app starts."""
        table = self.query_one(DataTable)
        table.add_columns("Name", "Username", "Email", "Directory", "Status")
        self.refresh_table()
        self.update_status(f"Detected shell: {self.shell_manager.shell_type}")

    def refresh_table(self) -> None:
        """Refresh the accounts table."""
        table = self.query_one(DataTable)
        table.clear()

        accounts = self.config.list_accounts()
        default = self.config.get_default_account()

        for account in accounts:
            # Check if SSH key exists
            from pathlib import Path
            key_exists = Path(account['ssh_key_path']).exists()
            status = "✅ Ready" if key_exists else "⚠️ No Key"

            # Mark default account
            name = account['name']
            if name == default:
                name = f"⭐ {name}"

            table.add_row(
                name,
                account['username'],
                account['email'],
                account['directory'],
                status,
            )

    def action_add_account(self) -> None:
        """Show the add account dialog."""
        self.push_screen(AddAccountScreen())

    def action_delete_account(self) -> None:
        """Delete the selected account."""
        table = self.query_one(DataTable)
        if table.cursor_row < 0:
            self.show_message("Error", "Please select an account to delete")
            return

        accounts = self.config.list_accounts()
        if table.cursor_row >= len(accounts):
            return

        account = accounts[table.cursor_row]
        self.config.remove_account(account['name'])
        self.git_manager.remove_account_gitconfig(account['name'])
        self.refresh_table()
        self.update_status(f"Deleted account: {account['name']}")

    def action_test_connection(self) -> None:
        """Test SSH connection for selected account."""
        table = self.query_one(DataTable)
        if table.cursor_row < 0:
            self.show_message("Error", "Please select an account to test")
            return

        accounts = self.config.list_accounts()
        if table.cursor_row >= len(accounts):
            return

        account = accounts[table.cursor_row]
        success, message = self.ssh_manager.test_connection(account['host_alias'])

        if success:
            self.show_message("✅ Connection Successful", message)
        else:
            self.show_message("❌ Connection Failed", message)

    def action_refresh(self) -> None:
        """Refresh the table."""
        self.refresh_table()
        self.update_status("Refreshed")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "btn-add":
            self.action_add_account()
        elif event.button.id == "btn-delete":
            self.action_delete_account()
        elif event.button.id == "btn-test":
            self.action_test_connection()
        elif event.button.id == "btn-copy-key":
            self.copy_ssh_key()
        elif event.button.id == "btn-apply":
            self.apply_all_config()

    def copy_ssh_key(self) -> None:
        """Copy SSH public key to clipboard."""
        table = self.query_one(DataTable)
        if table.cursor_row < 0:
            self.show_message("Error", "Please select an account")
            return

        accounts = self.config.list_accounts()
        if table.cursor_row >= len(accounts):
            return

        account = accounts[table.cursor_row]
        pub_key = self.ssh_manager.get_public_key(account['ssh_key_path'])

        if pub_key:
            try:
                pyperclip.copy(pub_key)
                self.show_message("✅ Copied!", f"SSH key copied to clipboard for {account['name']}")
            except Exception as e:
                self.show_message("Error", f"Failed to copy to clipboard: {str(e)}")
        else:
            self.show_message("Error", "SSH key not found. Generate one first!")

    def apply_all_config(self) -> None:
        """Apply all configurations (SSH, git, shell)."""
        accounts = self.config.list_accounts()
        default = self.config.get_default_account()

        # Update SSH config
        success, msg = self.ssh_manager.update_ssh_config(accounts)
        if not success:
            self.show_message("Error", msg)
            return

        # Update gitconfig
        success, msg = self.git_manager.update_gitconfig(accounts)
        if not success:
            self.show_message("Error", msg)
            return

        # Create individual gitconfigs
        for account in accounts:
            self.git_manager.create_account_gitconfig(account)

        # Update shell config
        success, msg = self.shell_manager.update_shell_config(accounts, default)
        if not success:
            self.show_message("Error", msg)
            return

        reload_cmd = self.shell_manager.get_reload_command()
        self.show_message(
            "✅ Configuration Applied!",
            f"All configs updated successfully!\\n\\nReload your shell with:\\n{reload_cmd}"
        )
        self.update_status("Configuration applied successfully")

    def add_account(self, name: str, username: str, email: str, directory: str) -> bool:
        """Add a new account with full setup."""
        # Add to config
        success = self.config.add_account(name, username, email, directory)
        if not success:
            return False

        account = self.config.get_account(name)
        
        # Generate SSH key
        key_success, key_msg = self.ssh_manager.generate_key(
            account['ssh_key_path'],
            email
        )

        if key_success:
            # Add to ssh-agent
            self.ssh_manager.add_to_ssh_agent(account['ssh_key_path'])
            self.update_status(f"Added account: {name} (SSH key generated)")
        else:
            self.update_status(f"Added account: {name} (SSH key generation failed)")

        self.refresh_table()
        return True

    def show_message(self, title: str, message: str) -> None:
        """Show a message dialog."""
        self.push_screen(MessageScreen(title, message))

    def update_status(self, message: str) -> None:
        """Update the status bar."""
        status = self.query_one("#status-bar", Static)
        status.update(message)


def run_tui():
    """Run the TUI application."""
    app = GHMMApp()
    app.run()
