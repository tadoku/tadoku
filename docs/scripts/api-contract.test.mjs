import {readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
import test from 'node:test';
import {parseContract, sourceView} from './api-contract.mjs';

const contract = parseContract(await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8'));
const serverCodegen = parseContract(await readFile('../services/tadoku-api/spec/server-codegen.yaml', 'utf8'));

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
    if (name === 'Content') {
      legacy.paths['/announcements/{namespace}'].post.responses['409'] = {description: 'Announcement already exists'};
      if (contract.paths['/content/pages/{namespace}'].post['x-tadoku-owner'] === 'native') {
        legacy.paths['/pages/{namespace}'].post.responses['409'] = {description: 'Page ID or slug already exists'};
      }
      if (contract.paths['/content/posts/{namespace}'].post['x-tadoku-owner'] === 'native') {
        legacy.paths['/posts/{namespace}'].post.responses['409'] = {description: 'Post ID or slug already exists'};
      }
      if (contract.paths['/content/posts/{namespace}/{slug}'].put['x-tadoku-owner'] === 'native') {
        legacy.paths['/posts/{namespace}/{id}'].put.responses['409'] = {description: 'Post slug already exists'};
      }
      legacy.components.schemas.Announcement.properties.href.maxLength = 2048;
      legacy.components.schemas.AnnouncementList.allOf[1].properties.announcements.maxItems = 100;
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

test('native-owned operations match server generation', () => {
  const owned = [];
  for (const item of Object.values(contract.paths)) {
    for (const operation of Object.values(item)) {
      assert.ok(['native', 'legacy'].includes(operation['x-tadoku-owner']));
      assert.ok(['public', 'internal', 'callback'].includes(operation['x-tadoku-exposure']));
      if (operation['x-tadoku-owner'] === 'native') owned.push(operation.operationId);
    }
  }
  assert.deepEqual(owned.sort(), [...serverCodegen['output-options']['include-operation-ids']].sort());
});
