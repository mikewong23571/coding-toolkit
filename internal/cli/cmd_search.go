package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/bridge"
	"owlx/internal/git"
)

func newCmdSearch() *cobra.Command {
	var printCD bool
	cmd := &cobra.Command{
		Use:   "search",
		Short: "search repos",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			root := app.Config.Root
			if root == "" {
				return errors.New("root is empty")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				return err
			}

			var lines []string
			for _, layoutEntry := range entries {
				if !layoutEntry.IsDir() {
					continue
				}
				layout := layoutEntry.Name()
				if !validLayouts[layout] {
					continue
				}
				layoutDir := filepath.Join(root, layout)
				repoEntries, err := os.ReadDir(layoutDir)
				if err != nil {
					continue
				}
				for _, repoEntry := range repoEntries {
					if !repoEntry.IsDir() {
						continue
					}
					repo := repoEntry.Name()
					repoDir := filepath.Join(layoutDir, repo)
					if !git.IsRepo(ctx, repoDir) {
						continue
					}
					rel := filepath.Join(layout, repo)
					lines = append(lines, fmt.Sprintf("%s\t%s", rel, repoDir))
				}
			}
			if len(lines) == 0 {
				return nil
			}
			input := strings.Join(lines, "\n") + "\n"
			out, err := bridge.RunWithInput(ctx, input, "fzf", "--prompt=repo> ", "--delimiter=\t", "--with-nth=1", "--exit-0")
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					code := exitErr.ExitCode()
					if code == 1 || code == 130 {
						return nil
					}
				}
				return err
			}
			out = strings.TrimSpace(out)
			if out == "" {
				return nil
			}
			parts := strings.SplitN(out, "\t", 2)
			path := parts[0]
			if len(parts) == 2 {
				path = parts[1]
			}
			if printCD {
				fmt.Fprintf(cmd.OutOrStdout(), "cd %s\n", shellQuote(path))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&printCD, "cd", false, "print a shell cd command")
	return cmd
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	replacer := strings.NewReplacer("'", "'\"'\"'")
	return "'" + replacer.Replace(value) + "'"
}
