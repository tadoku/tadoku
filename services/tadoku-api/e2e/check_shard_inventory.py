#!/usr/bin/env python3
"""Compare an unfiltered unsharded Go test XML with every shard's XML."""

import argparse
from collections import Counter
import json
from pathlib import Path
import sys
import xml.etree.ElementTree as ET


def inventory(paths):
    cases = Counter()
    for path in paths:
        root = ET.parse(path).getroot()
        found = root.findall('.//testcase')
        if not found:
            raise ValueError(f'{path}: empty test inventory')
        for case in found:
            key = (case.get('classname', ''), case.attrib['name'])
            if any(case.find(tag) is not None for tag in ('skipped', 'failure', 'error')):
                raise ValueError(f'{path}: unsuccessful case {key}')
            cases[key] += 1
    duplicates = {str(key): count for key, count in cases.items() if count != 1}
    if duplicates:
        raise ValueError(f'duplicated cases: {duplicates}')
    return cases


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--baseline', required=True, type=Path)
    parser.add_argument('--shards', required=True, nargs='+', type=Path)
    args = parser.parse_args()
    try:
        baseline = inventory([args.baseline])
        sharded = inventory(args.shards)
        missing = sorted(baseline.keys() - sharded.keys())
        unexpected = sorted(sharded.keys() - baseline.keys())
        if missing or unexpected:
            raise ValueError(f'missing cases: {missing}; unexpected cases: {unexpected}')
    except (ValueError, KeyError, OSError, ET.ParseError) as error:
        print(error, file=sys.stderr)
        return 1
    print(json.dumps({'cases': len(baseline), 'shards': len(args.shards), 'exactly_once': True}))
    return 0


if __name__ == '__main__':
    sys.exit(main())
