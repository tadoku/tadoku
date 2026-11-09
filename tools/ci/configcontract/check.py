#!/usr/bin/env python3
"""Fail when a production config contract item no longer appears where the code consumes it.

Each entry in contract.json names an item, a repository file and the text that must appear in it.
The text must not be followed by a letter, digit, `_` or `-`, so `release-webv3` does not match
`release-webv3-access` and `TADOKU_API_URL` does not match `TADOKU_API_URL_V2`.
"""

import argparse
import json
from pathlib import Path
import re
import sys


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
        elif not re.search(re.escape(entry['text']) + r'(?![\w-])', path.read_text()):
            failures.append(f"{entry['item']}: {entry['text']!r} not found in {entry['file']}")

    if failures:
        print('Production config contract broken (see docs/docs/develop/ci-workflows.md):', file=sys.stderr)
        for failure in failures:
            print(f'  {failure}', file=sys.stderr)
        sys.exit(1)
    print(f'Production config contract holds: {len(entries)} entries checked.')


if __name__ == '__main__':
    main()
