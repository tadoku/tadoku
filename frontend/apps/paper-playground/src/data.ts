export type Activity = 'Reading' | 'Listening'
export interface SampleUser { id: string; name: string; joined: string }
export interface SampleContest {
  id: string; title: string; scope: 'official' | 'community'; start: string; end: string
  registrationDeadline: string; description: string; languages: string[]; activities: Activity[]
  ownerId?: string; unlisted?: boolean
}
export interface Submission { contestId: string; score: number; basis: string; closed?: boolean }
export interface SampleLog {
  id: string; userId: string; title: string; language: string; activity: Activity
  amount: number; unit: 'pages' | 'minutes'; minutes?: number; date: string
  note: string; tags: string[]; media: string; submissions: Submission[]
}

export const sampleToday = '2026-09-05'
export const supportedLanguages = ['Japanese', 'Spanish', 'French', 'Korean', 'Chinese', 'German', 'Mandarin']
export const users: SampleUser[] = ['Kai', 'Mei', 'Ari', 'Lina', 'Noor', 'Sora', 'Elliot', 'Anton', 'Jun', 'Lea', 'Robin', 'Camille', 'Sam', 'Alex', 'Taylor'].map(name => ({ id: name.toLowerCase(), name, joined: name === 'Anton' ? '2023-01-12' : '2024-02-01' }))
export const leaderboardPeople: {userId: string; scores: [number, number, number, number]}[] = [
  {userId:'kai',scores:[760,190,320,130]}, {userId:'mei',scores:[580,230,320,70]},
  {userId:'ari',scores:[510,160,290,90]}, {userId:'lina',scores:[610,170,180,60]},
  {userId:'noor',scores:[360,170,210,90]}, {userId:'sora',scores:[550,110,80,20]},
  {userId:'elliot',scores:[210,160,270,120]}, {userId:'anton',scores:[370,80,110,40]},
  {userId:'jun',scores:[300,100,65,35]}, {userId:'lea',scores:[180,75,110,35]},
  {userId:'robin',scores:[110,55,145,20]}, {userId:'camille',scores:[80,55,100,45]},
  {userId:'sam',scores:[75,35,55,35]}, {userId:'alex',scores:[35,15,40,10]},
  {userId:'taylor',scores:[0,0,0,0]},
]
export const initialContests: SampleContest[] = [
  {id:'round5',title:'2026 Round 5',scope:'official',start:'2026-09-01',end:'2026-09-30',registrationDeadline:'2026-09-23',description:'A month of immersion, in the languages you are learning.',languages:['All languages'],activities:['Reading','Listening']},
  {id:'round6',title:'2026 Round 6',scope:'official',start:'2026-11-01',end:'2026-11-14',registrationDeadline:'2026-11-07',description:'The next official round. Choose your languages and make time for immersion.',languages:['All languages'],activities:['Reading','Listening']},
  {id:'round4',title:'2026 Round 4',scope:'official',start:'2026-07-01',end:'2026-07-31',registrationDeadline:'2026-07-24',description:'Browse the results from the July round.',languages:['All languages'],activities:['Reading','Listening']},
  {id:'round3',title:'2026 Round 3',scope:'official',start:'2026-05-01',end:'2026-05-14',registrationDeadline:'2026-05-07',description:'Two weeks of reading and listening together.',languages:['All languages'],activities:['Reading','Listening']},
  {id:'novels',title:'Japanese novel club',scope:'community',start:'2026-09-01',end:'2026-09-30',registrationDeadline:'2026-09-15',description:'Bring a novel and read a little each day.',languages:['Japanese'],activities:['Reading']},
  {id:'listening',title:'September listening hour',scope:'community',start:'2026-09-01',end:'2026-09-30',registrationDeadline:'2026-09-01',description:'A month of podcasts, audiobooks and conversations.',languages:['All languages'],activities:['Listening']},
  {id:'french',title:'Autumn in French',scope:'community',start:'2026-10-01',end:'2026-10-31',registrationDeadline:'2026-10-05',description:'Find something good to read or listen to in French.',languages:['French'],activities:['Reading','Listening'],ownerId:'anton'},
  {id:'summer',title:'Summer reading weekend',scope:'community',start:'2026-08-22',end:'2026-08-23',registrationDeadline:'2026-08-22',description:'The weekend is over. The reading stays with you.',languages:['All languages'],activities:['Reading']},
  {id:'friends',title:'Our Sunday reading group',scope:'community',start:'2026-09-06',end:'2026-09-27',registrationDeadline:'2026-09-13',description:'A shared reading challenge for the group. Accessible to anyone with the link.',languages:['Japanese','Korean'],activities:['Reading'],ownerId:'anton',unlisted:true},
  {id:'reading-circle',title:'September reading circle',scope:'community',start:'2026-09-01',end:'2026-09-30',registrationDeadline:'2026-09-15',description:'Make a little time for a book each day.',languages:['Japanese'],activities:['Reading']},
  {id:'listening-circle',title:'September listening circle',scope:'community',start:'2026-09-01',end:'2026-09-30',registrationDeadline:'2026-09-15',description:'Spend time listening together.',languages:['Japanese'],activities:['Listening']},
]

const readingLog: SampleLog = {id:'reading-konbini',userId:'anton',title:'コンビニ人間 — finished chapter 3',language:'Japanese',activity:'Reading',amount:42,unit:'pages',minutes:45,date:'2026-09-05',note:'The small routines in the shop say so much about Keiko. Finished chapter three on the train.',tags:['fiction','bookclub'],media:'Book',submissions:[{contestId:'round5',score:42,basis:'42 pages × 1 point per page = 42 points.'},{contestId:'reading-circle',score:22.5,basis:'45 tracked minutes × 0.5 points per minute = 22.5 points.'}]}
const listeningLog: SampleLog = {id:'listening-teppei',userId:'anton',title:'Nihongo con Teppei — episode 920',language:'Japanese',activity:'Listening',amount:35,unit:'minutes',date:'2026-09-04',note:'Listened on my morning walk. A few new words about everyday habits.',tags:['podcast'],media:'Podcast',submissions:[{contestId:'round5',score:17.5,basis:'35 minutes × 0.5 points per minute = 17.5 points.'},{contestId:'listening-circle',score:35,basis:'35 minutes × 1 point per minute = 35 points.'}]}
const works = [
  ['Le Petit Prince','French','Reading','Book'], ['InnerFrench','French','Listening','Podcast'],
  ['NHK Easy','Japanese','Reading','Article'], ['Midnight Diner','Japanese','Listening','Video'],
  ['Chants of Sennaar','Spanish','Reading','Game'], ['Radio Ambulante','Spanish','Listening','Podcast'],
] as const
const history: SampleLog[] = users.flatMap((user, userIndex) => [2023,2024,2025,2026].flatMap(year => Array.from({length:user.id==='anton'?48:8},(_,i) => {
  const work=works[(i+userIndex)%works.length]; const amount=12+(i*7+year+userIndex)%55
  const date=new Date(Date.UTC(year,0,1+(i*5+userIndex*3)%242)).toISOString().slice(0,10)
  return {id:`${user.id}-${year}-${i}`,userId:user.id,title:work[0],language:work[1],activity:work[2],amount,unit:work[2]==='Reading'?'pages':'minutes',date,note:i%4===0?'A little more each day.':'' ,tags:i%3===0?['daily']:[],media:work[3],submissions:[]} satisfies SampleLog
})))
export const initialLogs: SampleLog[] = [readingLog,listeningLog,
  {...readingLog,id:'reading-mei',userId:'mei',note:'An evening with a good book.'},
  {...readingLog,id:'personal-reading',title:'',minutes:undefined,tags:[],submissions:[]},
  {...readingLog,id:'ended-reading',amount:48,minutes:50,date:'2026-09-05',submissions:readingLog.submissions.map(s=>({...s,closed:true}))},
  ...history,
]

// These explicit sample rates exercise score explanations; they are not production scoring policy.
export function scoreLog(log: Pick<SampleLog,'activity'|'amount'>) { return log.activity === 'Reading' ? log.amount : log.amount * 0.5 }
export function scoreSubmission(log: Pick<SampleLog, 'activity'|'amount'|'unit'|'minutes'>, contestId: string): Submission {
  const timedReading = contestId === 'reading-circle'
  const amount = timedReading ? log.minutes : log.amount
  if (amount === undefined || !Number.isFinite(amount) || amount <= 0) throw new RangeError('A positive tracked amount is required for this sample contest rule.')
  const rate = timedReading ? 0.5 : contestId === 'listening-circle' ? 1 : log.activity === 'Reading' ? 1 : 0.5
  const unit = timedReading ? 'tracked minutes' : log.unit
  const score = amount * rate
  return { contestId, score, basis: `${formatNumber(amount)} ${unit} × ${rate} ${rate === 1 ? 'point' : 'points'} per ${timedReading || log.unit === 'minutes' ? 'minute' : 'page'} = ${formatNumber(score)} points. Sample contest rule.` }
}
export function formatNumber(value: number) { return new Intl.NumberFormat('en',{maximumFractionDigits:1}).format(value) }
export { formatDate } from './dates'
export function contestStatus(contest: SampleContest, asOf=sampleToday): 'live'|'upcoming'|'ended' { return asOf < contest.start ? 'upcoming' : asOf > contest.end ? 'ended' : 'live' }
