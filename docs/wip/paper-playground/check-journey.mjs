import assert from 'node:assert/strict'
import {writeFile} from 'node:fs/promises'
import {openBrowser} from '/tmp/paper-browser/browser.mjs'
const browser=await openBrowser();const page=await browser.newPage({viewport:{width:1440,height:1000}});const errors=[];page.on('pageerror',e=>errors.push(e.message));const steps=[]
try{
 await page.goto('http://127.0.0.1:33946/');await page.locator('h1').waitFor()
 await page.getByRole('link',{name:'Join Round 5',exact:true}).click();await page.getByRole('link',{name:'Sign in to join',exact:true}).click();await page.getByRole('button',{name:'Continue as Anton'}).click()
 await page.getByRole('heading',{name:'Join 2026 Round 5'}).waitFor();steps.push('Guest sign-in returns to intended contest registration')
 await page.getByRole('button',{name:'Save registration'}).click();await page.getByRole('heading',{name:'2026 Round 5',exact:true}).waitFor();steps.push('Registration saves and returns to the selected contest')
 await page.locator('main').getByRole('link',{name:'Log activity',exact:true}).click();await page.getByRole('heading',{name:'Log activity',exact:true}).waitFor()
 await page.getByLabel('What did you read or listen to?').fill('Night reading')
 await page.getByLabel('Pages read',{exact:false}).fill('12');await page.getByLabel('Time spent reading').fill('20')
 await page.getByRole('button',{name:'Save activity',exact:true}).click();await page.getByRole('heading',{name:'Night reading',exact:true}).waitFor();const logPath=new URL(page.url()).pathname;steps.push('New activity saves an entity record and selected contest contribution')
 await page.getByRole('link',{name:'Edit log',exact:true}).click();await page.getByLabel('Pages read',{exact:false}).fill('14');await page.getByRole('button',{name:'Save changes',exact:true}).click();await page.getByRole('heading',{name:'Night reading',exact:true}).waitFor();steps.push('Editing returns to the same log ID');assert.equal(new URL(page.url()).pathname,logPath)
 await page.goto('http://127.0.0.1:33946/users/anton/activity');await page.getByRole('link',{name:'Night reading',exact:true}).waitFor();steps.push('Saved activity appears in profile history')
 await page.goto('http://127.0.0.1:33946/leaderboard/latest');await page.getByRole('button',{name:'Find my row'}).click();await page.waitForFunction(()=>document.activeElement?.id==='standing-anton');assert.match(await page.locator('#standing-anton').innerText(),/614/);steps.push('Leaderboard applies local score delta; Find my row pages and focuses the participant')
 await page.getByLabel('Find a participant').fill('Kai');assert.match(await page.locator('main tbody').innerText(),/1\s+Kai/);steps.push('Name search preserves the full-board rank')
 await page.evaluate(()=>window.scrollTo(0,document.body.scrollHeight));await page.locator('.app-header').getByRole('link',{name:'Contests',exact:true}).click();await page.waitForFunction(()=>scrollY===0);steps.push('Changing destination resets document scroll')
 await page.goto('http://127.0.0.1:33946'+logPath);await page.locator('h1').waitFor();await page.getByText('More actions',{exact:true}).click();await page.getByRole('button',{name:'Delete log',exact:true}).click();await page.getByRole('dialog').waitFor();await page.keyboard.press('Escape');await page.getByRole('dialog').waitFor({state:'hidden'});assert.equal(await page.getByRole('heading',{name:'Night reading'}).count(),1);steps.push('Delete confirmation cancels with Escape')
 await page.screenshot({path:'/var/lib/t3/worktrees/tadoku/t3code-a6ef6616/docs/wip/paper-playground/evidence/shared/saved-log-desktop.png',fullPage:true})
 assert.equal(errors.length,0);console.log(JSON.stringify({steps,errors,logPath}));await writeFile('/var/lib/t3/worktrees/tadoku/t3code-a6ef6616/docs/wip/paper-playground/journey-browser.json',JSON.stringify({steps,errors,logPath},null,2))
}catch(e){console.log(JSON.stringify({steps,errors,url:page.url(),body:(await page.locator('body').innerText()).slice(-1800)}));throw e}finally{await browser.close()}
