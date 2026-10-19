import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { usePaperForm, Button, Flash, PageHeader, buttonClassName, useToast } from 'paper-ui'
import { renderLink } from './routerLink'
import { usePlayground } from './state'

function safeNext(value:string|null){return value?.startsWith('/')&&!value.startsWith('//')?value:'/users/anton'}

export function SignInPage(){
  const app=usePlayground();const navigate=useNavigate();const [params]=useSearchParams()
  return <section className="app-form app-entry"><PageHeader title="Join in, keep going." lead="Join a contest and keep a record of what you read and listen to."/><div className="app-entry__account"><h2>Try Tadoku as Anton</h2><p>This playground uses a sample account. No email address or password is needed.</p><div className="app-form__actions"><Button onClick={()=>{app.setViewer('member');navigate(safeNext(params.get('next')))}}>Continue as Anton</Button><Link to="/" className={buttonClassName({variant:'outline'})}>Keep browsing</Link></div></div></section>
}

export { LogEditorPage } from './pages/LogEditorPage';

export function SettingsPage(){
  const app=usePlayground();const toast=useToast();const methods=usePaperForm({defaultValues:{name:app.user?.name||'Anton',theme:app.theme,density:app.density}, onSubmit: ({ value: v }) => {app.setDisplayName(v.name.trim());app.setTheme(v.theme);app.setDensity(v.density);toast.add({title:'Settings saved',description:'Your preferences stay in this browser.'})} })
  if(app.viewer==='guest')return <Navigate replace to="/sign-in?next=/settings"/>
  return <section className="app-form"><PageHeader title="Account settings" backLink={{href:'/users/anton',label:'Your profile'}} renderLink={renderLink}/><methods.AppForm><methods.Form ><fieldset className="app-form__section"><legend>Your profile</legend><methods.AppField name="name" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value=>String(value).trim().length>0||'Enter a name.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Display name" hint="Shown on your profile, activity and leaderboard entries." required maxLength={40} />}</methods.AppField></fieldset><fieldset className="app-form__section"><legend>Appearance</legend><div className="app-form__row"><methods.AppField name="theme">{field => <field.Select label="Color theme" options={[{value:'light',label:'Light'},{value:'dark',label:'Dark'}]} />}</methods.AppField><methods.AppField name="density">{field => <field.Select label="Control density" hint="Compact makes form controls and lists shorter." options={[{value:'comfortable',label:'Comfortable'},{value:'compact',label:'Compact'}]} />}</methods.AppField></div></fieldset><div className="app-form__actions"><Button type="submit">Save settings</Button><Link className={buttonClassName({variant:'outline'})} to="/users/anton">View profile</Link></div></methods.Form></methods.AppForm></section>
}

export function ContactPage(){
  const methods=usePaperForm({defaultValues:{topic:'Question',message:''}, onSubmit: ({ value: values }) => {setDraft(`${values.topic}\n\n${values.message}`)} });const [draft,setDraft]=useState('');const preview=useRef<HTMLDivElement>(null)
  useEffect(()=>{if(draft)preview.current?.focus()},[draft])
  return <section><PageHeader title="Get in touch"/><div className="app-editor-layout"><div className="app-form"><methods.AppForm><methods.Form ><div className="app-form__fields"><methods.AppField name="topic">{field => <field.Select label="Topic" options={['Question','Problem report','Feature idea','Missing language'].map(value=>({value,label:value}))} />}</methods.AppField><methods.AppField name="message" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; } }}>{field => <field.TextArea label="Your message" required rows={7} hint="Include what happened and what you expected, if you’re reporting a problem." />}</methods.AppField></div><div className="app-form__actions"><Button type="submit">Prepare message</Button></div><p className="review-note">This playground prepares a local draft. It does not send a message.</p></methods.Form></methods.AppForm>{draft&&<div className="mt-6" ref={preview} tabIndex={-1}><Flash title="Message prepared"><p className="whitespace-pre-wrap">{draft}</p><Button variant="outline" onClick={()=>{void navigator.clipboard.writeText(draft)}}>Copy message</Button></Flash></div>}</div><aside className="app-editor-aside"><h2>A quick answer</h2><p>Find help with logging, scoring and joining a contest.</p><Link className="text-link" to="/guide/questions">Questions &amp; help</Link><h2>Software problems</h2><p>Check whether someone has already reported the issue.</p><a className="text-link" href="https://github.com/tadoku/tadoku/issues" target="_blank" rel="noreferrer">View GitHub issues</a></aside></div></section>
}
