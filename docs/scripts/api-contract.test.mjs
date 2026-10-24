import {readFile} from 'node:fs/promises';
import assert from 'node:assert/strict';
import test from 'node:test';
import YAML from 'yaml';
import {parseContract, sourceView} from './api-contract.mjs';

const contractText = await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8');
const contract = parseContract(contractText);
const serverCodegen = parseContract(await readFile('../services/tadoku-api/spec/server-codegen.yaml', 'utf8'));
const callbackServerCodegen = parseContract(await readFile('../services/tadoku-api/spec/callback-server-codegen.yaml', 'utf8'));

test('bundled contract has unique YAML anchors', () => {
  const anchors = new Set();
  YAML.visit(YAML.parseDocument(contractText), {
    Node(_, node) {
      if (!node.anchor) return;
      assert.ok(!anchors.has(node.anchor), `duplicate YAML anchor: ${node.anchor}`);
      anchors.add(node.anchor);
    },
  });
});

test('public views cover canonical operations', () => {
  let operations = 0;
  let publicOperations = 0;
  const ids = new Set();
  for (const [name, source] of Object.entries(contract['x-tadoku-sources'])) {
    const view = sourceView(contract, name);
    const sourceOperations = Object.values(contract.paths).flatMap(item => Object.values(item))
      .filter(operation => operation['x-tadoku-source'] === name);
    const viewOperations = Object.values(view.paths).flatMap(item => Object.values(item));
    assert.ok(sourceOperations.length > 0, `${name} has no operations`);
    assert.deepEqual(
      viewOperations.map(operation => operation.operationId).sort(),
      sourceOperations.map(operation => operation['x-legacy-operation-id']).sort(),
      `${name} operations`,
    );
    if (source.exposure === 'public') assert.doesNotMatch(JSON.stringify(view), /\/internal\/v1\/|:8080/);
  }
  for (const item of Object.values(contract.paths)) {
    for (const operation of Object.values(item)) {
      assert.ok(!ids.has(operation.operationId), operation.operationId);
      ids.add(operation.operationId);
      operations++;
      if (operation['x-tadoku-exposure'] === 'public') publicOperations++;
    }
  }
  assert.equal(operations, 83);
  assert.equal(publicOperations, 81);
});

test('operations match server generation', () => {
  const owned = [];
  for (const item of Object.values(contract.paths)) {
    for (const operation of Object.values(item)) {
      assert.ok(['public', 'internal', 'callback'].includes(operation['x-tadoku-exposure']));
      owned.push(operation.operationId);
    }
  }
  const generated = [
    ...serverCodegen['output-options']['include-operation-ids'],
    ...callbackServerCodegen['output-options']['include-operation-ids'],
  ];
  assert.equal(new Set(generated).size, generated.length, 'generated operations must belong to one server');
  assert.deepEqual(owned.sort(), generated.sort());
});
