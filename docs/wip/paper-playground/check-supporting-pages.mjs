import assert from 'node:assert/strict'
import {writeFile} from 'node:fs/promises'
import {openBrowser} from '/tmp/paper-browser/browser.mjs'
const browser=await openBrowser();const results=[];const errors=[]
try{for(const width of [1440,390])for(const appearance of ['light','dark']){
 const context=await browser.newContext({viewport:{width,height:900}});const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message))
 await page.goto('http://127.0.0.1:33946/?scenario=participant-live');await page.locator('h1').waitFor();await page.goto('http://127.0.0.1:33946/settings');await page.getByLabel('Appearance').selectOption(appearance);await page.getByLabel('Control density').selectOption(appearance==='dark'?'compact':'comfortable');await page.getByRole('button',{name:'Save settings'}).click();await page.waitForFunction(theme=>document.documentElement.dataset.theme===theme,appearance)
 for(const route of ['/settings','/logs/new','/logs/reading-konbini/edit','/contact']){
  await page.goto('http://127.0.0.1:33946'+route);await page.locator('h1').waitFor();await page.evaluate(()=>document.fonts.ready)
  const record=await page.evaluate(()=>({route:location.pathname,h1:document.querySelector('h1').textContent,h1Count:document.querySelectorAll('h1').length,width:innerWidth,scrollWidth:document.documentElement.scrollWidth,theme:document.documentElement.dataset.theme,density:document.documentElement.dataset.density,fields:[...document.querySelectorAll('.paper-input:not(textarea)')].map(e=>({name:e.getAttribute('name'),height:e.getBoundingClientRect().height}))}))
  assert.equal(record.scrollWidth,width,route+' document overflow');assert.equal(record.h1Count,1,route+' heading count');results.push(record)
  if(route==='/logs/new'||route==='/settings')await page.screenshot({path:'/var/lib/t3/worktrees/tadoku/t3code-a6ef6616/docs/wip/paper-playground/evidence/shared/'+route.slice(1).replaceAll('/','-')+'-'+width+'-'+appearance+'.png',fullPage:true})
 }
 await page.getByLabel('Your message').fill('Please add Icelandic to the sample language choices.');await page.getByRole('button',{name:'Prepare message'}).click();await page.getByText('Message prepared',{exact:true}).waitFor();await context.close()
}
assert.equal(errors.length,0);console.log(JSON.stringify({cases:results.length,errors}));await writeFile('/var/lib/t3/worktrees/tadoku/t3code-a6ef6616/docs/wip/paper-playground/supporting-browser.json',JSON.stringify({results,errors},null,2))
}finally{await browser.close()}
