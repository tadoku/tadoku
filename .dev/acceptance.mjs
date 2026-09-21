// Read-only application acceptance; login creates sessions, never accounts/data.
// Run after the documented temporary response markers and isolated seed exist.
import assert from 'node:assert/strict';
const {chromium} = await import(process.env.PLAYWRIGHT_MODULE || 'playwright');
const required = name => { assert(process.env[name], `set ${name}`); return process.env[name]; };
const host = required('PILOT_URL');
const auth = required('AUTH_URL');
const routeA = required('ROUTE_A');
const routeB = required('ROUTE_B');
const markerA = required('MARKER_A');
const markerB = required('MARKER_B');
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
  async function login(page, email, password) {
    await page.goto(`${auth}/login`);
    const form = page.locator('form.kratos-form');
    await form.waitFor({state:'visible'});
    await page.waitForLoadState('networkidle');
    const identifier = form.locator('input[name="identifier"]');
    const secret = form.locator('input[name="password"]');
    let identifierRetained = false, passwordRetained = false;
    for (let attempt = 0; attempt < 2 && !(identifierRetained && passwordRetained); attempt++) {
      await secret.fill('');
      await secret.pressSequentially(password);
      await identifier.fill('');
      await identifier.pressSequentially(email);
      await page.waitForTimeout(250);
      identifierRetained = (await identifier.inputValue()) === email;
      passwordRetained = (await secret.inputValue()) === password;
    }
    assert(identifierRetained && passwordRetained, `login inputs did not retain values: identifier=${identifierRetained}, password=${passwordRetained}`);
    const submitted = page.waitForResponse(r => r.url().includes('/self-service/login') && r.request().method() === 'POST');
    await form.getByRole('button', {name:'Log in', exact:true}).click();
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
  const select = (page,route) => page.goto(`${host}/?dev-branch=${route}`);
  const ping = async (context, expected, headers = {}) => {
    const response = await context.request.get(`${host}/api/internal/immersion/ping`, {headers});
    assert.equal(response.status(),200);
    assert.equal(await response.text(),'pong');
    assert.equal(response.headers()['x-tadoku-pilot'] || '',expected);
  };
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
  if (!routingOnly) {
    const identityA = await login(pageA, required('ADMIN_EMAIL'), required('ADMIN_PASSWORD'));
    const identityB = await login(pageB, required('READER_EMAIL'), required('READER_PASSWORD'));
    assert.notEqual(identityA,identityB,'use two distinct fixture identities');
    await select(pageA,routeA);
    await select(pageB,routeB);
    const sessionA = await pageA.evaluate(()=>window.__NEXT_DATA__.props.pageProps.session?.identity.id);
    assert.equal(sessionA,identityA,'Next.js SSR must retain the authenticated session');
    const listPath = `${host}/api/internal/content/pages/main`;
    assert.equal((await a.request.get(listPath)).status(),200,'administrator should pass native authorization');
    assert.equal((await b.request.get(listPath)).status(),403,'reader must not pass administrator authorization');
    assert.equal((await guest.request.get(listPath)).status(),401,'guest must not pass authentication');
    console.log('PASS: real login for two users, authenticated SSR and native authorization');
  }
  console.log(JSON.stringify({passed:true,scope:routingOnly?'routing-only; authentication NOT tested':'full'},null,2));
} finally { await browser.close(); }
