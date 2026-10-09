package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// Skill installers (npx skills and the like) make <folder>/skills for
// every agent they know. A folder holding only that is no sign the agent is
// here; anything the agent writes beside it is (qtwaiter on X: Goose, Crush,
// Command Code, fx, Devin, Hermes and Droid listed, none installed).
func TestSkillsOnlyFolderIsNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, mk := range []func(home, cfg string) *Agent{
		goose, crush, devin,
		func(h, _ string) *Agent { return commandCode(h) },
		func(h, _ string) *Agent { return fx(h) },
		func(h, _ string) *Agent { return hermes(h) },
		func(h, _ string) *Agent { return droid(h) },
	} {
		home := t.TempDir()
		cfg := filepath.Join(home, ".config")
		a := mk(home, cfg)
		t.Run(a.ID, func(t *testing.T) {
			skill := filepath.Join(a.Dir, "skills", "grill", "SKILL.md")
			if err := os.MkdirAll(filepath.Dir(skill), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(skill, []byte("---\nname: grill\n---\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			os.WriteFile(filepath.Join(a.Dir, ".DS_Store"), nil, 0o644)
			if mk(home, cfg).Detected() {
				t.Errorf("a %s holding only skills is taken for %s", a.Dir, a.Name)
			}
			if err := os.WriteFile(filepath.Join(a.Dir, "history.jsonl"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if !mk(home, cfg).Detected() {
				t.Errorf("%s with its own files beside skills isn't found", a.Name)
			}
		})
	}
}
