package handlers

import (
	"fmt"
	"strconv"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
	"owlx/internal/util"
)

func setupLayout(cfg config.Config, tm tmux.Manager, sess string) error {
	panes, err := tm.ListPanes(sess)
	if err != nil {
		return err
	}
	if len(panes) == 0 {
		return nil
	}
	base := panes[0]

	if cfg.StatusOnNew {
		if err := statusOn(cfg, tm, sess); err != nil {
			return err
		}
	}

	_, err = tm.SplitWindow("-v", "-p", strconv.Itoa(cfg.LeftPaneBottomPct), "-t", base)
	if err != nil {
		return err
	}

	tm.SelectPane(base)
	return nil
}

func statusIsOn(tm tmux.Manager, sess string) bool {
	return tm.ShowOption(sess, domain.StatusKey) == "1"
}

func statusOn(cfg config.Config, tm tmux.Manager, sess string) error {
	if statusIsOn(tm, sess) {
		return nil
	}

	if err := tm.SetOption(sess, domain.StatusSavedKey, tm.ShowOption(sess, "status")); err != nil {
		return err
	}
	if err := tm.SetOption(sess, domain.StatusFormat0Saved, tm.ShowOption(sess, "status-format[0]")); err != nil {
		return err
	}
	if err := tm.SetOption(sess, domain.StatusFormat1Saved, tm.ShowOption(sess, "status-format[1]")); err != nil {
		return err
	}

	lineFmt := cfg.StatusLineFmt
	if lineFmt == "" {
		lineFmt = fmt.Sprintf("#[align=left]%s#[align=right]%s", cfg.StatusLeftFmt, cfg.StatusRightFmt)
	}

	if err := tm.Run("set-option", "-t", sess, "status", strconv.Itoa(cfg.StatusLines)); err != nil {
		return err
	}
	if tm.ShowOption(sess, "status-format[0]") == "" {
		fmt0 := tm.OutputQuiet("show-option", "-gqv", "status-format[0]")
		if fmt0 != "" {
			if err := tm.Run("set-option", "-t", sess, "status-format[0]", fmt0); err != nil {
				return err
			}
		}
	}
	if err := tm.Run("set-option", "-t", sess, "status-format[1]", lineFmt); err != nil {
		return err
	}
	return tm.SetOption(sess, domain.StatusKey, "1")
}

func statusOff(tm tmux.Manager, sess string) error {
	if !statusIsOn(tm, sess) {
		return nil
	}
	statusVal := tm.ShowOption(sess, domain.StatusSavedKey)
	format0 := tm.ShowOption(sess, domain.StatusFormat0Saved)
	format1 := tm.ShowOption(sess, domain.StatusFormat1Saved)

	if statusVal != "" {
		if err := tm.Run("set-option", "-t", sess, "status", statusVal); err != nil {
			return err
		}
	} else {
		_ = tm.Run("set-option", "-t", sess, "-u", "status")
	}

	if format1 != "" {
		if err := tm.Run("set-option", "-t", sess, "status-format[1]", format1); err != nil {
			return err
		}
	} else {
		_ = tm.Run("set-option", "-t", sess, "-u", "status-format[1]")
	}

	if format0 != "" {
		if err := tm.Run("set-option", "-t", sess, "status-format[0]", format0); err != nil {
			return err
		}
	} else {
		_ = tm.Run("set-option", "-t", sess, "-u", "status-format[0]")
	}

	tm.UnsetOption(sess, domain.StatusKey)
	tm.UnsetOption(sess, domain.StatusSavedKey)
	tm.UnsetOption(sess, domain.StatusFormat0Saved)
	tm.UnsetOption(sess, domain.StatusFormat1Saved)
	return nil
}

func sessionID(tm tmux.Manager, sess string) string {
	id := tm.ShowOption(sess, domain.IDKey)
	if id != "" {
		return id
	}
	id = util.GenID(sess)
	_ = tm.SetOption(sess, domain.IDKey, id)
	return id
}

func resolveSession(tm tmux.Manager, token string) (string, error) {
	if tm.HasSession(token) {
		return token, nil
	}
	sessions, err := tm.ListSessions()
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "", Die(fmt.Sprintf("no such session or id: %s", token))
	}
	match := ""
	count := 0
	for _, sess := range sessions {
		if !isOwlxSession(tm, sess) {
			continue
		}
		if sessionID(tm, sess) == token {
			match = sess
			count++
		}
	}
	if count == 1 {
		return match, nil
	}
	if count > 1 {
		return "", Die(fmt.Sprintf("ambiguous id: %s", token))
	}
	return "", Die(fmt.Sprintf("no such session or id: %s", token))
}

func isOwlxSession(tm tmux.Manager, sess string) bool {
	return tm.ShowOption(sess, domain.FlagKey) == "1"
}
