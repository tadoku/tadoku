import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useParams } from 'react-router-dom'
import { ToastProvider } from 'paper-ui'
import { beforeEach, expect, it } from 'vitest'
import { initialContests, initialLogs } from '../src/data'
import { LogEditorPage } from '../src/SupportingPages'
import { PlaygroundProvider, usePlayground } from '../src/state'

function Saved(){const {logId}=useParams();const {logs}=usePlayground();return <output aria-label="Saved record">{JSON.stringify(logs.find(log=>log.id===logId))}</output>}
function show(path:string){render(<MemoryRouter initialEntries={[path]}><PlaygroundProvider><ToastProvider><Routes><Route path="/logs/new" element={<LogEditorPage/>}/><Route path="/logs/:logId/edit" element={<LogEditorPage/>}/><Route path="/logs/:logId" element={<Saved/>}/></Routes></ToastProvider></PlaygroundProvider></MemoryRouter>)}
beforeEach(()=>{localStorage.clear();localStorage.setItem('tadoku-paper-playground-v1',JSON.stringify({viewer:'participant',logs:initialLogs,contests:initialContests,joinedContests:['round5','reading-circle']}))})
it('validates a positive amount before saving a new personal entry',async()=>{
  show('/logs/new')
  fireEvent.change(screen.getByLabelText('Pages read',{exact:false}),{target:{value:'0'}})
  fireEvent.click(screen.getByRole('button',{name:'Save activity'}))
  expect(await screen.findByRole('alert')).toHaveTextContent('Enter an amount greater than zero')
  expect(screen.queryByLabelText('Saved record')).not.toBeInTheDocument()
  fireEvent.change(screen.getByLabelText('Pages read',{exact:false}),{target:{value:'12'}})
  fireEvent.change(screen.getByLabelText('What did you read or listen to?'),{target:{value:'A quiet chapter'}})
  fireEvent.click(screen.getByRole('button',{name:'Save activity'}))
  const saved=JSON.parse((await screen.findByLabelText('Saved record')).textContent||'{}')
  expect(saved).toMatchObject({title:'A quiet chapter',amount:12,language:'Japanese',activity:'Reading',submissions:[]})
})
it('updates active contest scores with their own rule when editing',async()=>{
  show('/logs/reading-konbini/edit')
  fireEvent.change(screen.getByLabelText('Pages read',{exact:false}),{target:{value:'48'}})
  fireEvent.change(screen.getByLabelText('Time spent reading'),{target:{value:'50'}})
  fireEvent.click(screen.getByRole('button',{name:'Save changes'}))
  const saved=JSON.parse((await screen.findByLabelText('Saved record')).textContent||'{}')
  expect(saved.submissions).toEqual(expect.arrayContaining([expect.objectContaining({contestId:'round5',score:48}),expect.objectContaining({contestId:'reading-circle',score:25})]))
})
it('keeps closed contribution snapshots while editing the personal record',async()=>{
  show('/logs/ended-reading/edit')
  fireEvent.change(screen.getByLabelText('Pages read',{exact:false}),{target:{value:'60'}})
  fireEvent.click(screen.getByRole('button',{name:'Save changes'}))
  await waitFor(()=>expect(screen.getByLabelText('Saved record')).toBeInTheDocument())
  const saved=JSON.parse(screen.getByLabelText('Saved record').textContent||'{}')
  expect(saved.amount).toBe(60)
  expect(saved.submissions.map((s:{score:number})=>s.score)).toEqual([42,22.5])
})

it.each(['reading-konbini','listening-teppei'])('keeps the existing language and activity fixed while saving %s',async id=>{
  const existing=initialLogs.find(log=>log.id===id)!
  show(`/logs/${id}/edit`)
  expect(screen.getByRole('combobox',{name:'Activity'})).toBeDisabled()
  expect(screen.getByRole('combobox',{name:'Activity'})).toHaveValue(existing.activity)
  expect(screen.getByRole('combobox',{name:/Language/})).toBeDisabled()
  expect(screen.getByRole('combobox',{name:/Language/})).toHaveValue(existing.language)
  fireEvent.change(screen.getByLabelText('Notes'),{target:{value:'An updated note'}})
  fireEvent.click(screen.getByRole('button',{name:'Save changes'}))
  const saved=JSON.parse((await screen.findByLabelText('Saved record')).textContent||'{}')
  expect(saved).toMatchObject({language:existing.language,activity:existing.activity,unit:existing.unit,note:'An updated note'})
})

it('allows a later activity to be edited when reopened without its scenario date',async()=>{
  const october={...initialLogs[0],id:'october-reading',date:'2026-10-01',submissions:[]}
  localStorage.setItem('tadoku-paper-playground-v1',JSON.stringify({viewer:'participant',logs:[...initialLogs,october],contests:initialContests,joinedContests:['round5']}))
  show('/logs/october-reading/edit')
  expect(screen.getByLabelText(/Activity date/)).toHaveAttribute('max','2026-10-01')
  fireEvent.change(screen.getByLabelText(/Pages read/),{target:{value:'50'}})
  fireEvent.click(screen.getByRole('button',{name:'Save changes'}))
  const saved=JSON.parse((await screen.findByLabelText('Saved record')).textContent||'{}')
  expect(saved).toMatchObject({date:'2026-10-01',amount:50})
})

it('retains explicit later date context and finished scores when editing an older activity',async()=>{
  show('/logs/reading-konbini/edit?asOf=2026-10-05')
  expect(screen.getByLabelText(/Activity date/)).toHaveAttribute('max','2026-10-05')
  fireEvent.change(screen.getByLabelText(/Pages read/),{target:{value:'50'}})
  fireEvent.click(screen.getByRole('button',{name:'Save changes'}))
  const saved=JSON.parse((await screen.findByLabelText('Saved record')).textContent||'{}')
  expect(saved).toMatchObject({date:'2026-09-05',amount:50})
  expect(saved.submissions.map((submission:{score:number})=>submission.score)).toEqual([42,22.5])
})
