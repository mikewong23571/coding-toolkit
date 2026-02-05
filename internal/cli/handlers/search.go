package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/util"
)

func SearchCmd() *cobra.Command {
	var cd bool
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search repos using fzf",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdSearch(app.Cfg, cd)
		},
	}
	cmd.Flags().BoolVar(&cd, "cd", false, "print cd command")
	cmd.Aliases = []string{"s"}
	return cmd
}

func cmdSearch(cfg config.Config, printCd bool) error {
	if !hasCommand("fzf") {
		return Die("missing dependency: fzf")
	}
	if !hasCommand("git") {
		return Die("missing dependency: git")
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return err
	}
	if _, err := os.Stat(root); err != nil {
		return Die(fmt.Sprintf("missing OXL_ROOT directory: %s", root))
	}

	pairs := findRepos(root)
	if len(pairs) == 0 {
		return errors.New("")
	}

	selection, err := runFzf(pairs)
	if err != nil {
		return err
	}
	if selection == "" {
		return errors.New("")
	}
	path := strings.SplitN(selection, "\t", 2)
	if len(path) < 2 {
		return nil
	}
	if printCd {
		fmt.Printf("cd %s\n", util.ShellQuote(path[1]))
	} else {
		fmt.Println(path[1])
	}
	return nil
}

func findRepos(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var pairs []string
	for _, layout := range entries {
		if !layout.IsDir() {
			continue
		}
		layoutName := layout.Name()
		if !domain.IsValidLayout(layoutName) {
			continue
		}
		layoutPath := filepath.Join(root, layoutName)
		repos, _ := os.ReadDir(layoutPath)
		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}
			repoPath := filepath.Join(layoutPath, repo.Name())
			if !isGitRepo(repoPath) {
				continue
			}
			rel := filepath.Join(layoutName, repo.Name())
			pairs = append(pairs, rel+"\t"+repoPath)
		}
	}
	return pairs
}

func isGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func runFzf(pairs []string) (string, error) {
	cmd := exec.Command("fzf", "--prompt=repo> ", "--delimiter=\t", "--with-nth=1", "--exit-0")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		w := bufio.NewWriter(stdin)
		for _, line := range pairs {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		stdin.Close()
	}()
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 0 {
				return "", nil
			}
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
