package state

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	sessionsDirName = "sessions"
	statusDirName   = "status"
	layoutDirName   = "layouts"
)

// Session is the persisted metadata for an owlx session.
type Session struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Layout      string    `json:"layout"`
	Repo        string    `json:"repo"`
	Branch      string    `json:"branch"`
	Category    string    `json:"category"`
	Intent      string    `json:"intent"`
	Worktree    string    `json:"worktree"`
	RepoDir     string    `json:"repo_dir"`
	WorktreeDir string    `json:"worktree_dir"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store manages persisted session metadata.
type Store struct {
	Dir string
}

func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

func (s *Store) Ensure() error {
	if s.Dir == "" {
		return errors.New("state dir is empty")
	}
	if err := os.MkdirAll(s.sessionsDir(), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(s.statusDir(), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(s.layoutsDir(), 0o755); err != nil {
		return err
	}
	return nil
}

func (s *Store) Save(session Session) (Session, error) {
	if session.Name == "" {
		return Session{}, errors.New("session name is empty")
	}
	if session.ID == "" {
		session.ID = GenID(session.Name)
	}
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	if err := s.Ensure(); err != nil {
		return Session{}, err
	}
	payload, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return Session{}, fmt.Errorf("encode session: %w", err)
	}
	path := s.sessionPath(session.ID)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return Session{}, fmt.Errorf("write session: %w", err)
	}
	return session, nil
}

func (s *Store) LoadByID(id string) (Session, error) {
	path := s.sessionPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("decode session: %w", err)
	}
	return session, nil
}

func (s *Store) FindByToken(token string) (Session, error) {
	if token == "" {
		return Session{}, errors.New("token is empty")
	}
	if session, err := s.LoadByID(token); err == nil {
		return session, nil
	}

	sessions, err := s.List()
	if err != nil {
		return Session{}, err
	}
	for _, session := range sessions {
		if session.Name == token {
			return session, nil
		}
	}
	return Session{}, fmt.Errorf("no session for token: %s", token)
}

func (s *Store) List() ([]Session, error) {
	entries, err := os.ReadDir(s.sessionsDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Session{}, nil
		}
		return nil, err
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		session, err := s.LoadByID(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
	})
	return sessions, nil
}

func (s *Store) Delete(id string) error {
	if id == "" {
		return errors.New("id is empty")
	}
	path := s.sessionPath(id)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.removeAuxFiles(id)
		}
		return err
	}
	return s.removeAuxFiles(id)
}

func GenID(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])[:6]
}

func (s *Store) sessionsDir() string {
	return filepath.Join(s.Dir, sessionsDirName)
}

func (s *Store) sessionPath(id string) string {
	return filepath.Join(s.sessionsDir(), fmt.Sprintf("%s.json", id))
}

func (s *Store) StatusPath(id string) string {
	return filepath.Join(s.statusDir(), fmt.Sprintf("%s.txt", id))
}

func (s *Store) LayoutPath(id string) string {
	return filepath.Join(s.layoutsDir(), fmt.Sprintf("%s.kdl", id))
}

func (s *Store) statusDir() string {
	return filepath.Join(s.Dir, statusDirName)
}

func (s *Store) layoutsDir() string {
	return filepath.Join(s.Dir, layoutDirName)
}

func (s *Store) removeAuxFiles(id string) error {
	statusPath := s.StatusPath(id)
	layoutPath := s.LayoutPath(id)
	_ = os.Remove(statusPath)
	_ = os.Remove(layoutPath)
	return nil
}
