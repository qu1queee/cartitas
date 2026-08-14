#!/usr/bin/env python3
"""Validate cartitas card files for basic kid-content rules."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

CARD_TYPES = re.compile(r"^(Q:|A:|C:)", re.MULTILINE)
MAX_ANSWER_LEN = 500  # soft limit for kid answers


def validate_file(path: Path) -> list[str]:
    errors: list[str] = []
    text = path.read_text(encoding="utf-8")
    rel = path.relative_to(ROOT)

    if not text.strip():
        errors.append(f"{rel}: file is empty")
        return errors

    blocks = [b.strip() for b in re.split(r"^---\s*$", text, flags=re.MULTILINE) if b.strip()]
    for i, block in enumerate(blocks, start=1):
        has_q = any(line.startswith("Q:") for line in block.splitlines())
        has_a = any(line.startswith("A:") for line in block.splitlines())
        has_c = any(line.startswith("C:") for line in block.splitlines())

        if has_c and (has_q or has_a):
            errors.append(f"{rel} card {i}: cloze (C:) mixed with Q/A in same block")
        elif has_q != has_a:
            errors.append(f"{rel} card {i}: Q/A pair incomplete")
        elif not has_c and not has_q:
            # Allow section-only blocks at file start (headers before first card)
            if i == 1 and block.startswith("#"):
                continue
            errors.append(f"{rel} card {i}: no Q/A or C: card found")

        for line in block.splitlines():
            if line.startswith("A:") and len(line) > MAX_ANSWER_LEN:
                errors.append(f"{rel} card {i}: answer very long ({len(line)} chars)")

    return errors


def main() -> int:
    dirs = [ROOT / "cards", ROOT / "queue", ROOT / "draft"]
    all_errors: list[str] = []

    for base in dirs:
        if not base.is_dir():
            continue
        for path in sorted(base.rglob("*.md")):
            all_errors.extend(validate_file(path))

    if all_errors:
        for err in all_errors:
            print(err, file=sys.stderr)
        return 1

    print("All card files passed basic validation.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
