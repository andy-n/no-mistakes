package steps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/config"
	"github.com/kunchenguid/no-mistakes/internal/pipeline"
	"github.com/kunchenguid/no-mistakes/internal/verificationplan"
)

func TestReviewAndTestRefuseLostVerificationPlan(t *testing.T) {
	for _, step := range []pipeline.Step{&ReviewStep{}, &TestStep{}} {
		t.Run(string(step.Name()), func(t *testing.T) {
			dir, base, head := setupGitRepo(t)
			ag := &mockAgent{name: "test"}
			sctx := newTestContextWithDBRecords(t, ag, dir, base, head, config.Commands{})
			root := t.TempDir()
			source := filepath.Join(root, "plan.txt")
			if err := os.WriteFile(source, []byte("verify the output\n"), 0600); err != nil {
				t.Fatal(err)
			}
			plan, err := verificationplan.Capture(filepath.Join(root, "inputs"), source, sctx.Run.RepoID, sctx.Run.Branch, head)
			if err != nil {
				t.Fatal(err)
			}
			sctx.Run.VerificationPlan = plan
			if err := os.Chmod(plan.Path, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(plan.Path); err != nil {
				t.Fatal(err)
			}
			outcome, err := step.Execute(sctx)
			if err == nil || !strings.Contains(err.Error(), "read captured verification plan") || outcome != nil {
				t.Fatalf("lost attachment accepted: %+v %v", outcome, err)
			}
			if len(ag.calls) != 0 {
				t.Fatal("agent invoked without its required captured evidence")
			}
		})
	}
}
