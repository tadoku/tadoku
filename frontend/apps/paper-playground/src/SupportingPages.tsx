import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Button, Checkbox, Flash, Input, Select, TagsInput, TextArea, buttonClassName, useToast } from 'paper-ui'
import { contestStatus, formatNumber, sampleToday, scoreLog, scoreSubmission, supportedLanguages, type Activity, type SampleLog } from './data'
import { usePlayground } from './state'

function safeNext(value:string|null){return value?.startsWith('/')&&!value.startsWith('//')?value:'/users/anton'}
export function SignInPage(){
  const app=usePlayground();const navigate=useNavigate();const [params]=useSearchParams()
  return <section className="app-form"><header className="page-header"><h1 className="paper-type-page">Join in, keep going.</h1><p className="page-lead">Use the sample account to join a contest and keep a record of what you read and listen to.</p></header><Flash title="Try Tadoku as Anton">This unpublished playground uses a sample account. No email address or password is needed.</Flash><div className="app-form__actions"><Button onClick={()=>{app.setViewer('member');navigate(safeNext(params.get('next')))}}>Continue as Anton</Button><Link to="/" className={buttonClassName({variant:'outline'})}>Keep browsing</Link></div></section>
}

export function LogEditorPage(){
  const {logId}=useParams();const app=usePlayground();const existing=app.logs.find(log=>log.id===logId)
  if(app.viewer==='guest')return <Navigate replace to={`/sign-in?next=${encodeURIComponent(logId?`/logs/${logId}/edit`:'/logs/new')}`}/>
  if(logId&&!existing)return <section className="empty-state"><h1 className="paper-type-page">Log not found</h1><Link className="text-link" to="/users/anton/activity">View your activity</Link></section>
  if(existing&&existing.userId!==app.userId)return <section className="empty-state"><h1 className="paper-type-page">This is someone else’s record</h1><Link className="text-link" to={`/logs/${existing.id}`}>View the log</Link></section>
  return <LogForm key={logId||'new'} existing={existing}/>
}

function LogForm({existing}:{existing?:SampleLog}){
  const app=usePlayground();const navigate=useNavigate();const toast=useToast();const [params]=useSearchParams()
  const requestedDate=params.get('asOf')
  const contextDate=requestedDate&&/^\d{4}-\d{2}-\d{2}$/.test(requestedDate)?requestedDate:existing?.submissions.some(s=>s.closed)?'2026-10-03':sampleToday
  const today=existing&&existing.date>contextDate?existing.date:contextDate
  const detailHref=existing?`/logs/${existing.id}?asOf=${today}`:'/users/anton/activity'
  const closedSubmissions=(existing?.submissions||[]).filter(submission=>{
    const contest=app.contests.find(c=>c.id===submission.contestId)
    return submission.closed||contest&&contestStatus(contest,today)==='ended'
  }).map(submission=>({...submission,closed:true}))
  const methods=useForm({defaultValues:{title:existing?.title||'',language:existing?.language||'Japanese',activity:existing?.activity||'Reading',amount:existing?.amount??'',minutes:existing?.minutes??'',date:existing?.date||today,note:existing?.note||'',tags:existing?.tags||[],media:existing?.media||'Book',submissions:existing?.submissions.filter(s=>!closedSubmissions.some(closed=>closed.contestId===s.contestId)).map(s=>s.contestId)||[params.get('contest')].filter(Boolean) as string[]}})
  const selectedActivity=useWatch({control:methods.control,name:'activity'}) as Activity
  const selectedLanguage=useWatch({control:methods.control,name:'language'})
  const activity=existing?.activity??selectedActivity
  const language=existing?.language??selectedLanguage
  const date=useWatch({control:methods.control,name:'date'})
  const amount=Number(useWatch({control:methods.control,name:'amount'}))
  const minutes=Number(useWatch({control:methods.control,name:'minutes'}))
  const eligible=app.contests.filter(c=>app.joinedContests.includes(c.id)&&contestStatus(c,today)==='live'&&date>=c.start&&date<=c.end&&c.activities.includes(activity)&&(c.languages.includes('All languages')||c.languages.includes(language))&&(!app.registrations[c.id]||app.registrations[c.id].includes(language))&&(c.id!=='reading-circle'||minutes>0))
  return <section><header className="page-header"><Link className="text-link" to={detailHref}>{existing?'Back to log':'Your activity log'}</Link><h1 className="paper-type-page mt-4">{existing?'Edit activity':'Log activity'}</h1><p className="page-lead">A little today becomes a lasting habit. Record what you read or listened to.</p></header>
    <div className="page-columns"><FormProvider {...methods}><form className="app-form" noValidate onSubmit={methods.handleSubmit(values=>{
      const id=existing?.id||`local-${crypto.randomUUID()}`
      const submissions=[...closedSubmissions,...values.submissions.filter(id=>eligible.some(c=>c.id===id)).map(contestId=>scoreSubmission({activity,amount:Number(values.amount),unit:activity==='Reading'?'pages':'minutes',minutes:values.minutes!==''?Number(values.minutes):undefined},contestId))]
      const log:SampleLog={id,userId:app.userId,title:values.title.trim(),language,activity,amount:Number(values.amount),unit:activity==='Reading'?'pages':'minutes',minutes:activity==='Reading'&&values.minutes!==''?Number(values.minutes):undefined,date:values.date,note:values.note,tags:values.tags,media:values.media,submissions}
      app.saveLog(log);toast.add({title:existing?'Activity updated':'Activity saved',description:'Your personal record is up to date.'});navigate(`/logs/${id}?asOf=${today}`)
    })}><div className="app-form__fields">
      <div className="app-form__row"><Select name="activity" label="Activity" disabled={Boolean(existing)} value={existing?.activity} options={[{value:'Reading',label:'Reading'},{value:'Listening',label:'Listening'}]}/><Select name="language" label="Language" disabled={Boolean(existing)} value={existing?.language} required={!existing} options={supportedLanguages.map(value=>({value,label:value}))}/></div>
      {existing&&<p className="muted text-sm m-0">Language and activity stay fixed for existing entries.</p>}
      <Input name="title" label="What did you read or listen to?" hint="Optional. A book, episode, article, or anything you enjoyed." maxLength={180}/>
      <div className="app-form__row"><Input name="amount" label={activity==='Reading'?'Pages read':'Minutes listened'} type="number" min="0.1" step="0.1" required rules={{valueAsNumber:true,min:{value:.1,message:'Enter an amount greater than zero.'},max:{value:1000000,message:'Check this amount.'}}}/><Input name="date" label="Activity date" type="date" max={today} required rules={{validate:value=>String(value)<=today||'Choose today or an earlier date.'}} hint="All dates use UTC."/></div>
      {activity==='Reading'&&<Input name="minutes" label="Time spent reading" type="number" min="0" step="1" hint="Optional, in minutes. Leave blank if you didn’t track time." rules={{validate:value=>value===''||Number(value)>=0||'Use zero or a positive duration.'}}/>}
      <Select name="media" label="Medium" options={['Book','Article','Manga','Game','Podcast','Video','Audiobook','Conversation'].map(value=>({value,label:value}))}/>
      <TextArea name="note" label="Notes" rows={4} hint="Optional. Keep a thought, a new word, or where you left off."/>
      <TagsInput name="tags" label="Tags" options={[...new Set(['fiction','bookclub','daily','podcast',...(existing?.tags||[])])]} hint="Optional. Choose tags to find this entry again."/>
      <fieldset className="border border-solid border-rule p-4 m-0"><legend className="px-2 font-semibold">Submit to a contest</legend><p className="muted text-sm mt-0">Logging and contest participation are separate. Choose where this entry counts.</p>{eligible.length?eligible.map(c=><Checkbox key={c.id} name="submissions" value={c.id} label={c.title}/>):<p className="m-0">No joined contests accept this language, activity, and date. You can still save your personal record.</p>}</fieldset>
      {closedSubmissions.length>0&&<Flash title="Finished contests keep their saved scores">Editing this personal record leaves its finished-contest snapshots unchanged.</Flash>}
    </div><div className="app-form__actions"><Button type="submit">{existing?'Save changes':'Save activity'}</Button><Link to={detailHref} className={buttonClassName({variant:'outline'})}>Cancel</Link></div></form></FormProvider>
    <aside className="paper-accent-rail pl-6"><h2 className="paper-type-section">Your personal record</h2><p className="paper-type-page my-4">{Number.isFinite(amount)&&amount>0?formatNumber(scoreLog({activity,amount})):0} <span className="paper-type-body">points</span></p><p className="muted">This preview uses sample rates: one point per page and half a point per listening minute. Contest scores can use a different rule.</p><Link className="text-link" to="/guide/scoring">Understand score scopes</Link></aside></div>
  </section>
}

export function SettingsPage(){
  const app=usePlayground();const toast=useToast();const methods=useForm({defaultValues:{name:app.user?.name||'Anton',theme:app.theme,density:app.density}})
  if(app.viewer==='guest')return <Navigate replace to="/sign-in?next=/settings"/>
  return <section className="app-form"><header className="page-header"><h1 className="paper-type-page">Account settings</h1><p className="page-lead">Make your Tadoku workspace feel comfortable.</p></header><FormProvider {...methods}><form className="grid gap-6" onSubmit={methods.handleSubmit(v=>{app.setDisplayName(v.name.trim());app.setTheme(v.theme);app.setDensity(v.density);toast.add({title:'Settings saved',description:'Your preferences stay in this browser.'})})}><Input name="name" label="Display name" required maxLength={40} rules={{validate:value=>String(value).trim().length>0||'Enter a name.'}}/><Select name="theme" label="Appearance" options={[{value:'light',label:'Light'},{value:'dark',label:'Dark'}]}/><Select name="density" label="Control density" options={[{value:'comfortable',label:'Comfortable'},{value:'compact',label:'Compact'}]}/><div className="flex gap-2"><Button type="submit">Save settings</Button><Link className={buttonClassName({variant:'outline'})} to="/users/anton">View profile</Link></div></form></FormProvider></section>
}

export function ContactPage(){
  const methods=useForm({defaultValues:{topic:'Question',message:''}});const [draft,setDraft]=useState('');const preview=useRef<HTMLDivElement>(null)
  useEffect(()=>{if(draft)preview.current?.focus()},[draft])
  return <section className="app-form"><header className="page-header"><h1 className="paper-type-page">Get in touch</h1><p className="page-lead">Have a question, found a problem, or have an idea for Tadoku?</p></header><p>Search <Link className="text-link" to="/guide/questions">Questions &amp; help</Link> first. For a software problem, you can review the project’s <a className="text-link" href="https://github.com/tadoku/tadoku/issues" target="_blank" rel="noreferrer">GitHub issues</a>.</p><FormProvider {...methods}><form className="grid gap-4" onSubmit={methods.handleSubmit(values=>{setDraft(`${values.topic}\n\n${values.message}`)})}><Select name="topic" label="Topic" options={['Question','Problem report','Feature idea','Missing language'].map(value=>({value,label:value}))}/><TextArea name="message" label="Your message" required rows={6}/><Button type="submit">Prepare message</Button><p className="review-note">This playground prepares a local draft. It does not send a message.</p></form></FormProvider>{draft&&<div className="mt-6" ref={preview} tabIndex={-1}><Flash title="Message prepared"><p className="whitespace-pre-wrap">{draft}</p><Button variant="outline" onClick={()=>{void navigator.clipboard.writeText(draft)}}>Copy message</Button></Flash></div>}</section>
}
