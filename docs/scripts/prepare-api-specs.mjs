import {readFile, mkdir, writeFile} from 'node:fs/promises';
import YAML from 'yaml';
import {parseContract, publicViews} from './api-contract.mjs';

const contract = parseContract(await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8'));
await mkdir('.generated/api', {recursive: true});
for (const {name, domain, view} of publicViews(contract)) {
  const text = YAML.stringify(view);
  if (text.includes('/internal/v1/') || text.includes(':8080')) throw new Error(`Internal surface in public view: ${name}`);
  await writeFile(`.generated/api/${domain}.yaml`, text);
}
