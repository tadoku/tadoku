// Frontend image startup proof with external networking disabled. Legacy SSR
// API hosts resolve to loopback fixtures inside each container: the private
// Oathkeeper host answers sessions, the public HTTPS host only counts connections.
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')
const Y = require('../../../../frontend/node_modules/yaml')

function proveFrontends({ evidence, prefix, containers, report, run }, docs) {
  const rules = Y.parse(docs.find(d => d.kind === 'ConfigMap' && d.data?.['rules.yaml']).data['rules.yaml'])
  const privateSession = new URL(rules.find(rule => rule.id === 'tadoku:identity:session:internal').match.url)
  const privateRules = rules.filter(rule => rule.match.url.startsWith(`${privateSession.origin}/`))
    .map(rule => ({ url: rule.match.url, methods: rule.match.methods }))
  const publicApi = 'https://tadoku.dev.lab/api/internal'
  const branch = 'proof-0123abcd'
  const apiHosts = [privateSession.hostname, new URL(publicApi).hostname, `${branch}.${new URL(publicApi).hostname}`]
  const frontends = docs.filter(d => d.kind === 'Deployment' && d.metadata.name.startsWith('frontend-'))
  for (const deployment of frontends) {
    const spec = deployment.spec.template.spec
    const container = spec.containers[0]
    const image = deployment.metadata.name === 'frontend-authv2' && process.env.TADOKU_BASE_AUTHV2_IMAGE || container.image
    if (image !== container.image) report.unverified.push('authv2 uses a locally built image; publication and shared deployment are unverified')
    const native = ['frontend-webv3', 'frontend-authv2'].includes(deployment.metadata.name)
    const cached = spawnSync('docker', ['image', 'inspect', image], { stdio: 'ignore', timeout: 30000 })
    if (cached.status !== 0) run('docker', ['pull', image])
    report.images[image] = JSON.parse(run('docker', ['image', 'inspect', image, '--format', '{{json .RepoDigests}}']))
    const name = `${prefix}-${deployment.metadata.name}`
    const volume = spec.volumes.find(v => v.name === (native ? 'ca' : 'startup'))
    const mount = container.volumeMounts.find(m => m.name === volume.name)
    const config = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === deployment.metadata.namespace && d.metadata.name === volume.configMap.name)
    const dir = path.join(evidence, deployment.metadata.name)
    fs.mkdirSync(dir)
    for (const [file, content] of Object.entries(config.data)) fs.writeFileSync(path.join(dir, file), content)
    // No external network: this checks server/asset startup, not
    // login or API acceptance, and cannot accidentally contact production.
    const runtime = native ? docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === deployment.metadata.namespace && d.metadata.name === container.envFrom[0].configMapRef.name) : undefined
    const env = native ? Object.entries(runtime.data) : container.env.map(e => [e.name, e.value])
    run('docker', ['create', '--name', name, '--network', 'none', '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '512m', ...env.flatMap(([key, value]) => ['-e', `${key}=${value}`]), ...(native ? ['--read-only', '--cap-drop=ALL', '--security-opt=no-new-privileges', '--volume', mount.mountPath] : ['--entrypoint', container.command[0], ...apiHosts.flatMap(host => ['--add-host', `${host}:127.0.0.1`])]), image, ...(container.command?.slice(1) || [])])
    containers.push(name)
    run('docker', ['cp', native ? `${dir}/.` : dir, `${name}:${mount.mountPath}`])
    run('docker', ['start', name])
    try {
      if (deployment.metadata.name === 'frontend-authv2') {
        run('docker', ['exec', name, 'node', '-e', String.raw`
          (async () => {
            if (process.getuid() !== 1000 || process.getgid() !== 1000) throw new Error('authv2 must run as its published non-root identity');
            if (!require('node:fs').readFileSync(process.env.NODE_EXTRA_CA_CERTS, 'utf8').includes('BEGIN CERTIFICATE')) throw new Error('Lab CA is not mounted');
            const base = 'http://127.0.0.1:3000';
            let health;
            for (let i = 0; i < 60; i++) {
              try { health = await fetch(base + '/healthz', {signal: AbortSignal.timeout(2000)}); if (health.ok) break; } catch {}
              await new Promise(resolve => setTimeout(resolve, 500));
            }
            if (!health?.ok) throw new Error('authv2 Nitro server did not become healthy');
            for (const host of ['account.tadoku.dev.lab', 'proof-0123abcd.account.tadoku.dev.lab']) {
              const response = await new Promise((resolve, reject) => {
                require('node:http').get(base + '/login', {headers: {Host: host}}, response => {
                  let html = '';
                  response.on('data', chunk => { html += chunk });
                  response.on('end', () => resolve({status: response.statusCode, html}));
                  response.on('error', reject);
                }).on('error', reject);
              });
              if (response.status !== 200 || !response.html.includes('We can’t reach your account right now') || !response.html.includes('href="/login"')) throw new Error('authv2 must offer a retry during API outage');
              const web = host.replace('.account.', '.').replace(/^account\./, '');
              if (!response.html.includes('href="https://' + web + '"')) throw new Error('authv2 home link lost its branch');
              const assets = [...response.html.matchAll(/(?:src|href)="([^" ]+\.(?:js|css)(?:\?[^" ]*)?)"/g)].map(m => m[1]);
              if (!assets.length) throw new Error('authv2 has no compiled assets');
              for (const asset of assets) if (!(await fetch(new URL(asset, base))).ok) throw new Error('Missing authv2 asset: ' + asset);
            }
            if (!(await fetch(base + '/healthz')).ok) throw new Error('API outage broke authv2 health');
            console.log(JSON.stringify({app: 'authv2', outage: 'visible', health: 200}));
          })().catch(error => { console.error(error); process.exit(1) });
        `])
      } else if (native) {
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
      } else run('docker', ['exec', '-e', `PROOF=${JSON.stringify({ app: deployment.metadata.name, privateSession, privateRules, publicApi, branch })}`, name, 'node', '-e', String.raw`
        (async () => {
          const http = require('node:http');
          const proof = JSON.parse(process.env.PROOF);
          const privateSession = new URL(proof.privateSession);
          const cookie = 'ory_kratos_session=base-proof; tadoku_session_' + proof.branch + '=branch-proof; unrelated=proof';
          const account = {id: '7d0c6a3e-3f43-4f4b-9a7e-1c2b3d4e5f60', email: 'reader@example.test', display_name: 'Private SSR proof', created_at: '2026-01-01T00:00:00Z', authenticated_at: '2026-01-01T00:00:00Z'};
          let observed = [];
          let publicConnections = 0;
          const listen = (server, port) => new Promise((resolve, reject) => server.once('error', reject).listen(Number(port), '127.0.0.1', resolve).unref());
          await listen(http.createServer((request, response) => {
            observed.push({method: request.method, url: request.url, cookie: request.headers.cookie, branch: request.headers['x-dev-branch']});
            response.setHeader('Content-Type', 'application/json');
            if (request.url !== '/api/internal/identity/session') { response.statusCode = 404; response.end('{}'); return }
            response.end(JSON.stringify({account_app: 'v1', account: request.headers.cookie ? account : null}));
          }), privateSession.port);
          await listen(require('node:net').createServer(socket => { publicConnections++; socket.destroy() }), 443);
          const readHtml = (host, headers) => new Promise((resolve, reject) => {
            const request = http.get('http://127.0.0.1:3000/', {headers: {Host: host, Cookie: cookie, ...headers}, timeout: 15000}, response => {
              let html = '';
              response.setEncoding('utf8');
              response.on('data', chunk => { html += chunk });
              response.on('end', () => resolve({status: response.statusCode, html}));
              response.on('error', reject);
            });
            request.on('timeout', () => request.destroy(new Error('SSR timed out for ' + host)));
            request.on('error', reject);
          });
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
          const host = new URL(runtime[proof.app === 'frontend-admin' ? 'adminUrl' : 'homeUrl']).host;
          const branchApi = proof.publicApi.replace('://', '://' + proof.branch + '.');
          for (const [requestHost, headers, browserApi] of [[host, {}, proof.publicApi], [proof.branch + '.' + host, {'X-Dev-Branch': proof.branch}, branchApi]]) {
            observed = [];
            publicConnections = 0;
            const {status, html} = await readHtml(requestHost, headers);
            const props = JSON.parse(html.match(/<script id="__NEXT_DATA__"[^>]*>(.*?)<\/script>/s)?.[1] ?? '{}').props?.pageProps;
            console.log(JSON.stringify({host: requestHost, status, publicConnections, observed}));
            if (publicConnections || !observed.some(row => row.method === 'GET' && row.url === '/api/internal/identity/session')) throw new Error(requestHost + ' SSR session read must use the private Oathkeeper session rule, not the public API; public connections: ' + publicConnections + ', private requests: ' + observed.length);
            const unmatched = observed.filter(row => !proof.privateRules.some(rule => rule.url === privateSession.origin + row.url && rule.methods.includes(row.method)));
            if (unmatched.length) throw new Error(requestHost + ' SSR requests have no private Oathkeeper rule: ' + JSON.stringify(unmatched));
            if (!observed.every(row => row.cookie === cookie && row.branch === headers['X-Dev-Branch'])) throw new Error(requestHost + ' SSR must forward the exact Cookie and only the incoming X-Dev-Branch');
            if (status !== 200 || !props || props.sessionUnavailable || props.session?.id !== account.id) throw new Error(requestHost + ' SSR did not render the private session response');
            if (props.appUrls?.apiEndpoint !== browserApi || html.includes(privateSession.hostname)) throw new Error(requestHost + ' browser must receive only the public API endpoint');
          }
        })().catch(error => { console.error(error); process.exit(1) });
      `])
    } finally { run('docker', ['logs', name]) }
    run('docker', ['rm', '-f', '-v', name])
    containers.pop()
  }
}

module.exports = { proveFrontends }
