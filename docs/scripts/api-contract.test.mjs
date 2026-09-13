import {readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
import test from 'node:test';
import {parseContract, sourceView} from './api-contract.mjs';

const contract = parseContract(await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8'));

test('all retained wire contracts survive the merge, including internal callers', async () => {
  let operations = 0;
  let publicOperations = 0;
  const ids = new Set();
  for (const [name, source] of Object.entries(contract['x-tadoku-sources'])) {
    const legacy = parseContract(await readFile(`../${source.path}`, 'utf8'));
    // Path-level parameters and security may be expressed on each operation.
    for (const item of Object.values(legacy.paths)) {
      const shared = item.parameters;
      delete item.parameters;
      for (const operation of Object.values(item)) {
        if (shared) operation.parameters = [...shared.filter(p => !(operation.parameters ?? []).some(q => p.name === q.name && p.in === q.in)), ...(operation.parameters ?? [])];
        if (legacy.security && !operation.security) operation.security = legacy.security;
      }
    }
    const view = sourceView(contract, name);
    assert.deepEqual(view.paths, legacy.paths, `${name} paths`);
    assert.deepEqual(view.components, legacy.components ?? {}, `${name} components`);
    if (source.exposure === 'public') {
      assert.doesNotMatch(JSON.stringify(view), /\/internal\/v1\/|:8080/);
    }
  }
  for (const item of Object.values(contract.paths)) {
    for (const operation of Object.values(item)) {
      assert.ok(!ids.has(operation.operationId), operation.operationId);
      ids.add(operation.operationId);
      operations++;
      if (operation['x-tadoku-exposure'] === 'public') publicOperations++;
    }
  }
  assert.equal(operations, 81);
  assert.equal(publicOperations, 72);
});

test('only the active-announcement read is owned by native code', () => {
  const owned = [];
  for (const [path, item] of Object.entries(contract.paths)) {
    for (const [method, operation] of Object.entries(item)) {
      assert.ok(['native', 'legacy'].includes(operation['x-tadoku-owner']));
      assert.ok(['public', 'internal', 'callback'].includes(operation['x-tadoku-exposure']));
      if (operation['x-tadoku-owner'] === 'native') owned.push(`${method} ${path}`);
    }
  }
  assert.deepEqual(owned, ['get /content/announcements/{namespace}/active']);
});
