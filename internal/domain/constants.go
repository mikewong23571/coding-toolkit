package domain

const (
	FlagKey            = "@owlx"
	LayoutKey          = "@owlx_layout"
	RepoKey            = "@owlx_repo"
	RepoDirKey         = "@owlx_repo_dir"
	WorktreeKey        = "@owlx_worktree"
	WorktreeDirKey     = "@owlx_worktree_dir"
	BranchKey          = "@owlx_branch"
	CatKey             = "@owlx_cat"
	IntentKey          = "@owlx_intent"
	IDKey              = "@owlx_id"
	BaseBranchKey      = "@owlx_base_branch"
	CreatedTsKey       = "@owlx_created_ts"
	StatusKey          = "@owlx_status_on"
	StatusSavedKey     = "@owlx_status_saved"
	StatusFormat0Saved = "@owlx_status_format0_saved"
	StatusFormat1Saved = "@owlx_status_format1_saved"
)

var (
	ValidLayouts = []string{"main", "fork", "playground", "archive"}
	ValidCats    = []string{"fix", "feat", "refactor", "research", "chore"}
)

func IsValidLayout(layout string) bool {
	for _, l := range ValidLayouts {
		if l == layout {
			return true
		}
	}
	return false
}

func IsValidCategory(cat string) bool {
	for _, c := range ValidCats {
		if c == cat {
			return true
		}
	}
	return false
}
