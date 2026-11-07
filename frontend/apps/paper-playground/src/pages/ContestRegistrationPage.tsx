import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button, buttonClassName, EmptyState, PageHeader, typography, usePaperForm } from 'paper-ui'
import { contestStatus, supportedLanguages } from '../data'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import { ContestFacts, ContestNotFound, ContestPrimaryAction } from './contestParts'
import { contestPath, registrationIsOpen, useContestDate } from './contestRound'

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
    <div className="grid grid-cols-main-aside items-start gap-9 *:min-w-0 max-tablet:grid-cols-single max-tablet:gap-7"><section>
      {!registrationIsOpen(contest, asOf) ? <EmptyState headingLevel={2} title="Registration is closed." description={contestStatus(contest, asOf) === 'live' ? 'This round is still running. Participants who already registered can continue logging.' : 'Browse another round to join, or explore the results from this one.'} actions={<ContestPrimaryAction contest={contest} asOf={asOf} />} /> : viewer === 'guest' ? <EmptyState headingLevel={2} title="Sign in to choose your languages." description="You’ll return to this registration after signing in." actions={<Link className={buttonClassName()} to={`/sign-in?next=${encodeURIComponent(contestPath(contest, asOf, '/registration'))}`}>Sign in to join</Link>} /> : <methods.AppForm><methods.Form className="w-full min-w-0 max-w-form-page"><div className="grid gap-5 [&>p]:m-0"><methods.AppField name="languages" validators={{ onChange: ({ value }) => { if (!value.length) return "Choose at least one option."; const result = (value => Array.isArray(value) && value.length >= 1 && value.length <= 3 && value.every(language => allowed.includes(language)) || 'Choose one to three of the available languages.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.AutocompleteMultiInput label="Your languages" hint="Choose one to three languages. Remove a language to make room for another." options={allowed} format={value => value} getId={value => value} required maxSelections={3} />}</methods.AppField><p className="text-muted">Allowed activities: {contest.activities.join(' and ')}. After joining, choose which eligible logs to submit; joining does not submit past activity automatically.</p></div><div className="mt-7 flex flex-wrap gap-3"><Button type="submit">Save registration</Button><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf)}>Cancel</Link></div></methods.Form></methods.AppForm>}
    </section><aside className="border border-rule bg-paper p-6 [&>h2]:mb-5 [&>h2]:mt-0 [&>p]:m-0 [&>p]:text-fine [&>p]:leading-165 [&>:last-child]:mb-0"><h2 className={typography({ role: 'component' })}>Your round</h2><ContestFacts contest={contest} asOf={asOf} /></aside></div>
  </>
}
