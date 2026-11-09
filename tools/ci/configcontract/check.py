#!/usr/bin/env python3
"""Fail when a production config contract item no longer appears where the code consumes it.

Each entry in contract.json names an item, a repository file and the text that must appear in it.
It must not be followed by a name character (a letter, digit, `_`, `$` or `-`). Text that starts with a name
character must not be preceded by one or a `.`, and text that ends with one must not be followed by a `.`
that starts another name: `release-webv3` does not match `release-webv3-access`, `HOST=` does not match
`APP_HOST=` and `tadoku.app` does not match `tadoku.app.example`. Write entries whose text ends in
punctuation, such as a template literal, with their closing delimiter.

Matching searches the raw file text, comments and strings included. It catches a missing file and an
ordinary rename; it does not prove the consumer is still active, so text left behind in a comment still
satisfies an entry. An equivalent rewrite of the consumer, such as `process.env['X']` for `process.env.X`,
needs its contract text updated too.
"""

import argparse
import json
from pathlib import Path
import re
import sys

NAME = r'[\w$-]'


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
        elif not found(entry['text'], path.read_text()):
            failures.append(f"{entry['item']}: {entry['text']!r} not found in {entry['file']}")

    if failures:
        print('Production config contract broken (see docs/docs/develop/ci-workflows.md):', file=sys.stderr)
        for failure in failures:
            print(f'  {failure}', file=sys.stderr)
        sys.exit(1)
    print(f'Production config contract holds: {len(entries)} entries checked.')


if __name__ == '__main__':
    main()
