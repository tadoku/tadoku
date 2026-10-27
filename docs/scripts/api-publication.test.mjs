import {readdir, readFile} from 'node:fs/promises';
import {basename, dirname, join} from 'node:path';
import {inflateSync} from 'node:zlib';
import assert from 'node:assert/strict';
import test from 'node:test';
import YAML from 'yaml';
import {parseContract, publicViews} from './api-contract.mjs';

// Run after `pnpm build`. Each generated page is identified by what it renders: an operation page by the
// operation in its compressed `api` front matter, an introduction by the view whose operation pages link to it
// through `info_path`. Tag and schema pages only group or describe a view's content, so inside a view's directory
// only the build is checked for them.
const contract = parseContract(await readFile('../services/tadoku-api/spec/openapi.yaml', 'utf8'));
const pages = await Promise.all((await readdir('docs/api', {recursive: true}))
  .filter(file => /\.mdx?$/.test(file) && file !== 'index.md')
  .map(async file => {
    const frontMatter = YAML.parse((await readFile(join('docs/api', file), 'utf8')).split(/^---$/m)[1]);
    const api = frontMatter.api && JSON.parse(inflateSync(Buffer.from(frontMatter.api, 'base64')));
    return {file, id: frontMatter.id, infoPath: frontMatter.info_path, api};
  }));
const viewTitles = new Map(pages.filter(page => page.api).map(page => [page.infoPath, page.api.info.title]));

function renders(page) {
  if (page.api) return `${page.api.info.title} ${page.api.method.toUpperCase()} ${page.api.path}`;
  if (!page.file.endsWith('.info.mdx')) return page.file;
  return `${viewTitles.get(`api/${dirname(page.file)}/${page.id}`) ?? page.file} introduction`;
}

function tagOrSchemaView(file) {
  if (file.endsWith('.tag.mdx')) return dirname(file);
  if (file.endsWith('.schema.mdx') && basename(dirname(file)) === 'schemas') return dirname(dirname(file));
}

function assertSamePages(actual, expected) {
  const missing = [...expected];
  const extra = actual.filter(page => {
    const index = missing.indexOf(page);
    if (index === -1) return true;
    missing.splice(index, 1);
    return false;
  });
  assert.deepEqual({missing, extra}, {missing: [], extra: []});
}

test('reference has one page per public operation and an introduction per public view', () => {
  const expected = publicViews(contract).flatMap(({view}) => [
    `${view.info.title} introduction`,
    ...Object.entries(view.paths).flatMap(([path, item]) =>
      Object.keys(item).map(method => `${view.info.title} ${method.toUpperCase()} ${path}`)),
  ]);
  const viewDirectories = new Set(pages.filter(page => page.api).map(page => dirname(page.file)));
  const compared = pages.filter(page => page.api || !viewDirectories.has(tagOrSchemaView(page.file)));
  assertSamePages(compared.map(renders), expected);
});

test('build publishes every reference page and no other API page', async () => {
  const built = (await readdir('build/api', {recursive: true})).filter(file => basename(file) === 'index.html');
  const expected = ['index.html', ...pages.map(page => join(dirname(page.file), page.id, 'index.html'))];
  assertSamePages(built, expected);
});
