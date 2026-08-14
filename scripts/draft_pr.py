#!/usr/bin/env python3
"""Commit draft/ changes and ensure one open PR exists (automation/draft-cards)."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DRAFT_BRANCH = "automation/draft-cards"
BASE_BRANCH = "main"


def run(cmd: list[str], *, check: bool = True, capture: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        cmd,
        cwd=ROOT,
        check=check,
        capture_output=capture,
        text=True,
    )


def has_draft_changes() -> bool:
    result = run(["git", "status", "--porcelain", "draft/"], capture=True)
    return bool(result.stdout.strip())


def branch_exists_on_remote(branch: str) -> bool:
    result = run(["git", "ls-remote", "--heads", "origin", branch], capture=True, check=False)
    return bool(result.stdout.strip())


def checkout_draft_branch() -> None:
    run(["git", "fetch", "origin", BASE_BRANCH])
    if branch_exists_on_remote(DRAFT_BRANCH):
        run(["git", "checkout", DRAFT_BRANCH])
        run(["git", "pull", "--rebase", "origin", DRAFT_BRANCH])
        run(["git", "rebase", f"origin/{BASE_BRANCH}"])
    else:
        run(["git", "checkout", "-B", DRAFT_BRANCH, f"origin/{BASE_BRANCH}"])


def open_pr(message: str, body_extra: str) -> str:
    existing = run(
        [
            "gh", "pr", "list",
            "--head", DRAFT_BRANCH,
            "--base", BASE_BRANCH,
            "--state", "open",
            "--json", "url,number",
        ],
        capture=True,
        check=False,
    )
    if existing.returncode == 0 and existing.stdout.strip():
        data = json.loads(existing.stdout)
        if data:
            url = data[0]["url"]
            print(f"Open PR already exists: {url}")
            return url

    title = "Draft cards (automation)"
    body = f"""Automated draft flashcards for review.

{body_extra}

## Review

- Check `draft/{{lang}}/{{Topic}}/` for new cards
- Move approved files to `queue/{{lang}}/{{Topic}}/`
- Merge when ready; publish workflow moves queue → cards

---
{message}
"""
    result = run(
        [
            "gh", "pr", "create",
            "--base", BASE_BRANCH,
            "--head", DRAFT_BRANCH,
            "--title", title,
            "--body", body,
        ],
        capture=True,
    )
    url = result.stdout.strip()
    print(f"Created PR: {url}")
    return url


def main() -> int:
    parser = argparse.ArgumentParser(description="Push draft/ to automation PR branch.")
    parser.add_argument("--message", required=True, help="Commit message")
    parser.add_argument(
        "--pr-note",
        default="",
        help="Extra line for PR body (e.g. lang/topic summary)",
    )
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    if not has_draft_changes():
        print("No changes under draft/; nothing to commit.")
        # Still report open PR if any
        run(
            ["gh", "pr", "list", "--head", DRAFT_BRANCH, "--base", BASE_BRANCH, "--state", "open"],
            check=False,
        )
        return 0

    if args.dry_run:
        run(["git", "status", "--short", "draft/"])
        print(f"Would commit on branch {DRAFT_BRANCH} and update open PR.")
        return 0

    checkout_draft_branch()
    run(["git", "add", "draft/"])
    run(["git", "commit", "-m", args.message])
    run(["git", "push", "-u", "origin", DRAFT_BRANCH])
    open_pr(args.message, args.pr_note)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
