import { lazy, Suspense, useEffect, useState } from 'react'
import { Link, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { FormProvider, useForm } from 'react-hook-form'
import { ActionMenu, Button, Drawer, Loading, Navbar, Select, ToastContainer, buttonClassName, type NavbarItem } from 'paper-ui'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'
import { scenarioGroup, scenarios, usePlayground, useScenario, type GuideIA, type Viewer } from './state'
import { ContactPage, LogEditorPage, SettingsPage, SignInPage } from './SupportingPages'

const HomePage=lazy(()=>import('./pages/HomePage').then(m=>({default:m.HomePage})))
const GuidePage=lazy(()=>import('./pages/GuidePage').then(m=>({default:m.GuidePage})))
const LeaderboardPage=lazy(()=>import('./pages/LeaderboardPage').then(m=>({default:m.LeaderboardPage})))
const ContestsPage=lazy(()=>import('./pages/ContestsPage').then(m=>({default:m.ContestsPage})))
const ContestDetailPage=lazy(()=>import('./pages/ContestsPage').then(m=>({default:m.ContestDetailPage})))
const ContestRegistrationPage=lazy(()=>import('./pages/ContestsPage').then(m=>({default:m.ContestRegistrationPage})))
const ContestEditorPage=lazy(()=>import('./pages/ContestsPage').then(m=>({default:m.ContestEditorPage})))
const ProfilePage=lazy(()=>import('./pages/ProfilePages').then(m=>({default:m.ProfilePage})))
const LogPage=lazy(()=>import('./pages/LogPage').then(m=>({default:m.LogPage})))
const AdminPage=lazy(()=>import('./pages/AdminPage').then(m=>({default:m.AdminPage})))

const reviewPages=[['Home','/'],['Profile overview','/users/anton'],['Activity log','/users/anton/activity'],['Admin workspace','/admin'],['About Tadoku','/about'],['Start here','/guide'],['Scoring & rules','/guide/scoring'],['Questions & help','/guide/questions'],['Leaderboard','/leaderboard/latest'],['Contests','/contests/official'],['Log details','/logs/reading-konbini?scenario=reading']] as const

function ReviewControls() {
  const app=usePlayground();const location=useLocation();const navigate=useNavigate();const [open,setOpen]=useState(false)
  const group=scenarioGroup(location.pathname);const [scenario]=useScenario(group)
  const methods=useForm({defaultValues:{theme:app.theme,density:app.density,viewer:app.viewer,ia:app.ia,scenario}})
  return <Drawer trigger={<Button variant="ghost">Playground controls</Button>} title="Paper playground" description="Unpublished application study · sample data only" open={open} onOpenChange={next=>{if(next)methods.reset({theme:app.theme,density:app.density,viewer:app.viewer,ia:app.ia,scenario});setOpen(next)}}>
    <FormProvider {...methods}><form className="grid gap-4" onSubmit={methods.handleSubmit(values=>{
      app.setTheme(values.theme);app.setDensity(values.density);app.setIa(values.ia as GuideIA);app.setViewer(values.viewer as Viewer)
      const next=new URLSearchParams(location.search)
      if(values.viewer!==app.viewer)next.delete('scenario');else if(values.scenario)next.set('scenario',values.scenario)
      navigate({pathname:location.pathname,search:next.toString()},{replace:true});setOpen(false)
    })}>
      <Select name="theme" label="Appearance" options={[{value:'light',label:'Light'},{value:'dark',label:'Dark'}]} />
      <Select name="density" label="Density" options={[{value:'comfortable',label:'Comfortable'},{value:'compact',label:'Compact'}]} />
      <Select name="viewer" label="Sample account" options={(['guest','member','participant','organizer','admin'] as const).map(value=>({value,label:{guest:'Guest',member:'Member, not joined',participant:'Participant',organizer:'Organizer',admin:'Administrator'}[value]}))} />
      {scenarios[group].length>0&&<Select name="scenario" label="Page scenario" options={scenarios[group].map(([value,label])=>({value,label}))} />}
      <Select name="ia" label="Help navigation" options={[{value:'guide',label:'B · Guide plus About'},{value:'separate',label:'A · Four familiar pages'},{value:'tasks',label:'C · Task-based help'}]} />
      <Button type="submit">Apply view</Button>
    </form></FormProvider>
    <p className="review-note">Entries, registrations, settings, and administrative actions stay in this browser. The sample date is 5 September 2026; lifecycle scenarios supply their own dates.</p>
    <h2 className="paper-type-component mt-8 mb-4">All page designs</h2>
    <nav aria-label="Playground pages" className="review-pages">{reviewPages.map(([label,to])=><Link className="text-link" key={to} to={to} onClick={()=>setOpen(false)}>{label}</Link>)}</nav>
    <p className="review-note mt-6"><a className="text-link" href="https://draft.apps.lab/documents/tadoku-redesign-master.html?version=3" target="_blank" rel="noreferrer">Open the pinned design brief</a></p>
    <Button variant="outline" onClick={()=>{app.reset();navigate('/');setOpen(false)}}>Reset sample changes</Button>
  </Drawer>
}

export function App() {
  const app=usePlayground();const location=useLocation();const navigate=useNavigate();const admin=location.pathname.startsWith('/admin')
  const helpLabel={guide:'Guide',separate:'Learn',tasks:'Help'}[app.ia]
  const accountItems=[{id:'profile',label:'My profile',onSelect:()=>navigate('/users/anton')},{id:'settings',label:'Account settings',onSelect:()=>navigate('/settings')},...(app.viewer==='admin'?[{id:'admin',label:'Admin workspace',onSelect:()=>navigate('/admin')}]:[]),{id:'signout',label:'Log out',onSelect:()=>{app.setViewer('guest');navigate('/')}}]
  const nav:NavbarItem[]=admin?[]:[{type:'link',id:'leaderboard',label:'Leaderboard',href:'/leaderboard/latest',current:location.pathname.includes('leaderboard')},{type:'link',id:'contests',label:'Contests',href:'/contests/official',current:location.pathname.startsWith('/contests')&&!location.pathname.includes('leaderboard')},{type:'link',id:'guide',label:helpLabel,href:'/guide',current:location.pathname.startsWith('/guide')}]
  useEffect(()=>{
    window.scrollTo({top:0,left:0,behavior:'instant'})
    const update=()=>{const title=document.querySelector('main h1')?.textContent;if(title)document.title=`${title} · Tadoku`}
    const observer=new MutationObserver(update);observer.observe(document.getElementById('main-content')!,{childList:true,subtree:true});update()
    return()=>observer.disconnect()
  },[location.pathname])
  useEffect(()=>{if(location.hash){const id=decodeURIComponent(location.hash.slice(1));requestAnimationFrame(()=>document.getElementById(id)?.scrollIntoView())}},[location.hash,location.pathname])
  return <div className="app-shell">
    <a className={`${buttonClassName()} app-skip`} href="#main-content">Skip to content</a>
    <header className="app-header">
      <Navbar
        brand={<div className="flex items-center gap-2"><img className="app-wordmark" src={app.theme==='dark'?reversedWordmark:wordmark} width="158" height="29" alt="Tadoku"/>{admin&&<span className="app-admin-label">Admin</span>}</div>}
        brandHref="/" navigation={nav} renderLink={({href,...props})=><Link to={href} {...props}/>}
        mobileNavigation={!admin}
        actions={admin ? <Link to="/" className={buttonClassName({variant:'outline'})}>Back to Tadoku</Link> :
          <div className="app-header-actions">
            {app.viewer==='guest' ? <Link to={`/sign-in?next=${encodeURIComponent(location.pathname)}`} className={buttonClassName({variant:'outline'})}>Sign in</Link> : <>
              <span className="app-account-menu--desktop"><ActionMenu label={app.user?.name||'Account'} triggerVariant="ghost" items={accountItems}/></span>
              <Link to="/logs/new" className={buttonClassName({className:'whitespace-nowrap shrink-0'})}>Log activity</Link>
            </>}
          </div>}
        mobileFooter={app.viewer==='guest' ? undefined : closeMenu =>
          <section className="app-mobile-account" aria-label={`${app.user?.name||'Your'} account`}>
            <p className="app-mobile-account__name">{app.user?.name||'Account'}</p>
            <Link className="paper-navbar__mobile-link" to="/users/anton" aria-current={location.pathname==='/users/anton'?'page':undefined} onClick={closeMenu}>My profile</Link>
            <Link className="paper-navbar__mobile-link" to="/settings" aria-current={location.pathname==='/settings'?'page':undefined} onClick={closeMenu}>Account settings</Link>
            {app.viewer==='admin'&&<Link className="paper-navbar__mobile-link" to="/admin" onClick={closeMenu}>Admin workspace</Link>}
            <Button variant="ghost" className="paper-navbar__mobile-link justify-start" onClick={()=>{closeMenu();app.setViewer('guest');navigate('/')}}>Log out</Button>
          </section>}
      />
    </header>
    <main id="main-content" className="app-content" tabIndex={-1}><Suspense fallback={<Loading label="Loading page"/>}><Routes>
      <Route path="/" element={<HomePage/>}/>
      <Route path="/about" element={<GuidePage page="about"/>}/><Route path="/guide" element={<GuidePage page="manual"/>}/><Route path="/guide/scoring" element={<GuidePage page="rules"/>}/><Route path="/guide/questions" element={<GuidePage page="faq"/>}/>
      <Route path="/leaderboard" element={<Navigate to="/leaderboard/latest" replace/>}/><Route path="/leaderboard/latest" element={<LeaderboardPage/>}/><Route path="/leaderboard/yearly/:year" element={<LeaderboardPage/>}/><Route path="/leaderboard/all-time" element={<LeaderboardPage/>}/><Route path="/contests/:contestId/leaderboard" element={<LeaderboardPage/>}/>
      <Route path="/contests" element={<Navigate to="/contests/official" replace/>}/><Route path="/contests/official" element={<ContestsPage/>}/><Route path="/contests/community" element={<ContestsPage/>}/><Route path="/contests/participating" element={<ContestsPage/>}/><Route path="/contests/managed" element={<ContestsPage/>}/><Route path="/contests/mine" element={<ContestsPage/>}/><Route path="/contests/new" element={<ContestEditorPage/>}/><Route path="/contests/:contestId/edit" element={<ContestEditorPage/>}/><Route path="/contests/:contestId/registration" element={<ContestRegistrationPage/>}/><Route path="/contests/:contestId" element={<ContestDetailPage/>}/>
      <Route path="/users/:userId" element={<ProfilePage/>}/><Route path="/users/:userId/activity" element={<ProfilePage activity/>}/>
      <Route path="/logs/new" element={<LogEditorPage/>}/><Route path="/logs/:logId/edit" element={<LogEditorPage/>}/><Route path="/logs/reading" element={<Navigate to="/logs/reading-konbini" replace/>}/><Route path="/logs/listening" element={<Navigate to="/logs/listening-teppei" replace/>}/><Route path="/logs/:logId" element={<LogPage/>}/>
      <Route path="/admin" element={<AdminPage/>}/><Route path="/admin/:section" element={<AdminPage/>}/>
      <Route path="/sign-in" element={<SignInPage/>}/><Route path="/settings" element={<SettingsPage/>}/><Route path="/account/settings" element={<Navigate to="/settings" replace/>}/><Route path="/contact" element={<ContactPage/>}/>
      {Object.entries({about:'/about',manual:'/guide',rules:'/guide/scoring',faq:'/guide/questions'}).map(([slug,to])=><Route key={slug} path={`/pages/${slug}`} element={<Navigate to={to} replace/>}/>)}
      <Route path="*" element={<section className="empty-state"><h1 className="paper-type-page">Page not found</h1><p>This address doesn’t match a Tadoku page.</p><Link to="/" className="text-link">Return home</Link></section>}/>
    </Routes></Suspense></main>
    <footer className="app-footer"><div className="app-footer__inner"><nav aria-label="Footer"><Link className="text-link" to="/about">About Tadoku</Link><Link className="text-link" to="/guide/questions">Questions &amp; help</Link><a className="text-link" href="https://github.com/tadoku/tadoku" target="_blank" rel="noreferrer">GitHub</a></nav><div className="flex flex-wrap items-center gap-2"><small className="muted">Sample data</small><ReviewControls key={scenarioGroup(location.pathname)}/></div></div></footer>
    <ToastContainer />
  </div>
}
