#!/usr/bin/env python3
"""Fail when a production config contract item no longer appears where the code consumes it.

Each entry in contract.json names an item, a repository file and the text that must appear in it,
outside comments. It must not be followed by a name character (a letter, digit, `_`, `$` or `-`). Text
that starts with a name character must not be preceded by one or a `.`, and text that ends with one must
not be followed by a `.` that starts another name: `release-webv3` does not match `release-webv3-access`,
`HOST=` does not match `APP_HOST=` and `tadoku.app` does not match `tadoku.app.example`. Write entries
whose text ends in punctuation, such as a template literal, with their closing delimiter. Matching is textual, so an equivalent rewrite of the consumer, such as
`process.env['X']` for `process.env.X`, needs its contract text updated too.
"""

import argparse
import json
from pathlib import Path
import re
import sys

NAME = r'[\w$-]'


# ponytail: a quote-aware scan, not a parser; a JavaScript regex literal holding a quote or comment opener can
# confuse it. The clean-tree run catches an entry that trips on it; use a real tokenizer if one does.
def strip_comments(text, quotes, line_comment, block_comments):
    kept, i, quote = [], 0, None
    while i < len(text):
        char = text[i]
        if quote:
            if char == '\\':
                kept.append(text[i:i + 2])
                i += 2
                continue
            if char == quote or (char == '\n' and quote != '`'):
                quote = None
            kept.append(char)
        elif text.startswith(line_comment, i) and (line_comment != '#' or i == 0 or text[i - 1].isspace()):
            newline = text.find('\n', i)
            i = len(text) if newline < 0 else newline
            continue
        elif block_comments and text.startswith('/*', i):
            close = text.find('*/', i + 2)
            i = len(text) if close < 0 else close + 2
            kept.append(' ')
            continue
        else:
            if char in quotes:
                quote = char
            kept.append(char)
        i += 1
    return ''.join(kept)


def without_comments(path):
    text = path.read_text()
    if path.suffix in {'.ts', '.tsx', '.js', '.go'}:
        return strip_comments(text, '\'"`', '//', True)
    if path.suffix in {'.yaml', '.yml'}:
        return strip_comments(text, '\'"', '#', False)
    if path.name == 'Dockerfile':
        return re.sub(r'(?m)^[ \t]*#.*$', '', text)
    return text


def found(text, source):
    before = r'(?<![\w$.-])' if re.match(NAME, text) else ''
    after = rf'(?!{NAME}|\.\w)' if re.search(NAME + '$', text) else rf'(?!{NAME})'
    return re.search(before + re.escape(text) + after, source)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--root', type=Path, default=Path(__file__).resolve().parents[3])
    root = parser.parse_args().root
    entries = json.loads((root / 'tools/ci/configcontract/contract.json').read_text())['entries']

    failures = []
    for entry in entries:
        path = root / entry['file']
        if not path.is_file():
            failures.append(f"{entry['item']}: expected file {entry['file']} does not exist")
        elif not found(entry['text'], without_comments(path)):
            failures.append(f"{entry['item']}: {entry['text']!r} not found in {entry['file']}")

    if failures:
        print('Production config contract broken (see docs/docs/develop/ci-workflows.md):', file=sys.stderr)
        for failure in failures:
            print(f'  {failure}', file=sys.stderr)
        sys.exit(1)
    print(f'Production config contract holds: {len(entries)} entries checked.')


if __name__ == '__main__':
    main()
