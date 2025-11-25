"""Configuration management for ghmm."""

import os
from pathlib import Path
from typing import Dict, List, Optional
import yaml


class Config:
    """Manages ghmm configuration."""

    def __init__(self, config_dir: Optional[Path] = None):
        self.config_dir = config_dir or Path.home() / ".ghmm"
        self.config_file = self.config_dir / "config.yaml"
        self.config_dir.mkdir(parents=True, exist_ok=True)
        self._load()

    def _load(self):
        """Load configuration from file."""
        if self.config_file.exists():
            with open(self.config_file, "r") as f:
                self.data = yaml.safe_load(f) or {}
        else:
            self.data = {"accounts": [], "default_account": None}
            self._save()

    def _save(self):
        """Save configuration to file."""
        with open(self.config_file, "w") as f:
            yaml.dump(self.data, f, default_flow_style=False, sort_keys=False)

    def add_account(
        self,
        name: str,
        username: str,
        email: str,
        directory: str,
        ssh_key_path: Optional[str] = None,
    ) -> bool:
        """Add a new GitHub account."""
        # Check if account already exists
        if any(acc["name"] == name for acc in self.data["accounts"]):
            return False

        account = {
            "name": name,
            "username": username,
            "email": email,
            "directory": str(Path(directory).expanduser()),
            "ssh_key_path": ssh_key_path or str(Path.home() / ".ssh" / f"{name}_ssh"),
            "host_alias": f"github.com-{name}",
        }

        self.data["accounts"].append(account)
        
        # Set as default if it's the first account
        if len(self.data["accounts"]) == 1:
            self.data["default_account"] = name
        
        self._save()
        return True

    def remove_account(self, name: str) -> bool:
        """Remove a GitHub account."""
        original_len = len(self.data["accounts"])
        self.data["accounts"] = [acc for acc in self.data["accounts"] if acc["name"] != name]
        
        if len(self.data["accounts"]) < original_len:
            # Reset default if we removed it
            if self.data["default_account"] == name:
                self.data["default_account"] = (
                    self.data["accounts"][0]["name"] if self.data["accounts"] else None
                )
            self._save()
            return True
        return False

    def get_account(self, name: str) -> Optional[Dict]:
        """Get account by name."""
        for account in self.data["accounts"]:
            if account["name"] == name:
                return account
        return None

    def list_accounts(self) -> List[Dict]:
        """List all accounts."""
        return self.data["accounts"]

    def set_default_account(self, name: str) -> bool:
        """Set the default account."""
        if any(acc["name"] == name for acc in self.data["accounts"]):
            self.data["default_account"] = name
            self._save()
            return True
        return False

    def get_default_account(self) -> Optional[str]:
        """Get the default account name."""
        return self.data.get("default_account")
