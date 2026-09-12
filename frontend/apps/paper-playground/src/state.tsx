import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import { initialContests, initialLogs, users, type SampleContest, type SampleLog } from './data'

export type Viewer = 'guest'|'member'|'participant'|'organizer'|'admin'
export type GuideIA = 'guide'|'separate'|'tasks'
export type ScenarioGroup = 'home'|'profile'|'leaderboard'|'contests'|'log'|'guide'|'other'
export const scenarios: Record<ScenarioGroup, readonly (readonly [string,string])[]> = {
  home:[['guest-live','Guest · live round'],['member-open','Member · not joined'],['participant-live','Participant · live round'],['upcoming','Upcoming round'],['between','Between rounds'],['unavailable','Contest unavailable']],
  leaderboard:[['participant','Live · participant'],['guest','Live · guest'],['member','Live · not joined'],['closed','Registration closed'],['closed-participant','Closed · participant'],['upcoming','Upcoming round'],['ended','Round ended'],['empty','No participants'],['error','Standings unavailable']],
  contests:[['member','Active participant'],['guest','Signed out'],['between','Between rounds'],['empty','No organized contests'],['creator','Organizer'],['unavailable','Data unavailable']],
  log:[['reading','Your reading log'],['listening','Your listening log'],['visitor','Someone else’s log'],['personal','No submissions'],['ended','After a contest ends'],['missing','Log not found'],['unavailable','Temporary error']],
  profile:[['profile','Profile overview'],['visitor','Visitor'],['empty','Empty activity']], guide:[],other:[],
}
export function scenarioGroup(path: string): ScenarioGroup {
  if(path==='/')return 'home'
  if(path.includes('leaderboard'))return 'leaderboard'
  if(path.startsWith('/contests'))return 'contests'
  if(path.startsWith('/logs/')&&path!='/logs/new'&&!path.endsWith('/edit'))return 'log'
  if(path.startsWith('/users'))return 'profile'
  if(path.startsWith('/guide')||path==='/about')return 'guide'
  return 'other'
}
function scenarioViewer(group: ScenarioGroup, value: string | null): Viewer | undefined {
  if(!value)return undefined
  if(['guest','guest-live','visitor'].includes(value))return 'guest'
  if(value==='creator')return 'organizer'
  if(group==='home'&&value==='unavailable')return 'participant'
  if(['member-open','closed'].includes(value)||(group==='leaderboard'&&value==='member'))return 'member'
  if(group==='log'||group==='profile'||['participant','participant-live','closed-participant','member','empty','between','upcoming','ended'].includes(value))return 'participant'
  return undefined
}
interface LocalState {
  viewer: Viewer; joinedContests: string[]; registrations: Record<string,string[]>; logs: SampleLog[]
  contests: SampleContest[]; theme:'light'|'dark'; density:'comfortable'|'compact'; ia:GuideIA; displayName:string
}
const initialState: LocalState = {viewer:'guest',joinedContests:['round5','reading-circle','listening-circle'],registrations:{round5:['Japanese','Spanish']},logs:initialLogs,contests:initialContests,theme:'light',density:'comfortable',ia:'guide',displayName:'Anton'}
const storageKey='tadoku-paper-playground-v1'
function readState():LocalState {
  try { const saved=JSON.parse(localStorage.getItem(storageKey)||'null') as LocalState|null
    if(saved&&Array.isArray(saved.logs)&&Array.isArray(saved.contests)&&Array.isArray(saved.joinedContests))return {...initialState,...saved,
      // Older previews predate moderator metadata; keep all other saved contest edits.
      contests:saved.contests.map(contest=>({...contest,moderatorIds:contest.moderatorIds??initialContests.find(sample=>sample.id===contest.id)?.moderatorIds})),
    }
  } catch { /* Local storage may be disabled; the playground still works in memory. */ }
  return initialState
}
function useStore() {
  const [state,setState]=useState(readState)
  const location=useLocation();const [params,setParams]=useSearchParams()
  const explicitViewer=scenarioViewer(scenarioGroup(location.pathname),params.get('scenario'))
  const scenarioKey=explicitViewer?`${scenarioGroup(location.pathname)}:${params.get('scenario')}`:null
  const [appliedScenario,setAppliedScenario]=useState<string|null>(null)
  // Selecting a sample also selects its account for the links that follow it.
  if(scenarioKey!==appliedScenario){
    setAppliedScenario(scenarioKey)
    if(explicitViewer&&state.viewer!==explicitViewer)setState({...state,viewer:explicitViewer})
  }
  const viewer=explicitViewer??(location.pathname.startsWith('/admin')?'admin':state.viewer)
  const joinedContests=viewer==='guest'||viewer==='member'?[]:state.joinedContests
  const userId='anton'
  const user=viewer==='guest'?null:{...users.find(u=>u.id===userId)!,name:state.displayName}
  useEffect(()=>{
    document.documentElement.dataset.theme=state.theme
    document.documentElement.dataset.density=state.density
    try {localStorage.setItem(storageKey,JSON.stringify(state))}catch{/* Keep in-memory state when storage is unavailable. */}
  },[state])
  const clearScenario=()=>{if(params.has('scenario')){const next=new URLSearchParams(params);next.delete('scenario');setParams(next,{replace:true})}}
  return {
    ...state,viewer,joinedContests,userId,user,
    setViewer:(next:Viewer)=>{clearScenario();setState(s=>({...s,viewer:next}))},
    setTheme:(theme:LocalState['theme'])=>setState(s=>({...s,theme})),
    setDensity:(density:LocalState['density'])=>setState(s=>({...s,density})),
    setIa:(ia:GuideIA)=>setState(s=>({...s,ia})),
    setDisplayName:(displayName:string)=>setState(s=>({...s,displayName})),
    joinContest:(id:string,languages:string[])=>{clearScenario();setState(s=>({...s,viewer:s.viewer==='organizer'||s.viewer==='admin'?s.viewer:'participant',joinedContests:[...new Set([...s.joinedContests,id])],registrations:{...s.registrations,[id]:languages}}))},
    saveLog:(log:SampleLog)=>setState(s=>({...s,logs:s.logs.some(x=>x.id===log.id)?s.logs.map(x=>x.id===log.id?log:x):[log,...s.logs]})),
    removeLog:(id:string)=>setState(s=>({...s,logs:s.logs.filter(x=>x.id!==id)})),
    restoreLog:(log:SampleLog)=>setState(s=>({...s,logs:[log,...s.logs.filter(x=>x.id!==log.id)]})),
    saveContest:(contest:SampleContest)=>setState(s=>({...s,contests:s.contests.some(x=>x.id===contest.id)?s.contests.map(x=>x.id===contest.id?contest:x):[contest,...s.contests]})),
    reset:()=>{clearScenario();setState(initialState)},
  }
}
const PlaygroundContext=createContext<ReturnType<typeof useStore>|null>(null)
export function PlaygroundProvider({children}:{children:ReactNode}) { const value=useStore();return <PlaygroundContext.Provider value={value}>{children}</PlaygroundContext.Provider> }
export function usePlayground(){const value=useContext(PlaygroundContext);if(!value)throw new Error('PlaygroundProvider is required');return value}
export function useScenario(group: ScenarioGroup):[string,(value:string)=>void] {
  const [params,setParams]=useSearchParams();const {viewer,joinedContests}=usePlayground()
  const explicit=params.get('scenario')
  const fallback=group==='home'?(viewer==='guest'?'guest-live':joinedContests.includes('round5')?'participant-live':'member-open'):group==='leaderboard'?(viewer==='guest'?'guest':joinedContests.includes('round5')?'participant':'member'):group==='contests'?(viewer==='guest'?'guest':viewer==='organizer'?'creator':'member'):group==='log'?(viewer==='guest'?'visitor':'reading'):group==='profile'?'profile':''
  return [explicit&&scenarios[group].some(([id])=>id===explicit)?explicit:fallback,(value)=>{const next=new URLSearchParams(params);if(value)next.set('scenario',value);else next.delete('scenario');setParams(next,{replace:true})}]
}
