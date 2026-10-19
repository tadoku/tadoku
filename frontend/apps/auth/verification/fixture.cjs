const http=require('node:http'),fs=require('node:fs');
const root=require('node:path').dirname(process.env.IDENTITY_UI_STATE); const account={id:'11111111-1111-4111-8111-111111111111',email:'fixture@example.test',display_name:'Fixture Reader',created_at:'2020-01-01T00:00:00Z',authenticated_at:'2026-10-19T00:00:00Z'};
for(const [port,target] of [[43270,43271],[43274,43272],[43275,43273]]) http.createServer((req,res)=>{
 const state=JSON.parse(fs.readFileSync(root+'/ui-state.json'));
 if(req.url.startsWith('/api/internal/')){
  fs.appendFileSync(root+'/ui-requests.jsonl',JSON.stringify({port,method:req.method,path:req.url,branch:req.headers['x-dev-branch']||'',cookieNames:(req.headers.cookie||'').split(';').map(v=>v.trim().split('=')[0])})+'\n'); res.setHeader('content-type','application/json');
  if(req.url==='/api/internal/identity/session'){
   if(state.status==='timeout')return setTimeout(()=>{res.statusCode=503;res.end('{}')},12000);
   res.statusCode=state.status||200;return res.end(res.statusCode===200?JSON.stringify({account_app:state.app||'v1',account:state.signedIn?account:null}):'{}');
  }
  if(req.url==='/api/internal/identity/logout'){res.statusCode=state.logoutStatus||204;return res.end('{}')}
  if(req.url.startsWith('/api/internal/profile/users'))return res.end(JSON.stringify({users:[{...account,role:'user'}],total_size:1}));
  if(req.url.includes('/authz/current-user/role'))return res.end(JSON.stringify({role:state.role||'user'}));
  if(req.url.includes('feature-flags'))return res.end('{"decisions":{}}');
  if(state.queryStatus){res.statusCode=state.queryStatus;return res.end('{}')}
  return res.end('{}');
 }
 const u=http.request({hostname:'127.0.0.1',port:target,path:req.url,method:req.method,headers:req.headers},r=>{res.writeHead(r.statusCode,r.headers);r.pipe(res)});u.on('error',()=>{res.statusCode=502;res.end('frontend unavailable')});req.pipe(u);
}).listen(port,'0.0.0.0');
console.log('Fixture API only; not authentication or live gateway proof.');
