// Read-only application acceptance; login creates sessions, never accounts/data.
// Run after the documented temporary response markers and isolated seed exist.
import assert from 'node:assert/strict';
const {chromium} = await import(process.env.PLAYWRIGHT_MODULE || 'playwright');
const required = name => { assert(process.env[name], `set ${name}`); return process.env[name]; };
const host = required('PILOT_URL');
const auth = required('AUTH_URL');
const baseOnly = process.argv.includes('--base-only');
const routeA = baseOnly ? '' : required('ROUTE_A');
const routeB = baseOnly ? '' : required('ROUTE_B');
const markerA = baseOnly ? '' : required('MARKER_A');
const markerB = baseOnly ? '' : required('MARKER_B');
const routingOnly = process.argv.includes('--routing-only');
for (const value of [host, auth]) {
  const url = new URL(value);
  assert(url.protocol === 'https:' && url.hostname.endsWith('.dev.lab'), 'development Lab hosts only');
}
const browser = await chromium.launch({headless:true});
try {
  // The browser's temporary profile does not carry the host Lab CA trust store.
  // Verify the endpoint with normal curl CA validation before running this test.
  const a = await browser.newContext({ignoreHTTPSErrors:true});
  const b = await browser.newContext({ignoreHTTPSErrors:true});
  const guest = await browser.newContext({ignoreHTTPSErrors:true});
  const pageA = await a.newPage(), pageB = await b.newPage();
  // Fail as a harness problem if the browser lacks working fonts/input support.
  await pageA.setContent('<input name="control-a"><input name="control-b">');
  await pageA.locator('[name="control-a"]').fill('first');
  await pageA.locator('[name="control-b"]').fill('second');
  assert(await pageA.locator('[name="control-a"]').inputValue() === 'first' && await pageA.locator('[name="control-b"]').inputValue() === 'second', 'browser input control failed; install browser system dependencies/fonts before testing authentication');
  async function login(page, email, password) {
    await page.goto(`${auth}/login`);
    const form = page.locator('form.kratos-form');
    await form.waitFor({state:'visible'});
    await page.waitForLoadState('networkidle');
    const identifier = form.locator('input[name="identifier"]');
    const secret = form.locator('input[name="password"]');
    await identifier.fill(email);
    await secret.fill(password);
    assert(await identifier.inputValue() === email && await secret.inputValue() === password, 'login inputs did not retain values');
    const submitted = page.waitForResponse(r => r.url().includes('/self-service/login') && r.request().method() === 'POST');
    await form.locator('button[type="submit"]').click();
    const loginResponse = await submitted;
    if (loginResponse.status() >= 400) {
      const body = await loginResponse.json();
      // Report only provider error codes, never credentials, cookies or node values.
      const messages = [...(body.ui?.messages || []), ...(body.ui?.nodes || []).flatMap(n=>n.messages || [])];
      throw new Error(`real login rejected: HTTP ${loginResponse.status()}, submitted fields ${Object.keys(loginResponse.request().postDataJSON() || {}).join(',')}, missing ${messages.map(m=>m.context?.property || m.id).join(',')}, error ${body.error?.id || ''}`);
    }
    const response = await page.request.get(`${auth}/kratos/sessions/whoami`);
    assert.equal(response.status(),200,'real Kratos session required');
    return (await response.json()).identity.id;
  }
  async function select(page, route) {
    const leaderboard = page.waitForResponse(r => new URL(r.url()).pathname.includes('/api/internal/immersion/leaderboard/yearly/'));
    // Attach a handler immediately so a failed navigation cannot leave an unhandled timeout.
    leaderboard.catch(() => {});
    await page.goto(`${host}/?dev-branch=${encodeURIComponent(route)}`);
    const endpoint = await page.evaluate(() => window.__NEXT_DATA__.runtimeConfig.apiEndpoint);
    assert.equal(new URL(endpoint, host).origin, new URL(host).origin, 'browser API must stay on the selected host, including base fallback');
    const response = await leaderboard;
    assert.equal(response.status(), 200, 'actual browser leaderboard request');
    const result = await response.json();
    assert(result.entries.length > 0, 'existing development leaderboard fixture must be visible');
    const table = page.getByRole('table').filter({has: page.getByRole('columnheader', {name:'Rank', exact:true})});
    await table.waitFor({state:'visible'});
    assert.equal(await table.locator('tbody tr').count(), result.entries.length, 'leaderboard rows must render from the successful API response');
    assert(!(await page.getByText('Could not retrieve leaderboard, please try again later.').count()));
  }
  const ping = async (context, expected, headers = {}) => {
    const response = await context.request.get(`${host}/api/internal/immersion/ping`, {headers});
    assert.equal(response.status(),200);
    assert.equal(await response.text(),'pong');
    assert.equal(response.headers()['x-tadoku-pilot'] || '',expected);
  };
  if (baseOnly) {
    await select(pageA,'base');
    await ping(guest,'');
    console.log('PASS: base browser leaderboard renders and API is available without overlays');
  } else {
  await select(pageA,routeA);
  await select(pageB,routeB);
  await ping(a,markerA);
  await ping(b,markerB);
  await ping(guest,'');
  await ping(guest,'',{'x-dev-branch':routeA});
  await ping(a,markerA,{'x-dev-branch':routeB});
  assert.equal((await a.cookies(host)).find(c=>c.name==='dev_branch').value,routeA);
  assert.equal((await b.cookies(host)).find(c=>c.name==='dev_branch').value,routeB);
  const isolated = `${host}/api/internal/content/pages/main/dev-cli-pilot`;
  assert.equal((await b.request.get(isolated)).status(),200,'isolated seed must be visible in B');
  assert.equal((await a.request.get(isolated)).status(),404,'shared A database must not receive isolated seed');
  assert.equal((await guest.request.get(isolated)).status(),404,'base database must not receive isolated seed');
  assert.notEqual(await pageB.evaluate(()=>getComputedStyle(document.body).backgroundColor),'rgb(219, 234, 254)','API-only B must fall back to base frontend');
  await select(pageA,routeB);
  await ping(a,markerB);
  await select(pageA,'base');
  await ping(a,'');
  assert.equal((await a.cookies(host)).some(c=>c.name==='dev_branch'),false);
  await ping(b,markerB);
  await select(pageA,'nonexistent-pilot-branch');
  await ping(a,'');
  console.log('PASS: two-owner routing, spoof protection, legacy hop, partial frontend fallback, isolated seed, switch/clear/missing selection');
  }
  if (!routingOnly && !baseOnly) {
    const identityA = await login(pageA, required('ADMIN_EMAIL'), required('ADMIN_PASSWORD'));
    const identityB = await login(pageB, required('READER_EMAIL'), required('READER_PASSWORD'));
    assert.notEqual(identityA,identityB,'use two distinct fixture identities');
    await select(pageA,routeA);
    await select(pageB,routeB);
    const sessionA = await pageA.evaluate(()=>window.__NEXT_DATA__.props.pageProps.session?.identity.id);
    assert(sessionA === identityA,'Next.js SSR must retain the authenticated session');
    const listPath = `${host}/api/internal/content/pages/main`;
    assert.equal((await a.request.get(listPath)).status(),200,'administrator should pass native authorization');
    assert.equal((await b.request.get(listPath)).status(),403,'reader must not pass administrator authorization');
    assert.equal((await guest.request.get(listPath)).status(),401,'guest must not pass authentication');
    console.log('PASS: real login for two users, authenticated SSR and native authorization');
  }
  console.log(JSON.stringify({passed:true,scope:baseOnly?'base-only; authentication NOT tested':routingOnly?'routing-only; authentication NOT tested':'full'},null,2));
} finally { await browser.close(); }
