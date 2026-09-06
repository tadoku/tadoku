import assert from 'node:assert/strict'
import {openBrowser} from '/tmp/paper-browser/browser.mjs'
const browser=await openBrowser()
try{const page=await browser.newPage({viewport:{width:1440,height:1000}})
 await page.goto('http://127.0.0.1:33946/?scenario=participant-live');await page.locator('h1').waitFor();await page.goto('http://127.0.0.1:33946/logs/new');await page.getByLabel('Pages read',{exact:false}).waitFor();await page.evaluate(()=>document.fonts.ready)
 for(const [density,height] of [['comfortable',44],['compact',36]]){
  await page.evaluate(density=>document.documentElement.dataset.density=density,density)
  const fields=await page.locator('input[name="amount"],input[name="date"]').evaluateAll(inputs=>inputs.map(input=>({name:input.name,height:input.getBoundingClientRect().height,top:input.getBoundingClientRect().top,padding:getComputedStyle(input).padding,lineHeight:getComputedStyle(input).lineHeight})))
  console.log(JSON.stringify({density,fields}))
  for(const field of fields){assert.equal(field.height,height,`${field.name} must use the ${density} height alongside a field with a hint`);assert.equal(field.top,fields[0].top,`${field.name} must align with its neighboring control even when only one has a hint`)}
 }
}finally{await browser.close()}
