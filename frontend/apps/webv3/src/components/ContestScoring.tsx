import { useQuery } from '@tanstack/react-query'
import { useHydrated } from '@tanstack/react-router'
import type { ImmersionLogConfigurationOptions, ImmersionScoringRule, ImmersionScoringRuleSet } from 'api-client'
import { Button, EmptyState, Flash, Modal, PageHeader, SectionHeading, StatusTag, Surface, buttonClassName } from 'paper-ui'
import { useRef, useState } from 'react'
import { contestsHref } from '../lib/contests'
import { useLinks } from '../lib/links'
import type { LogFailure } from '../lib/logs'
import { contestScoringQuery, useChangeScoring, type RuleSetChange } from '../lib/scoring'
import { useLoginUrl } from '../lib/session'
import { renderLink } from './router-link'
import { RuleSetForm } from './RuleSetForm'
import { scoringFailures } from './scoring-failures'

/**
 * A contest's scoring rule-set versions for an administrator, who can draft, publish and activate versions until the
 * contest's first UTC day.
 */
export function ContestScoring({ contestId, viewerId }: { contestId: string; viewerId: string }) {
  const links = useLinks()
  const loginUrl = useLoginUrl()
  const scoringQuery = useQuery(contestScoringQuery(contestId, viewerId))
  const result = scoringQuery.data

  if (!result?.ok) {
    switch (result?.reason) {
      case 'not-found':
        return (
          <EmptyState
            headingLevel={1}
            title="Contest not found"
            description="This contest does not exist or has been removed."
            actions={<a className={buttonClassName({ variant: 'outline' })} href={contestsHref.official()}>Browse contests</a>}
          />
        )
      case 'signed-out':
        return (
          <EmptyState
            headingLevel={1}
            title="Your session ended"
            description="Log in again to manage contest scoring."
            actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
          />
        )
      default:
        return (
          <EmptyState
            headingLevel={1}
            title="Scoring rules unavailable"
            description="Could not load scoring rules, please try again later."
            actions={(
              <Button loading={scoringQuery.isFetching} onClick={() => void scoringQuery.refetch()}>Try again</Button>
            )}
          />
        )
    }
  }

  const { contest, options, platformRuleSets, contestRuleSets } = result.data
  const locked = new Date().toISOString().slice(0, 10) >= contest.contest_start

  return (
    <div className="grid min-w-0 max-w-record gap-6">
      <PageHeader
        title="Contest scoring"
        className="mb-0"
        backLink={{ href: links.contest(contestId), label: contest.title }}
        renderLink={renderLink}
      />
      {locked ? (
        <Flash variant="information">
          Scoring is locked because this contest has started.{contestRuleSets.length > 0
            ? ' Its rule-set versions are listed below.'
            : <> It uses the <a className="text-link hover:text-link-hover" href={`${links.guide}#media`}>platform defaults</a>.</>}
        </Flash>
      ) : (
        <RuleSetForm
          contestId={contestId}
          viewerId={viewerId}
          options={options}
          platformRuleSets={platformRuleSets}
        />
      )}
      {contestRuleSets.length === 0 && !locked ? <p className="m-0">No contest-specific rule sets yet; this contest scores with the platform defaults.</p> : null}
      {locked && contestRuleSets.length === 0 ? <section className="grid min-w-0 gap-4 [&>p]:m-0" aria-labelledby="scoring-platform">
        <SectionHeading level={2} id="scoring-platform" title="Rules in effect" />
        {platformRuleSets.filter(ruleSet => ruleSet.active).map(ruleSet => <RuleSetVersion key={ruleSet.id} ruleSet={ruleSet} options={options} change={null} />)}
      </section> : null}
      {contestRuleSets.length === 0 ? null : <section className="grid min-w-0 gap-4 [&>p]:m-0" aria-labelledby="scoring-versions">
        <SectionHeading level={2} id="scoring-versions" title="Rule-set versions" />
        {contestRuleSets.map(ruleSet => (
          <RuleSetVersion
            key={ruleSet.id}
            ruleSet={ruleSet}
            options={options}
            change={locked ? null : { viewerId, contestId, ruleSetId: ruleSet.id }}
          />
        ))}
      </section>}
    </div>
  )
}

/** One version, its rules in priority order, and the next step it allows unless scoring is locked. */
function RuleSetVersion({ ruleSet, options, change }: {
  ruleSet: ImmersionScoringRuleSet
  options: ImmersionLogConfigurationOptions
  change: Omit<RuleSetChange, 'action'> | null
}) {
  const action = change && (ruleSet.status === 'draft' ? 'publish' : ruleSet.active ? null : 'activate')

  return (
    <Surface as="article" className="scoring-version grid min-w-0 gap-3 p-version wrap-anywhere" aria-labelledby={`version-${ruleSet.id}`}>
      <SectionHeading
        level={3}
        id={`version-${ruleSet.id}`}
        title={`Version ${ruleSet.version}`}
        status={(
          <>
            <StatusTag tone={ruleSet.status === 'draft' ? 'neutral' : 'information'}>
              {ruleSet.status === 'draft' ? 'Draft' : 'Published'}
            </StatusTag>
            {ruleSet.active ? <StatusTag tone="success">Active</StatusTag> : null}
          </>
        )}
        description={ruleSet.scope === 'platform' ? 'Platform defaults' : ruleSet.mode === 'replace'
          ? 'Replaces platform rules; uncovered inputs score zero'
          : 'Overrides platform rules with a pinned fallback'}
        actions={change && action ? <ConfirmChange change={{ ...change, action }} /> : null}
      />
      {ruleSet.rules.length ? <RuleTables rules={ruleSet.rules} options={options} /> : (
        <p>No rules: every input scores zero.</p>
      )}
    </Surface>
  )
}

/** A version's rules by activity: rates by unit, then modifiers as multipliers, each row a condition and its rate. */
function RuleTables({ rules, options }: { rules: ImmersionScoringRule[]; options: ImmersionLogConfigurationOptions }) {
  const activities = [...new Set(rules.map(rule => rule.activity_id))]

  return (
    <div className="grid gap-6">
      {activities.map(activityId => {
        const activity = options.activities.find(activity => activity.id === activityId)?.name ?? `Activity ${activityId}`
        const group = rules.filter(rule => rule.activity_id === activityId)
        const bases = group.filter(rule => !rule.stackable)
        const units = [...new Set(bases.map(rule => rule.unit_key ?? (rule.score_source === 'duration_minutes' ? 'minute' : '')))]
        return (
          <section key={activityId} aria-label={`${activity} rules`} className="grid gap-3">
            <h4 className="m-0 font-body text-base font-semibold">{activity}</h4>
            {units.map(unit => (
              <RateRows
                key={unit}
                title={unit === 'minute' ? 'Per minute' : `Per ${unitName(unit, options).toLowerCase()}`}
                rules={bases.filter(rule => (rule.unit_key ?? (rule.score_source === 'duration_minutes' ? 'minute' : '')) === unit)}
                options={options}
                rate={rule => rateFormat.format(rule.rate)}
              />
            ))}
            {group.some(rule => rule.stackable)
              ? <RateRows title="Modifiers" rules={group.filter(rule => rule.stackable)} options={options} rate={rule => `× ${rateFormat.format(rule.rate)}`} />
              : null}
          </section>
        )
      })}
    </div>
  )
}

function RateRows({ title, rules, options, rate }: {
  title: string
  rules: ImmersionScoringRule[]
  options: ImmersionLogConfigurationOptions
  rate: (rule: ImmersionScoringRule) => string
}) {
  return (
    <div>
      <h5 className="mb-1 mt-0 text-metadata font-semibold text-muted">{title}</h5>
      <dl className="m-0">
        {rules.map(rule => (
          <div key={rule.id ?? rule.priority} className="flex justify-between gap-4 border-b border-rule py-1.5 text-metadata tabular-nums last:border-b-0">
            <dt>{condition(rule, options)}</dt>
            <dd className="m-0 whitespace-nowrap font-semibold">{rate(rule)}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

const rateFormat = new Intl.NumberFormat('en-GB', { maximumSignificantDigits: 3 })

const unitName = (unitKey: string, options: ImmersionLogConfigurationOptions) =>
  options.units.find(unit => unit.unit_key === unitKey)?.name ?? unitKey.replaceAll('_', ' ')

function condition(rule: ImmersionScoringRule, options: ImmersionLogConfigurationOptions) {
  const language = options.languages.find(language => language.code === rule.language_code)?.name ?? rule.language_code
  return [language, rule.tag?.replaceAll('_', ' ')].filter(Boolean).join(' · ') || 'All languages'
}

const confirmations = {
  publish: { label: 'Publish', pending: 'Publishing…', title: 'Publish this version?', body: 'Published rules are immutable.' },
  activate: {
    label: 'Activate',
    pending: 'Activating…',
    title: 'Activate this version?',
    body: 'This version will score new contest submissions.',
  },
}

/** Publishes or activates a version after confirmation. */
function ConfirmChange({ change }: { change: RuleSetChange }) {
  const hydrated = useHydrated()
  const changeScoring = useChangeScoring()
  const [open, setOpen] = useState(false)
  const [pending, setPending] = useState(false)
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const cancelRef = useRef<HTMLButtonElement>(null)
  const copy = confirmations[change.action]

  const confirm = async () => {
    setPending(true)
    setFailure(null)
    const result = await changeScoring({ update: change })
    setPending(false)
    if (result.ok) setOpen(false)
    else setFailure(result.reason)
  }

  return (
    <Modal
      trigger={<Button variant={change.action === 'publish' ? 'outline' : 'default'} disabled={!hydrated}>{copy.label}</Button>}
      title={copy.title}
      initialFocus={cancelRef}
      open={open}
      onOpenChange={next => {
        if (pending) return
        setOpen(next)
        setFailure(null)
      }}
      footer={(
        <>
          <Button ref={cancelRef} variant="outline" disabled={pending} onClick={() => setOpen(false)}>Cancel</Button>
          <Button disabled={pending} onClick={() => void confirm()}>{pending ? copy.pending : copy.label}</Button>
        </>
      )}
    >
      <p>{copy.body}</p>
      {failure ? <p role="alert">{scoringFailures(failure, 'Nothing changed.')}</p> : null}
    </Modal>
  )
}
