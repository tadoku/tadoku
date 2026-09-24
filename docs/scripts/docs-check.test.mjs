import {execFileSync} from 'node:child_process';
import {existsSync, readFileSync} from 'node:fs';
import {dirname, join} from 'node:path';
import assert from 'node:assert/strict';
import test from 'node:test';

const repo = execFileSync('git', ['rev-parse', '--show-toplevel'], {encoding: 'utf8'}).trim();
const markdown = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '*.md', '*.mdx'], {cwd: repo, encoding: 'utf8'})
  .split('\n')
  .filter(file => file && existsSync(join(repo, file)));
const generatedApi = /^docs\/docs\/api\/(immersion|content|profile|authorization)\//;
const pages = markdown.filter(file => file.startsWith('docs/docs/') && !generatedApi.test(file));
const repoPath = /^(services|frontend|docs|k8s|infra|jobs|scripts|tools|\.dev|\.agents|\.github)\/[\w.\-/]*$/;

const read = file => readFileSync(join(repo, file), 'utf8');
const ignored = path => {
  try {
    execFileSync('git', ['check-ignore', '-q', path], {cwd: repo});
    return true;
  } catch {
    return false;
  }
};
const withoutCodeBlocks = text => text.replace(/^(```|~~~)[\s\S]*?^\1/gm, '');

test('every docs page has a description', () => {
  const missing = pages.filter(file => {
    const frontmatter = read(file).match(/^---\n([\s\S]*?)\n---\n/);
    return !frontmatter || !/^description:[ \t]*\S/m.test(frontmatter[1]);
  });
  assert.deepEqual(missing, [], 'pages need a frontmatter description');
});

test('inline repository paths exist', () => {
  // Ignored local files may be mentioned. ADRs record past decisions and may name code that no longer exists.
  const files = markdown.filter(file => !generatedApi.test(file) && !file.startsWith('docs/docs/adr/'));
  const missing = [];
  for (const file of files) {
    for (const [, code] of withoutCodeBlocks(read(file)).matchAll(/`([^`\s]+)`/g)) {
      const path = code.replace(/^\.\//, '');
      if (repoPath.test(path) && !existsSync(join(repo, path)) && !ignored(path)) missing.push(`${file}: ${code}`);
    }
  }
  assert.deepEqual(missing, [], 'inline repository paths must exist');
});

test('relative links outside the docs site resolve', () => {
  // Docusaurus validates links between docs pages; this covers AGENTS.md, READMEs and skills.
  const missing = [];
  for (const file of markdown.filter(file => !file.startsWith('docs/docs/'))) {
    for (const [, target] of withoutCodeBlocks(read(file)).matchAll(/\]\(([^)\s]+)\)/g)) {
      if (/^([a-z]+:|#|\/)/.test(target)) continue;
      if (!existsSync(join(repo, dirname(file), decodeURI(target.split('#')[0])))) missing.push(`${file}: ${target}`);
    }
  }
  assert.deepEqual(missing, [], 'relative links must resolve');
});
