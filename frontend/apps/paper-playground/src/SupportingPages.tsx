import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { FormProvider, useForm } from 'react-hook-form'
import { Button, Flash, Input, Select, TextArea, buttonClassName, useToast } from 'paper-ui'
import { usePlayground } from './state'

function safeNext(value:string|null){return value?.startsWith('/')&&!value.startsWith('//')?value:'/users/anton'}

export function SignInPage(){
  const app=usePlayground();const navigate=useNavigate();const [params]=useSearchParams()
  return <section className="app-form app-entry"><header className="page-header"><h1 className="paper-type-page">Join in, keep going.</h1><p className="page-lead">Join a contest and keep a record of what you read and listen to.</p></header><div className="app-entry__account"><h2>Try Tadoku as Anton</h2><p>This playground uses a sample account. No email address or password is needed.</p><div className="app-form__actions"><Button onClick={()=>{app.setViewer('member');navigate(safeNext(params.get('next')))}}>Continue as Anton</Button><Link to="/" className={buttonClassName({variant:'outline'})}>Keep browsing</Link></div></div></section>
}

export { LogEditorPage } from './pages/LogEditorPage'

export function SettingsPage(){
  const app=usePlayground();const toast=useToast();const methods=useForm({defaultValues:{name:app.user?.name||'Anton',theme:app.theme,density:app.density}})
  if(app.viewer==='guest')return <Navigate replace to="/sign-in?next=/settings"/>
  return <section className="app-form"><header className="page-header"><Link className="text-link page-back-link" to="/users/anton">Your profile</Link><h1 className="paper-type-page">Account settings</h1></header><FormProvider {...methods}><form onSubmit={methods.handleSubmit(v=>{app.setDisplayName(v.name.trim());app.setTheme(v.theme);app.setDensity(v.density);toast.add({title:'Settings saved',description:'Your preferences stay in this browser.'})})}><fieldset className="app-form__section"><legend>Your profile</legend><Input name="name" label="Display name" hint="Shown on your profile, activity and leaderboard entries." required maxLength={40} rules={{validate:value=>String(value).trim().length>0||'Enter a name.'}}/></fieldset><fieldset className="app-form__section"><legend>Appearance</legend><div className="app-form__row"><Select name="theme" label="Color theme" options={[{value:'light',label:'Light'},{value:'dark',label:'Dark'}]}/><Select name="density" label="Control density" hint="Compact makes form controls and lists shorter." options={[{value:'comfortable',label:'Comfortable'},{value:'compact',label:'Compact'}]}/></div></fieldset><div className="app-form__actions"><Button type="submit">Save settings</Button><Link className={buttonClassName({variant:'outline'})} to="/users/anton">View profile</Link></div></form></FormProvider></section>
}

export function ContactPage(){
  const methods=useForm({defaultValues:{topic:'Question',message:''}});const [draft,setDraft]=useState('');const preview=useRef<HTMLDivElement>(null)
  useEffect(()=>{if(draft)preview.current?.focus()},[draft])
  return <section><header className="page-header"><h1 className="paper-type-page">Get in touch</h1></header><div className="app-editor-layout"><div className="app-form"><FormProvider {...methods}><form onSubmit={methods.handleSubmit(values=>{setDraft(`${values.topic}\n\n${values.message}`)})}><div className="app-form__fields"><Select name="topic" label="Topic" options={['Question','Problem report','Feature idea','Missing language'].map(value=>({value,label:value}))}/><TextArea name="message" label="Your message" required rows={7} hint="Include what happened and what you expected, if you’re reporting a problem."/></div><div className="app-form__actions"><Button type="submit">Prepare message</Button></div><p className="review-note">This playground prepares a local draft. It does not send a message.</p></form></FormProvider>{draft&&<div className="mt-6" ref={preview} tabIndex={-1}><Flash title="Message prepared"><p className="whitespace-pre-wrap">{draft}</p><Button variant="outline" onClick={()=>{void navigator.clipboard.writeText(draft)}}>Copy message</Button></Flash></div>}</div><aside className="app-editor-aside"><h2>A quick answer</h2><p>Find help with logging, scoring and joining a contest.</p><Link className="text-link" to="/guide/questions">Questions &amp; help</Link><h2>Software problems</h2><p>Check whether someone has already reported the issue.</p><a className="text-link" href="https://github.com/tadoku/tadoku/issues" target="_blank" rel="noreferrer">View GitHub issues</a></aside></div></section>
}
