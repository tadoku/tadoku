// Startup proof for the published frontend images with external networking disabled.
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')

function proveFrontends({ evidence, prefix, containers, report, run }, docs) {
  const frontends = docs.filter(d => d.kind === 'Deployment' && d.metadata.name.startsWith('frontend-'))
  for (const deployment of frontends) {
    const spec = deployment.spec.template.spec
    const container = spec.containers[0]
    const native = deployment.metadata.name === 'frontend-webv3'
    const cached = spawnSync('docker', ['image', 'inspect', container.image], { stdio: 'ignore', timeout: 30000 })
    if (cached.status !== 0) run('docker', ['pull', container.image])
    report.images[container.image] = JSON.parse(run('docker', ['image', 'inspect', container.image, '--format', '{{json .RepoDigests}}']))
    const name = `${prefix}-${deployment.metadata.name}`
    const volume = spec.volumes.find(v => v.name === (native ? 'ca' : 'startup'))
    const mount = container.volumeMounts.find(m => m.name === volume.name)
    const config = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === deployment.metadata.namespace && d.metadata.name === volume.configMap.name)
    const dir = path.join(evidence, deployment.metadata.name)
    fs.mkdirSync(dir)
    for (const [file, content] of Object.entries(config.data)) fs.writeFileSync(path.join(dir, file), content)
    // No external network: this checks the published server/asset startup, not
    // login or API acceptance, and cannot accidentally contact production.
    const runtime = native ? docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === deployment.metadata.namespace && d.metadata.name === container.envFrom[0].configMapRef.name) : undefined
    const env = native ? Object.entries(runtime.data) : container.env.map(e => [e.name, e.value])
    run('docker', ['create', '--name', name, '--network', 'none', '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '512m', ...env.flatMap(([key, value]) => ['-e', `${key}=${value}`]), ...(native ? ['--read-only', '--cap-drop=ALL', '--security-opt=no-new-privileges', '--volume', mount.mountPath] : ['--entrypoint', container.command[0]]), container.image, ...(container.command?.slice(1) || [])])
    containers.push(name)
    run('docker', ['cp', native ? `${dir}/.` : dir, `${name}:${mount.mountPath}`])
    run('docker', ['start', name])
    try {
      if (native) {
        run('docker', ['exec', name, 'node', '-e', String.raw`
          (async () => {
            if (process.getuid() !== 1000 || process.getgid() !== 1000) throw new Error('Webv3 image lost its non-root identity');
            if (!require('node:fs').readFileSync(process.env.NODE_EXTRA_CA_CERTS, 'utf8').includes('BEGIN CERTIFICATE')) throw new Error('Lab CA is not mounted');
            const base = 'http://127.0.0.1:3000';
            const readHtml = (path, host) => new Promise((resolve, reject) => {
              const request = require('node:http').get(base + path, {headers: {Host: host}, timeout: 15000}, response => {
                let html = '';
                response.setEncoding('utf8');
                response.on('data', chunk => { html += chunk });
                response.on('end', () => resolve({status: response.statusCode, html}));
                response.on('error', reject);
              });
              request.on('timeout', () => request.destroy(new Error('Webv3 SSR timed out for ' + host)));
              request.on('error', reject);
            });
            let health;
            for (let i = 0; i < 60; i++) {
              try { health = await fetch(base + '/healthz', {signal: AbortSignal.timeout(2000)}); if (health.ok) break; } catch {}
              await new Promise(resolve => setTimeout(resolve, 500));
            }
            if (!health?.ok || await health.text() !== 'ok\n') throw new Error('Native Nitro server did not become healthy');
            for (const [host, product] of [['next.tadoku.dev.lab', 'tadoku.dev.lab'], ['proof-0123abcd.next.tadoku.dev.lab', 'proof-0123abcd.tadoku.dev.lab']]) {
              const response = await readHtml('/plan', host);
              if (response.status !== 200) throw new Error('Webv3 SSR failed for ' + host);
              const html = response.html;
              if (!html.includes('Daily goal') || !html.includes('href="https://' + product + '/logs/new"') || /href="https:\/\/tadoku\.app/.test(html)) throw new Error('Webv3 SSR development product links are incorrect for ' + host);
              const assets = [...html.matchAll(/(?:src|href)="([^" ]+\.(?:js|css)(?:\?[^" ]*)?)"/g)].map(m => m[1]);
              if (!assets.length) throw new Error('Webv3 SSR has no compiled assets');
              for (const asset of assets) if (!(await fetch(new URL(asset, base), {signal: AbortSignal.timeout(5000)})).ok) throw new Error('Missing compiled Webv3 asset: ' + asset);
              console.log(JSON.stringify({host, product, assets: assets.length}));
            }
            const outage = await readHtml('/', 'next.tadoku.dev.lab');
            if (!outage.html.includes('Contest updates are delayed') || !outage.html.includes('The board could not be loaded')) throw new Error('API outage must remain visible on the homepage');
            const after = await fetch(base + '/healthz', {signal: AbortSignal.timeout(2000)});
            if (after.status !== 200) throw new Error('API outage broke Webv3 health');
            console.log(JSON.stringify({api: process.env.TADOKU_API_URL, outage: 'visible', health: after.status}));
          })().catch(error => { console.error(error); process.exit(1) });
        `])
      } else run('docker', ['exec', name, 'node', '-e', String.raw`
        (async () => {
          let response;
          for (let i = 0; i < 60; i++) {
            try {
              response = await fetch('http://127.0.0.1:3000/', {signal: AbortSignal.timeout(5000)});
              if (response.ok) break;
            } catch {}
            await new Promise(resolve => setTimeout(resolve, 1000));
          }
          if (!response?.ok) throw new Error('Frontend did not become ready');
          const html = await response.text();
          const match = html.match(/<script id="__NEXT_DATA__"[^>]*>(.*?)<\/script>/s);
          if (!match) throw new Error('No Next runtime data in rendered page');
          const runtime = JSON.parse(match[1]).runtimeConfig;
          if (runtime?.homeUrl !== 'https://tadoku.dev.lab' || runtime?.kratosPublicEndpoint !== 'https://account.tadoku.dev.lab/kratos' || runtime?.apiEndpoint !== 'https://tadoku.dev.lab/api/internal' || runtime?.cookieDomain !== '.tadoku.dev.lab' || runtime?.cookieSecure !== true) throw new Error('Development runtime URLs/cookies were not applied');
          const assets = [...html.matchAll(/<script[^>]*src="([^\"]+)"/g)].map(m => m[1]).filter(p => p.startsWith('/_next/'));
          if (!assets.length) throw new Error('No compiled Next assets');
          for (const asset of assets) {
            const result = await fetch('http://127.0.0.1:3000' + asset);
            if (!result.ok) throw new Error('Missing compiled asset: ' + asset);
            await result.arrayBuffer();
          }
          console.log(JSON.stringify({runtime, assets: assets.length}));
        })().catch(error => { console.error(error); process.exit(1) });
      `])
    } finally { run('docker', ['logs', name]) }
    run('docker', ['rm', '-f', '-v', name])
    containers.pop()
  }
}

module.exports = { proveFrontends }
