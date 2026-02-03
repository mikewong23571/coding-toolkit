package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func requireBinary(name string) error {
	if name == "" {
		return fmt.Errorf("missing dependency: <empty>")
	}
	if strings.ContainsRune(name, os.PathSeparator) {
		info, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf("missing dependency: %s", name)
		}
		if info.IsDir() {
			return fmt.Errorf("dependency is a directory: %s", name)
		}
		return nil
	}
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("missing dependency: %s", name)
	}
	return nil
}

func requireZellij() error {
	return requireBinary(app.Zellij.Bin)
}

func requireGit() error {
	return requireBinary("git")
}

func requireFzf() error {
	return requireBinary("fzf")
}
