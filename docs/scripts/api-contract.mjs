import YAML from 'yaml';

export function parseContract(text) {
  return YAML.parse(text);
}

// Build-time views only: the merged Tadoku specification is the sole source.
// Restore old names/paths so existing documentation URLs remain stable.
export function sourceView(contract, sourceName) {
  const source = contract['x-tadoku-sources'][sourceName];
  if (!source) throw new Error(`Unknown contract source: ${sourceName}`);
  const restore = value => {
    if (Array.isArray(value)) return value.map(restore);
    if (!value || typeof value !== 'object') return value;
    return Object.fromEntries(Object.entries(value).filter(([k]) => !k.startsWith('x-tadoku-') && !k.startsWith('x-legacy-')).map(([key, item]) => {
      if (key === '$ref') return [key, item.replace(new RegExp(`^(#/components/[^/]+/)${sourceName}`), '$1')];
      if (key === 'security') return [key, item.map(entry => Object.fromEntries(Object.entries(entry).map(([name, scopes]) => [name.slice(sourceName.length), scopes])))];
      return [key, restore(item)];
    }));
  };
  const view = { openapi: contract.openapi, info: source.info, ...(source.externalDocs ? {externalDocs: source.externalDocs} : {}), servers: source.servers, ...(source.tags ? {tags: source.tags} : {}), paths: {}, components: {} };
  for (const [path, item] of Object.entries(contract.paths)) {
    for (const [method, operation] of Object.entries(item)) {
      if (operation['x-tadoku-source'] !== sourceName) continue;
      const originalPath = operation['x-legacy-path'];
      const restored = restore(operation);
      restored.operationId = operation['x-legacy-operation-id'];
      const names = [...path.matchAll(/\{([^}]+)\}/g)].map(match => match[1]);
      const originalNames = [...originalPath.matchAll(/\{([^}]+)\}/g)].map(match => match[1]);
      for (const param of restored.parameters ?? []) {
        if (param.in === 'path') param.name = originalNames[names.indexOf(param.name)];
      }
      view.paths[originalPath] ??= {};
      view.paths[originalPath][method] = restored;
    }
  }
  const prefixes = Object.keys(contract['x-tadoku-sources']).sort((a, b) => b.length - a.length);
  for (const [kind, components] of Object.entries(contract.components)) {
    const entries = Object.entries(components).filter(([name]) => prefixes.find(prefix => name.startsWith(prefix)) === sourceName);
    if (entries.length) view.components[kind] = Object.fromEntries(entries.map(([name, value]) => [name.slice(sourceName.length), restore(value)]));
  }
  return view;
}
