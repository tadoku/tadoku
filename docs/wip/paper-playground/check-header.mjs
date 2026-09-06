import assert from 'node:assert/strict'
import { openBrowser } from '/tmp/paper-browser/browser.mjs'
const browser=await openBrowser()
try{
 const page=await browser.newPage({viewport:{width:390,height:844}})
 await page.goto('http://127.0.0.1:33946/users/anton?scenario=profile')
 await page.locator('h1').waitFor();await page.evaluate(()=>document.fonts.ready)
 const metrics=await page.locator('.app-header a[href="/logs/new"]').evaluate(a=>{const range=document.createRange();range.selectNodeContents(a);return {lines:new Set([...range.getClientRects()].map(r=>Math.round(r.top))).size,width:a.getBoundingClientRect().width,height:a.getBoundingClientRect().height,scrollWidth:document.documentElement.scrollWidth}})
 console.log(JSON.stringify(metrics))
 assert.equal(metrics.lines,1,'The primary Log activity action must stay on one line at phone width')
 assert.equal(metrics.scrollWidth,390,'Header must fit the phone viewport')
}finally{await browser.close()}
