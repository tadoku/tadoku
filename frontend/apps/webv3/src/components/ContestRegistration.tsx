import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { useHydrated, useNavigate } from '@tanstack/react-router'
import type { ImmersionContestView, ImmersionLanguage } from 'api-client'
import { Button, buttonClassName, EmptyState, Flash, formFooter, PageHeader, Surface, typography, usePaperForm } from 'paper-ui'
import { useState } from 'react'
import { contestBoardQuery, contestPhase, readDay, type ContestPhase } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import {
  contestLanguagesQuery,
  registrationQuery,
  useSaveRegistration,
  type RegistrationFailure,
} from '../lib/registration'
import { useLoginUrl, viewerKey, viewerQueryOptions } from '../lib/session'
import { ContestFacts } from './ContestLeaderboard'
import { ContestNotFound } from './ContestPage'
import { renderLink } from './router-link'
import { description, recordSurface, standingsAside, standingsHeader, workspace } from './standings-styles'
import { formControls } from './form-styles'

/** Joining a contest or changing the languages of a registration, for the signed-in viewer. */
export function ContestRegistrationPage({ contestId }: { contestId: string }) {
  const viewerQuery = useQuery(viewerQueryOptions())
  const viewer = viewerQuery.data
  const board = useQuery(contestBoardQuery(contestId, viewerKey(viewer)))
  const loginUrl = useLoginUrl()

  switch (viewer?.status) {
    case 'signed-in':
      break
    case 'anonymous':
      // The route sends anonymous visitors to log in; this covers a session that ends while the page is open.
      return (
        <EmptyState
          headingLevel={1}
          title="Log in to join this contest"
          description="Choose your languages after logging in."
          actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
        />
      )
    case 'banned':
      return <EmptyState headingLevel={1} title="Your account is suspended" description="This account cannot join contests." />
    default:
      return (
        <EmptyState
          headingLevel={1}
          title="Account service unavailable"
          description="We could not check your session. Try again."
          actions={<Button onClick={() => void viewerQuery.refetch()}>Try again</Button>}
        />
      )
  }

  if (board.data?.kind === 'none') return <ContestNotFound />
  if (board.data?.kind !== 'board') {
    return (
      <EmptyState
        headingLevel={1}
        title="This contest could not be loaded"
        description="Try loading the page again in a moment."
        actions={<Button variant="outline" loading={board.isFetching} onClick={() => void board.refetch()}>Try again</Button>}
      />
    )
  }
  const { contest } = board.data.board
  return <Registration key={viewer.userId} contest={contest} viewerId={viewer.userId} phase={contestPhase(contest, readDay(board.dataUpdatedAt))} />
}

function Registration({ contest, viewerId, phase }: { contest: ImmersionContestView & { id: string }; viewerId: string; phase: ContestPhase }) {
  const links = useLinks()
  const registration = useQuery(registrationQuery(contest.id, viewerId))
  const registered = registration.data?.kind === 'registered' ? registration.data.languages : null

  return (
    <div className="standings-page">
      <PageHeader
        className={standingsHeader}
        backLink={{ href: links.contest(contest.id), label: contest.title }}
        renderLink={renderLink}
        title={registered ? 'Update registration' : `Join ${contest.title}`}
      />
      <div className={workspace}>
        <div>
          <RegistrationState contest={contest} viewerId={viewerId} phase={phase} />
        </div>
        <aside className={standingsAside}>
          <section aria-labelledby="registration-round">
            <h2 id="registration-round" className={typography({ role: 'component' })}>About this contest</h2>
            {contest.description ? <p className={description}>{contest.description}</p> : null}
            <ContestFacts contest={contest} phase={phase} />
          </section>
        </aside>
      </div>
    </div>
  )
}

/** The form, or why it isn't offered: the registration can't be read, or the deadline Webv2 enforces has passed. */
function RegistrationState({ contest, viewerId, phase }: { contest: ImmersionContestView & { id: string }; viewerId: string; phase: ContestPhase }) {
  const links = useLinks()
  const registration = useQuery(registrationQuery(contest.id, viewerId))
  const restricted = Boolean(contest.allowed_languages?.length)
  const allLanguages = useQuery({ ...contestLanguagesQuery(), enabled: !restricted })

  if (registration.data?.kind === 'unavailable' || registration.isError) {
    return (
      <Flash variant="danger" action={<Button variant="outline" loading={registration.isFetching} onClick={() => void registration.refetch()}>Try again</Button>}>
        Could not load your registration, please try again.
      </Flash>
    )
  }
  if (!registration.data) return <p role="status">Loading your registration…</p>
  const previous = registration.data.kind === 'registered' ? registration.data.languages : null

  // As in Webv2: registered participants can change languages until the contest ends; others can join until registration closes.
  const closed = phase.ended
    ? previous
      ? 'This contest has ended, so your registration can no longer be changed.'
      : 'This contest has ended, so it no longer accepts registrations.'
    : !previous && !phase.registrationOpen
      ? 'Unfortunately, registrations for this contest have ended.'
      : null
  if (closed) {
    return (
      <>
        <Flash variant="danger">{closed}</Flash>
        {previous ? <p className="my-4">Your languages: {previous.map(language => language.name).join(', ')}.</p> : null}
        <p className="my-4"><a className="text-link" href={links.contest(contest.id)}>Back to the contest</a></p>
      </>
    )
  }

  if (restricted) return <RegistrationForm contest={contest} viewerId={viewerId} previous={previous} options={contest.allowed_languages!} />
  if (allLanguages.data?.ok) return <RegistrationForm contest={contest} viewerId={viewerId} previous={previous} options={allLanguages.data.data} />
  if (!allLanguages.data && allLanguages.isFetching) return <p role="status">Loading languages…</p>
  return (
    <Flash variant="danger" action={<Button variant="outline" loading={allLanguages.isFetching} onClick={() => void allLanguages.refetch()}>Try again</Button>}>
      Could not load contest configuration, please try again later.
    </Flash>
  )
}

function validateLanguages({ value }: { value: ImmersionLanguage[] }) {
  if (value.length === 0) return 'Select at least one language.'
  if (value.length > 3) return 'Select up to three languages.'
  if (new Set(value.map(language => language.code)).size !== value.length) return 'Select each language only once.'
  return undefined
}

function RegistrationForm({ contest, viewerId, previous, options }: {
  contest: ImmersionContestView & { id: string }
  viewerId: string
  /** The registered languages, or null when the viewer hasn't joined. */
  previous: ImmersionLanguage[] | null
  options: ImmersionLanguage[]
}) {
  const links = useLinks()
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const saveRegistration = useSaveRegistration()
  const [failure, setFailure] = useState<RegistrationFailure | null>(null)
  const form = usePaperForm({
    defaultValues: { languages: previous ?? [] },
    onSubmit: async ({ value }) => {
      setFailure(null)
      const result = await saveRegistration({
        contestId: contest.id,
        viewerId,
        languageCodes: value.languages.map(language => language.code),
      })
      if (result.ok) await navigate({ href: links.contest(contest.id), replace: true })
      else setFailure(result.reason)
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const selected = useStore(form.store, state => state.values.languages)
  const codes = new Set(selected.map(language => language.code))
  const removed = (previous ?? []).filter(language => !codes.has(language.code))
  const disabled = !hydrated || pending

  return (
    <Surface className={`${recordSurface} max-w-none`}>
      <form.AppForm>
        <form.Form className="min-w-0" aria-busy={pending}>
          <fieldset className={formControls} disabled={disabled}>
            <form.AppField name="languages" validators={{ onChange: validateLanguages, onSubmit: validateLanguages }}>
              {field => (
                <field.AutocompleteMultiInput
                  label="Languages (up to 3)"
                  required
                  maxSelections={3}
                  options={options}
                  format={language => language.name}
                  getId={language => language.code}
                  placeholder="Add a language"
                  disabled={disabled}
                />
              )}
            </form.AppField>
            {removed.length > 0 ? (
              <Flash variant="warning">
                Removing {removed.map(language => language.name).join(', ')} will also remove any contest logs you’ve
                submitted in {removed.length === 1 ? 'that language' : 'those languages'}.
              </Flash>
            ) : null}
          </fieldset>
          {failure ? <SaveFailure failure={failure} registered={previous !== null} /> : null}
          <div className={formFooter().root()}>
            <div className={formFooter().actions()}>
              <Button type="submit" disabled={disabled}>{pending ? 'Saving…' : previous ? 'Update registration' : 'Register'}</Button>
              <a className="text-link" href={links.contest(contest.id)}>Cancel</a>
            </div>
          </div>
        </form.Form>
      </form.AppForm>
    </Surface>
  )
}

function failureMessage(failure: RegistrationFailure, registered: boolean) {
  switch (failure) {
    case 'closed': return registered ? 'This registration can no longer be changed.' : 'Registration for this contest has closed.'
    case 'signed-out': return 'Your session ended, so nothing was saved. Log in and try again.'
    case 'switched': return 'You are now signed in as a different account, so nothing was saved. Reload the page to register that account.'
    case 'account-unavailable': return 'We could not check your session, so nothing was saved. Try again in a moment.'
    case 'suspended': return 'Your account is suspended and cannot join contests.'
    case 'invalid': return 'This contest doesn’t accept that selection of languages. Check them and try again.'
    case 'not-found': return 'This contest no longer exists.'
    case 'deleting': return 'Your account is being deleted, so it cannot join contests.'
    case 'unavailable': return 'Could not save your registration, please try again later.'
  }
}

function SaveFailure({ failure, registered }: { failure: RegistrationFailure; registered: boolean }) {
  const loginUrl = useLoginUrl()
  return (
    <Flash variant="danger" action={failure === 'signed-out' ? <a className="text-link" href={loginUrl}>Log in</a> : undefined}>
      {failureMessage(failure, registered)}
    </Flash>
  )
}
