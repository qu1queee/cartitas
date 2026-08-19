package queuepr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	QueueBranch = "automation/queue-cards"
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

func HasQueueChanges(root string) (bool, error) {
	out, code, err := RunStdout(root, "git", "status", "--porcelain", "queue/")
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

func CheckoutQueueBranch(root string) error {
	if _, _, err := must(root, "git", "fetch", "origin", BaseBranch); err != nil {
		return err
	}
	if BranchExistsOnRemote(root, QueueBranch) {
		if _, _, err := must(root, "git", "checkout", QueueBranch); err != nil {
			return err
		}
		if _, _, err := must(root, "git", "pull", "--rebase", "origin", QueueBranch); err != nil {
			return err
		}
		_, _, err := must(root, "git", "rebase", "origin/"+BaseBranch)
		return err
	}
	_, _, err := must(root, "git", "checkout", "-B", QueueBranch, "origin/"+BaseBranch)
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

type openPR struct {
	URL     string `json:"url"`
	Number  int    `json:"number"`
	IsDraft bool   `json:"isDraft"`
}

func OpenPR(root, message, bodyExtra string) (string, error) {
	if pr, ok := findOpenPR(root); ok {
		if err := EnsureReady(root, pr); err != nil {
			return pr.URL, err
		}
		fmt.Printf("Open PR already exists: %s\n", pr.URL)
		return pr.URL, nil
	}

	body := fmt.Sprintf(`Automated queued flashcards.

%s

## After merge

Publish workflow moves `+"`queue/{lang}/{Topic}/`"+` → `+"`cards/{lang}/{Topic}/`"+`.

---
%s
`, bodyExtra, message)

	// Never pass --draft: this PR must be ready for review.
	out, code, err := RunStdout(root, "gh", "pr", "create",
		"--base", BaseBranch,
		"--head", QueueBranch,
		"--title", "Queue cards (automation)",
		"--body", body,
	)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("gh pr create failed (%d): %s", code, out)
	}
	url := strings.TrimSpace(out)
	pr := openPR{URL: url}
	if err := EnsureReady(root, pr); err != nil {
		return url, err
	}
	fmt.Printf("Created PR: %s\n", url)
	return url, nil
}

func findOpenPR(root string) (openPR, bool) {
	out, _, _ := RunStdout(root, "gh", "pr", "list",
		"--head", QueueBranch,
		"--base", BaseBranch,
		"--state", "open",
		"--json", "url,number,isDraft",
	)
	if strings.TrimSpace(out) == "" {
		return openPR{}, false
	}
	var data []openPR
	if err := json.Unmarshal([]byte(out), &data); err != nil || len(data) == 0 {
		return openPR{}, false
	}
	return data[0], true
}

func EnsureReady(root string, pr openPR) error {
	ref := pr.URL
	if ref == "" && pr.Number > 0 {
		ref = fmt.Sprintf("%d", pr.Number)
	}
	if ref == "" {
		ref = QueueBranch
	}
	view, _, _ := RunStdout(root, "gh", "pr", "view", ref, "--json", "url,number,isDraft")
	var current openPR
	if err := json.Unmarshal([]byte(strings.TrimSpace(view)), &current); err == nil {
		pr = current
	}
	if !pr.IsDraft {
		return nil
	}
	out, code, err := RunStdout(root, "gh", "pr", "ready", ref)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("gh pr ready failed (%d): %s", code, out)
	}
	fmt.Printf("Marked PR ready for review: %s\n", pr.URL)
	return nil
}

func ListOpenPR(root string) {
	cmd := exec.Command("gh", "pr", "list", "--head", QueueBranch, "--base", BaseBranch, "--state", "open")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
