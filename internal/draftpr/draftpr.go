package draftpr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	DraftBranch = "automation/draft-cards"
	BaseBranch  = "main"
)

func Run(root string, args ...string) (string, string, int, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			return "", string(out), 1, err
		}
	}
	return string(out), "", code, nil
}

func RunStdout(root string, args ...string) (string, int, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = root
	out, err := cmd.Output()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
			return string(out) + string(ee.Stderr), code, nil
		}
		return "", 1, err
	}
	return string(out), code, nil
}

func HasDraftChanges(root string) (bool, error) {
	out, code, err := RunStdout(root, "git", "status", "--porcelain", "draft/")
	if err != nil {
		return false, err
	}
	if code != 0 {
		return false, fmt.Errorf("git status failed (%d): %s", code, out)
	}
	return strings.TrimSpace(out) != "", nil
}

func BranchExistsOnRemote(root, branch string) bool {
	out, _, _ := RunStdout(root, "git", "ls-remote", "--heads", "origin", branch)
	return strings.TrimSpace(out) != ""
}

func CheckoutDraftBranch(root string) error {
	if _, _, err := must(root, "git", "fetch", "origin", BaseBranch); err != nil {
		return err
	}
	if BranchExistsOnRemote(root, DraftBranch) {
		if _, _, err := must(root, "git", "checkout", DraftBranch); err != nil {
			return err
		}
		if _, _, err := must(root, "git", "pull", "--rebase", "origin", DraftBranch); err != nil {
			return err
		}
		_, _, err := must(root, "git", "rebase", "origin/"+BaseBranch)
		return err
	}
	_, _, err := must(root, "git", "checkout", "-B", DraftBranch, "origin/"+BaseBranch)
	return err
}

func must(root string, args ...string) (string, int, error) {
	out, _, code, err := Run(root, args...)
	if err != nil {
		return out, code, err
	}
	if code != 0 {
		return out, code, fmt.Errorf("%s failed (%d):\n%s", strings.Join(args, " "), code, out)
	}
	return out, 0, nil
}

func OpenPR(root, message, bodyExtra string) (string, error) {
	out, _, _ := RunStdout(root, "gh", "pr", "list",
		"--head", DraftBranch,
		"--base", BaseBranch,
		"--state", "open",
		"--json", "url,number",
	)
	if strings.TrimSpace(out) != "" {
		var data []struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(out), &data); err == nil && len(data) > 0 {
			fmt.Printf("Open PR already exists: %s\n", data[0].URL)
			return data[0].URL, nil
		}
	}

	body := fmt.Sprintf(`Automated draft flashcards for review.

%s

## Review

- Check `+"`draft/{lang}/{Topic}/`"+` for new cards
- Move approved files to `+"`queue/{lang}/{Topic}/`"+`
- Merge when ready; publish workflow moves queue → cards

---
%s
`, bodyExtra, message)

	out, code, err := RunStdout(root, "gh", "pr", "create",
		"--base", BaseBranch,
		"--head", DraftBranch,
		"--title", "Draft cards (automation)",
		"--body", body,
	)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("gh pr create failed (%d): %s", code, out)
	}
	url := strings.TrimSpace(out)
	fmt.Printf("Created PR: %s\n", url)
	return url, nil
}

func ListOpenPR(root string) {
	cmd := exec.Command("gh", "pr", "list", "--head", DraftBranch, "--base", BaseBranch, "--state", "open")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
