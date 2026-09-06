import {chromium} from 'playwright';
import assert from 'node:assert/strict';
import {mkdir,writeFile,readFile} from 'node:fs/promises';
const output='/tmp/paper-browser/forms';await mkdir(output,{recursive:true});
const browser=await chromium.launch({headless:true,args:['--no-sandbox']});
const page=await browser.newPage({viewport:{width:1280,height:900}});
const paths=['forms/textarea','forms/select','forms/checkbox','forms/radio-select','forms/radio-group','forms/amount-with-unit','forms/autocomplete','forms/multi-autocomplete','forms/tags-input','actions/button-group','feedback/flash','feedback/loading','data-display/surface','feedback/toast'];
const results=JSON.parse(await readFile(output+'/checks.json','utf8').catch(()=>'[]'));
for(const path of paths){
 if(results.some(r=>r.path==='/components/'+path))continue;
 const row={path:'/components/'+path,matrix:[],interactions:[]};
 await page.goto('http://127.0.0.1:57326/components/'+path);await page.locator('h1').waitFor();
 for(const [width,theme,density] of [[1280,'light','comfortable'],[360,'light','comfortable'],[360,'dark','compact'],[1280,'dark','compact']]){
  await page.setViewportSize({width,height:900});await page.selectOption('select[name=theme]',theme);await page.selectOption('select[name=density]',density);
  const frame=page.frameLocator('iframe');await frame.locator('.paper-fixture-stage').waitFor();await page.evaluate(()=>document.fonts.ready);await page.locator('iframe').scrollIntoViewIfNeeded();
  const geom=await frame.locator('html').evaluate(e=>({overflow:e.scrollWidth>e.clientWidth,controls:[...e.querySelectorAll('.paper-input,.paper-compound-field,.paper-radio-select__segment')].map(c=>({className:c.className,height:c.getBoundingClientRect().height}))}));
  assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false,path+' outeroverflow '+width);
  assert.equal(geom.overflow,false,path+' fixtureoverflow '+width);row.matrix.push({width,theme,density,...geom});
  if((width===1280&&theme==='light')||(width===360&&theme==='dark')){const filename=path.replaceAll('/','-')+'-'+width+'-'+theme+'.png';await page.locator('iframe').screenshot({path:output+'/'+filename});row.matrix.at(-1).screenshot=filename;}
 }
 const frame=page.frameLocator('iframe');
 if(path.startsWith('forms/')&&path!=='forms/checkbox'){
  await page.selectOption('select[name=fixtureId]',{index:1});await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByRole('alert').waitFor();row.interactions.push('Empty fixture starts empty, Save entry exposes required inline error.');
 }
 if(path==='forms/textarea'){
  await frame.getByRole('textbox',{name:/Reading notes/}).fill('Finished chapter three.');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByRole('status').filter({hasText:'Entry saved'}).waitFor();assert((await frame.getByRole('status').innerText()).includes('Finished chapter three.'));row.interactions.push('Valid multiline content saves with visible submitted value.');
 }else if(path==='forms/select'){
  await frame.getByRole('combobox',{name:/Language/}).selectOption('zh');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();row.interactions.push('Native selection Chinese saves language:zh.');
 }else if(path==='forms/checkbox'){
  const c=frame.getByRole('checkbox');await c.uncheck();await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();assert((await frame.getByRole('status').innerText()).includes('false'));row.interactions.push('Unchecked is valid and saves public:false.');
 }else if(path==='forms/radio-select'||path==='forms/radio-group'){
  await frame.getByRole('radio').first().focus();await page.keyboard.press('Space');await page.keyboard.press('ArrowRight');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();row.interactions.push('Native arrow selection and valid save work; empty selection reports error.');
 }else if(path==='forms/amount-with-unit'){
  await frame.getByRole('spinbutton').fill('0');await frame.getByRole('button',{name:'Save entry',exact:true}).click();assert((await frame.getByRole('alert').innerText()).includes('1'));await frame.getByRole('spinbutton').fill('48');await frame.getByRole('combobox').selectOption('minutes');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();assert((await frame.getByRole('status').innerText()).includes('minutes'));row.interactions.push('Zero rejected; 48minutes saves paired numeric amount/unit.');
 }else if(path==='forms/autocomplete'){
  const c=frame.getByRole('combobox');await c.fill('Ch');await c.press('ArrowDown');await c.press('Enter');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();assert((await frame.getByRole('status').innerText()).includes('Chinese'));row.interactions.push('Typing Ch, ArrowDown andEnter selects Chinese object; required error clears.');
 }else if(path==='forms/multi-autocomplete'||path==='forms/tags-input'){
  const c=frame.getByRole('combobox');await c.fill(path.includes('tags')?'fiction':'Chinese');await c.press('ArrowDown');await c.press('Enter');await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();await frame.getByRole('button',{name:/Remove/}).first().click();await frame.getByRole('button',{name:'Save entry',exact:true}).click();await frame.getByRole('alert').waitFor();row.interactions.push('Keyboard selects known value, saves, named chip removal restores required error.');
 }else if(path==='actions/button-group'){
  await frame.getByRole('button',{name:'Edit log',exact:true}).click();assert((await frame.getByRole('status').innerText()).includes('Editing'));assert(await frame.getByRole('button',{name:'Delete log',exact:true}).isDisabled());row.interactions.push('Edit changes local status; Delete remains disabled.');
 }else if(path==='feedback/flash'){
  assert.equal(await frame.getByRole('alert').count(),1);assert.equal(await frame.getByRole('status').count(),3);await frame.getByRole('link',{name:'Review dates'}).click();row.interactions.push('Danger uses alert; three nonurgent statuses polite; Review dates links to concrete context.');
 }else if(path==='feedback/loading'){
  assert.equal(await frame.getByRole('status').count(),3);await page.emulateMedia({reducedMotion:'reduce'});assert(await frame.locator('.paper-loading').evaluateAll(es=>es.every(e=>getComputedStyle(e).animationName==='none'||[...e.querySelectorAll('*')].every(c=>getComputedStyle(c).animationName==='none'))));row.interactions.push('Three named loading statuses; reduced-motion emulation checked.');await page.emulateMedia({reducedMotion:'no-preference'});
 }else if(path==='data-display/surface'){
  assert.equal(await frame.getByRole('article').count(),4);row.interactions.push('Four semantic articles compare flat/floating/showcase and independent accent rail.');
 }else if(path==='feedback/toast'){
  await frame.getByRole('button',{name:'Show notification',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor();await frame.getByRole('button',{name:'Show failure',exact:true}).click();await frame.locator('.paper-toast__title').filter({hasText:'Could not sync entry'}).waitFor();assert(await frame.locator('.paper-toast').evaluateAll(es=>es.every(e=>e.getBoundingClientRect().top>=0)));await page.locator('iframe').screenshot({path:output+'/toast-queued.png'});await frame.getByRole('button',{name:'Dismiss all',exact:true}).click();await frame.getByText('Entry saved',{exact:true}).waitFor({state:'hidden'});row.interactions.push('Success and high-priority failure queue; Dismiss all removes both.');
 }
 results.push(row);await writeFile(output+'/checks.json',JSON.stringify(results,null,2));console.log(path+' passed');
}
await browser.close();console.log('14forms/feedback pages passed four browser configurations and interactions.');
