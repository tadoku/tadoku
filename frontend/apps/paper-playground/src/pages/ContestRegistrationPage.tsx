import { Link, useNavigate, useParams } from 'react-router-dom'
import { usePaperForm, Button, EmptyState, PageHeader, buttonClassName } from 'paper-ui'
import { contestStatus, supportedLanguages } from '../data'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import { ContestFacts, ContestNotFound, ContestPrimaryAction } from './contestParts'
import { contestPath, registrationIsOpen, useContestDate } from './contestRound'
import './contests.css'

export function ContestRegistrationPage() {
  const { contestId } = useParams()
  const { contests, viewer, registrations, joinContest } = usePlayground()
  const navigate = useNavigate()
  const asOf = useContestDate()
  const contest = contests.find(item => item.id === contestId)
  const methods = usePaperForm({ defaultValues: ({ languages: registrations[contestId ?? ''] ?? [] } as { languages: string[] }), onSubmit: ({ value: values }) => { if (!contest) return; joinContest(contest.id, values.languages); navigate(`${contestPath(contest, asOf)}&registered=1`) } })
  if (!contest) return <ContestNotFound />
  const allowed = contest.languages.includes('All languages') ? supportedLanguages : contest.languages

  return <>
    <PageHeader backLink={{ href: contestPath(contest, asOf), label: `Back to ${contest.title}` }} renderLink={renderLink} title={`Join ${contest.title}`} lead="Choose the languages you’ll read or listen in. You can register up to three." />
    <div className="page-columns contest-registration-layout"><section>
      {!registrationIsOpen(contest, asOf) ? <EmptyState headingLevel={2} title="Registration is closed." description={contestStatus(contest, asOf) === 'live' ? 'This round is still running. Participants who already registered can continue logging.' : 'Browse another round to join, or explore the results from this one.'} actions={<ContestPrimaryAction contest={contest} asOf={asOf} />} /> : viewer === 'guest' ? <EmptyState headingLevel={2} title="Sign in to choose your languages." description="You’ll return to this registration after signing in." actions={<Link className={buttonClassName()} to={`/sign-in?next=${encodeURIComponent(contestPath(contest, asOf, '/registration'))}`}>Sign in to join</Link>} /> : <methods.AppForm><methods.Form className="app-form"><div className="app-form__fields"><methods.AppField name="languages" validators={{ onChange: ({ value }) => { if (!value.length) return "Choose at least one option."; const result = (value => Array.isArray(value) && value.length >= 1 && value.length <= 3 && value.every(language => allowed.includes(language)) || 'Choose one to three of the available languages.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.AutocompleteMultiInput label="Your languages" hint="Choose one to three languages. Remove a language to make room for another." options={allowed} format={value => value} getId={value => value} required maxSelections={3} />}</methods.AppField><p className="muted">Allowed activities: {contest.activities.join(' and ')}. After joining, choose which eligible logs to submit; joining does not submit past activity automatically.</p></div><div className="app-form__actions"><Button type="submit">Save registration</Button><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf)}>Cancel</Link></div></methods.Form></methods.AppForm>}
    </section><aside className="contest-detail-timeline"><h2 className="paper-type-component">Your round</h2><ContestFacts contest={contest} asOf={asOf} /></aside></div>
  </>
}
