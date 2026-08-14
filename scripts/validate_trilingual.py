#!/usr/bin/env python3
"""Ensure draft card sets exist in en, es, and de with matching filenames."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent


def load_languages(path: Path) -> list[str]:
    with path.open(encoding="utf-8") as f:
        cfg = yaml.safe_load(f)
    return list(cfg.get("languages", {}).keys())


def relative_key(path: Path) -> str:
    """e.g. Animals/ocean_early.md (lang-agnostic)."""
    parts = path.parts
    # draft/{lang}/{Topic}/file.md
    if len(parts) < 4 or parts[0] != "draft":
        return str(path)
    return str(Path(*parts[2:]))


def find_draft_files(base: Path) -> dict[str, set[str]]:
    """Map relative key -> set of language codes present."""
    groups: dict[str, set[str]] = {}
    if not base.is_dir():
        return groups
    for lang_dir in base.iterdir():
        if not lang_dir.is_dir():
            continue
        lang = lang_dir.name
        for md in lang_dir.rglob("*.md"):
            rel = md.relative_to(base)
            key = str(Path(*rel.parts[1:]))  # drop lang segment
            groups.setdefault(key, set()).add(lang)
    return groups


def validate_draft(base: Path, required_langs: list[str]) -> list[str]:
    errors: list[str] = []
    groups = find_draft_files(base)
    for key, present in sorted(groups.items()):
        missing = [lang for lang in required_langs if lang not in present]
        if missing:
            have = ", ".join(sorted(present))
            need = ", ".join(missing)
            errors.append(
                f"draft/{key}: has [{have}], missing [{need}] — "
                f"every new set needs all languages: {', '.join(required_langs)}"
            )
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate trilingual draft sets.")
    parser.add_argument(
        "--languages-config",
        type=Path,
        default=ROOT / "languages.yaml",
    )
    parser.add_argument(
        "--only",
        help="Check only this relative path, e.g. Animals/ocean_early.md",
    )
    args = parser.parse_args()

    required = load_languages(args.languages_config)
    errors = validate_draft(ROOT / "draft", required)

    if args.only:
        errors = [e for e in errors if args.only in e]

    if errors:
        for err in errors:
            print(err, file=sys.stderr)
        print(
            f"\nFix: add matching files under draft/{{en,es,de}}/ for each incomplete set.",
            file=sys.stderr,
        )
        return 1

    print(f"All draft sets complete for: {', '.join(required)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
