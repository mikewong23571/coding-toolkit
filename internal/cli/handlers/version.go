package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"owlx/assets"
	"owlx/internal/build"
)

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, _ []string) {
			out := cmd.OutOrStdout()
			commit := strings.TrimSpace(build.Commit)
			ts := strings.TrimSpace(build.Timestamp)
			if commit == "" {
				commit = "dev"
			}
			if ts == "" {
				ts = "unknown"
			}
			fmt.Fprintf(out, "owlx commit=%s timestamp=%s buildflag=%s\n", commit, ts, build.BuildFlag())

			data, ok := assets.EmbeddedTmuxMeta()
			if !ok {
				fmt.Fprintln(out, "tmux meta=unavailable")
				return
			}
			meta := parseEnv(data)
			fmt.Fprintf(
				out,
				"tmux version=%s musl=%s ncurses=%s libevent=%s\n",
				meta["TMUX_VERSION"],
				meta["MUSL_VERSION"],
				meta["NCURSES_VERSION"],
				meta["LIBEVENT_VERSION"],
			)
			conf := meta["TMUX_DEFAULT_CONF"]
			sock := meta["TMUX_DEFAULT_SOCK"]
			if conf != "" || sock != "" {
				fmt.Fprintf(out, "tmux defaults conf=%s sock=%s\n", conf, sock)
			}
		},
	}
	return cmd
}

func parseEnv(data []byte) map[string]string {
	out := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if key != "" {
			out[key] = val
		}
	}
	return out
}
