import {readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
import test from 'node:test';
import {parseContract, sourceView} from './api-contract.mjs';

const contract = parseContract(await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8'));
const serverCodegen = parseContract(await readFile('../services/tadoku-api/spec/server-codegen.yaml', 'utf8'));
const callbackServerCodegen = parseContract(await readFile('../services/tadoku-api/spec/callback-server-codegen.yaml', 'utf8'));

test('all retained wire contracts survive the merge, including internal callers', async () => {
  let operations = 0;
  let publicOperations = 0;
  const ids = new Set();
  for (const [name, source] of Object.entries(contract['x-tadoku-sources'])) {
    const view = sourceView(contract, name);
    if (!source.path) {
      let sourceOperations = 0;
      for (const item of Object.values(contract.paths)) {
        for (const operation of Object.values(item)) {
          if (operation['x-tadoku-source'] !== name) continue;
          sourceOperations++;
          assert.equal(operation['x-tadoku-owner'], 'native');
        }
      }
      assert.ok(sourceOperations > 0, `${name} has no operations`);
      if (source.exposure === 'public') assert.doesNotMatch(JSON.stringify(view), /\/internal\/v1\/|:8080/);
      continue;
    }

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
      if (contract.paths['/content/pages/{namespace}/{slug}'].put['x-tadoku-owner'] === 'native') {
        legacy.paths['/pages/{namespace}/{id}'].put.responses['409'] = {description: 'Page already exists'};
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
    if (name === 'Immersion') {
      const permissionFailure = {
        description: 'Permission check failed',
        content: {
          'application/json': {
            schema: {
              type: 'object',
              required: ['message'],
              properties: {
                message: {type: 'string'},
              },
            },
          },
        },
      };
      if (contract.paths['/immersion/contests/create-permissions'].get['x-tadoku-owner'] === 'native') {
        legacy.paths['/contests/create-permissions'].get.responses['404'] = {description: 'User identity not found'};
        legacy.paths['/contests/create-permissions'].get.responses['500'] = permissionFailure;
      }
      if (contract.paths['/immersion/contests'].post['x-tadoku-owner'] === 'native') {
        legacy.paths['/contests'].post.responses['500'] = {
          ...permissionFailure,
          description: 'Invalid authenticated identity',
        };
      }
      if (contract.paths['/immersion/contests/{id}/registration'].get['x-tadoku-owner'] === 'native') {
        delete legacy.paths['/contests/{id}/registration'].get.responses['404'];
        legacy.paths['/contests/{id}/registration'].get.responses['204'] = {description: 'registration does not exist'};
        legacy.paths['/contests/{id}/registration'].get.responses['500'] = {
          ...permissionFailure,
          description: 'Invalid authenticated identity',
        };
      }
      if (contract.paths['/immersion/contests/{id}/registration'].post['x-tadoku-owner'] === 'native') {
        legacy.paths['/contests/{id}/registration'].post.requestBody.required = true;
        legacy.paths['/contests/{id}/registration'].post.responses['500'] = {
          ...permissionFailure,
          description: 'Invalid authenticated identity or persistence failure',
        };
      }
      if (contract.paths['/immersion/contests/ongoing-registrations'].get['x-tadoku-owner'] === 'native') {
        legacy.paths['/contests/ongoing-registrations'].get.responses['500'] = {
          ...permissionFailure,
          description: 'Invalid authenticated identity',
        };
      }
      // Native profile reads document the empty 500 already returned by legacy.
      for (const path of [
        '/users/{userId}/profile',
        '/users/{userId}/activity/{year}',
        '/users/{userId}/scores/{year}',
        '/users/{userId}/activity-split/{year}',
      ]) {
        if (contract.paths[`/immersion${path}`].get['x-tadoku-owner'] === 'native') {
          legacy.paths[path].get.responses['500'] = {description: 'profile read failed'};
        }
      }
    }
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
  assert.equal(operations, 76);
  assert.equal(publicOperations, 71);
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
  const generated = [
    ...serverCodegen['output-options']['include-operation-ids'],
    ...callbackServerCodegen['output-options']['include-operation-ids'],
  ];
  assert.equal(new Set(generated).size, generated.length, 'generated operations must belong to one server');
  assert.deepEqual(owned.sort(), generated.sort());
});
