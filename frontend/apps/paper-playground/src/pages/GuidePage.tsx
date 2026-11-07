import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { accentRail, Button, buttonStyles, EmptyState, linkStyles, PageHeader, Table, usePaperForm } from 'paper-ui'
import { Link } from 'react-router-dom'
import { usePlayground } from '../state'
import { editorialStyles } from './editorial.styles'
import { utilities } from '../utilities.styles'
import { descendantStyles } from '../descendants.styles'

type GuideSection = 'about' | 'manual' | 'rules' | 'faq'

const routes: Record<GuideSection, string> = {
  about: '/about',
  manual: '/guide',
  rules: '/guide/scoring',
  faq: '/guide/questions',
}

const labels = {
  separate: { about: 'About', manual: 'Manual', rules: 'Rules', faq: 'FAQ' },
  guide: { about: 'About Tadoku', manual: 'Start here', rules: 'Scoring & rules', faq: 'Questions & help' },
  tasks: { about: 'About Tadoku', manual: 'Join & track', rules: 'Understand scores', faq: 'Get help' },
}

const headings: Record<GuideSection, { title: string; description: string }> = {
  about: {
    title: 'A little more immersion, together.',
    description: 'A community built around spending time with the languages you want to understand.',
  },
  manual: {
    title: 'Make your first entry.',
    description: 'Join a round, log what you do, and follow your progress.',
  },
  rules: {
    title: 'Fair play. Clear scores.',
    description: 'Understand the dates, eligibility and scoring context behind your participation.',
  },
  faq: {
    title: 'What would you like to know?',
    description: 'A few answers to keep you moving.',
  },
}

const questions = [
  {
    question: 'Can I join a contest after it has started?',
    answer: 'Yes, while its registration period remains open. Check the registration deadline on that contest’s page. A round can still be running after it stops accepting new participants.',
    href: '/contests',
    link: 'Find a contest',
  },
  {
    question: 'I can’t find my language. What should I do?',
    answer: 'Contact the Tadoku team to request it. The available language list does not include every language. Tell us the language’s name and, if you know it, its language code.',
    href: '/contact',
    link: 'Request a language',
  },
  {
    question: 'Do I have to enter a contest to track activity?',
    answer: 'No. You can keep a personal record throughout the year. Joining a contest is a separate way to take part, and you choose which eligible contests receive an activity.',
    href: '/logs/new',
    link: 'Log activity',
  },
  {
    question: 'Why does my log have more than one score?',
    answer: 'Your personal record and each contest can use different scoring rules. Log details shows those contexts separately, along with the amount and rule used for each score. These scores are not added together.',
    href: '/logs/reading-konbini',
    link: 'Inspect a sample score breakdown',
  },
  {
    question: 'Does a private contest hide everything from others?',
    answer: 'The setting sometimes called “private” means unlisted. The contest does not appear in public contest lists, but people with its link can access it. It is not a promise that only invited members can see the contest or its participants.',
    href: '/contests/friends',
    link: 'See an unlisted contest',
  },
  {
    question: 'How do I report a problem or suggest a feature?',
    answer: 'Contact the team or open an issue on GitHub. Describe what you expected and what happened. For a confusing score, include the relevant log and contest so the context is clear.',
    href: '/contact',
    link: 'Contact the team',
  },
]

function About() {
  return (
    <>
      <section {...stylex.props(editorialStyles.guideSection)} id="community">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>A book you want to finish. A voice you want to understand.</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Tadoku gives you a reason to come back tomorrow. Keep a personal record of your reading and listening, or join a friendly contest and follow the progress of other learners.</p>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Immersion can be a few pages on the train, an episode over lunch, or the book you finally feel ready to read. There is room for all of it here.</p>
        <Link {...stylex.props(...buttonStyles())} to="/contests">Explore contests</Link>
      </section>

      <section {...stylex.props(editorialStyles.guideSection)} id="history">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>From a reading contest to an immersion habit</h2>
        <dl {...stylex.props(editorialStyles.guideTimeline)}>
          <div {...stylex.props(descendantStyles.guideTimelineU_uDiv)}><dt {...stylex.props(descendantStyles.guideTimelineU_uDt)}>2010</dt><dd {...stylex.props(descendantStyles.guideTimelineU_uDd)}>LordSilent founded Tadoku as a reading contest on Twitter, with an online leaderboard.</dd></div>
          <div {...stylex.props(descendantStyles.guideTimelineU_uDiv)}><dt {...stylex.props(descendantStyles.guideTimelineU_uDt)}>2018</dt><dd {...stylex.props(descendantStyles.guideTimelineU_uDd)}>Changes to the Twitter API disrupted the original experience. The project later changed hands, and work began on a new website.</dd></div>
          <div {...stylex.props(descendantStyles.guideTimelineU_uDiv)}><dt {...stylex.props(descendantStyles.guideTimelineU_uDt)}>2023</dt><dd {...stylex.props(descendantStyles.guideTimelineU_uDd)}>Tadoku expanded into year-round tracking, activities beyond reading, and community-run contests.</dd></div>
        </dl>
      </section>

      <section {...stylex.props(editorialStyles.guideSection)} id="contribute">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Made by people who use it</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>antonve took over Tadoku after the contest became an important part of his own Japanese learning journey. The project continues as an open-source effort.</p>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Development, contest moderation and thoughtful feedback all help. Discuss a contribution with the team before starting a substantial change.</p>
        <div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_gap_4)}>
          <a {...stylex.props(linkStyles.base)} href="https://github.com/tadoku/tadoku">View the project on GitHub</a>
          <Link {...stylex.props(linkStyles.base)} to="/contact">Contact the team</Link>
        </div>
      </section>
    </>
  )
}

function Manual() {
  return (
    <>
      <p {...stylex.props(editorialStyles.guideIntroduction)}>You can log immersion anytime. Join a contest when you’d like a little company and a shared goal.</p>
      <ol {...stylex.props(editorialStyles.guideSteps)}>
        <li {...stylex.props(descendantStyles.guideStepsU_uLi)} id="participate">
          <h2 {...stylex.props(descendantStyles.guideStepsU_uH2)}>Choose how you want to participate</h2>
          <p {...stylex.props(descendantStyles.guideStepsU_uP)}>Track on your own, or choose an official or community contest. Check its dates and permitted activities. If registration is open, choose up to three languages and join.</p>
          <Link {...stylex.props(linkStyles.base)} to="/contests">Find a contest</Link>
        </li>
        <li {...stylex.props(descendantStyles.guideStepsU_uLi)} id="first-log">
          <h2 {...stylex.props(descendantStyles.guideStepsU_uH2)}>Read, listen, then log activity</h2>
          <p {...stylex.props(descendantStyles.guideStepsU_uP)}>Choose the language and activity, enter the amount or time requested by the form, and add a description or tags if useful. Check the score preview and contest submissions before saving.</p>
          <Link {...stylex.props(...buttonStyles())} to="/logs/new">Log activity</Link>
        </li>
        <li {...stylex.props(descendantStyles.guideStepsU_uLi)} id="score-context">
          <h2 {...stylex.props(descendantStyles.guideStepsU_uH2)}>Check where your activity counts</h2>
          <p {...stylex.props(descendantStyles.guideStepsU_uP)}>Your log details show its personal score. When you can view its contest submissions, each contribution is listed separately. Different scoring rules can produce different scores from the same activity.</p>
          <Link {...stylex.props(linkStyles.base)} to="/logs/reading-konbini">Explore a sample log</Link>
        </li>
      </ol>
      <aside {...stylex.props(editorialStyles.guideNote, accentRail.base)}>
        <h2 {...stylex.props(descendantStyles.guideNoteU_uH2)}>Already tracking and want to join a round?</h2>
        <p {...stylex.props(descendantStyles.guideNoteU_uP)}>Joining a contest and submitting a log are separate actions. Check the contest’s start, end and registration dates before assuming an activity can count.</p>
        <Link {...stylex.props(linkStyles.base)} to="/contests/round5">Check Round 5’s dates</Link>
      </aside>
      <section {...stylex.props(editorialStyles.guideSection)} aria-labelledby="guide-next">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)} id="guide-next">Find the detail you need</h2>
        <ul {...stylex.props(editorialStyles.guideRelated)}>
          <li><Link {...stylex.props(linkStyles.base)} to="/guide/scoring">Units and contest rules</Link></li>
          <li><Link {...stylex.props(linkStyles.base)} to="/guide/questions">Help with registration and languages</Link></li>
          <li><Link {...stylex.props(linkStyles.base)} to="/users/anton">Explore your profile</Link></li>
        </ul>
      </section>
    </>
  )
}

function Rules() {
  return (
    <>
      <p {...stylex.props(editorialStyles.guideIntroduction)}>The competition is friendly. The record should be honest. Every contest has its own dates, permitted activities and scoring rules; check that context before submitting.</p>
      <section {...stylex.props(editorialStyles.guideSection)} id="dates">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Contest times are local to you</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Contest start times and deadlines are shown in your device’s time zone. Activity dates stay on the day you recorded. The registration deadline and contest end date are different. A round may still be running after it stops accepting new participants. Activity must fall within the contest’s eligible period.</p>
        <Link {...stylex.props(linkStyles.base)} to="/contests/round5">See a contest’s dates and rules</Link>
      </section>
      <section {...stylex.props(editorialStyles.guideSection)} id="fair-participation">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Fair participation</h2>
        <ul {...stylex.props(editorialStyles.guideProseList)}>
          <li>Submit reading or listening you actually completed, as accurately as reasonably possible.</li>
          <li>Log regularly. Recording a little each day makes your progress easier to follow.</li>
          <li>Check repeat eligibility before submitting material you have read or listened to before.</li>
          <li>Use an appropriate display name and respect the other people taking part.</li>
        </ul>
      </section>
      <section {...stylex.props(editorialStyles.guideSection)} id="scores">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Understand the score you see</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Record reading in pages, sentences or characters, and listening in minutes. A log’s personal score and its contest scores are separate records. Never add them together to describe that log.</p>
        <Table
          caption="Personal score units in this playground"
          minWidth="0"
          columns={[
            { id: 'unit', header: 'Unit', rowHeader: true, cell: row => row.unit },
            { id: 'rate', header: 'Sample conversion', cell: row => row.rate },
          ]}
          rows={[
            { unit: 'Pages', rate: '1 point per page' },
            { unit: 'Sentences', rate: '1 point per 20 sentences' },
            { unit: 'Characters · Japanese, Korean, Chinese, Mandarin', rate: '1 point per 400 characters' },
            { unit: 'Characters · other languages', rate: 'About 1 point per 1,200 characters' },
            { unit: 'Listening minutes', rate: '0.5 points per minute' },
          ]}
        />
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>For pages, choose at most one modifier: Manga or Comic multiplies the score by 0.2; Two column by 1.6. Passive listening multiplies the listening score by 0.5. The reading and listening circle contests use their own fixed minute rates without these personal modifiers.</p>
        <Table
          caption="One record, two scoring contexts"
          captionVisibility="screen-reader"
          minWidth="0"
          columns={[
            { id: 'context', header: 'Context and example rule', rowHeader: true, cell: row => <span {...stylex.props(editorialStyles.guideScoreContext)}>{row.context}<small {...stylex.props(descendantStyles.guideScoreContextU_uSmall)}>{row.rule}</small></span> },
            { id: 'score', header: 'Score', align: 'end', cell: row => row.score },
          ]}
          rows={[
            { context: 'Personal record', rule: '20 pages × 1 point', score: '20' },
            { context: 'A community contest', rule: '20 pages × 0.5 points', score: '10' },
          ]}
        />
        <p {...stylex.props(editorialStyles.guideFineprint, descendantStyles.guideSectionU_uP, descendantStyles.guideSectionU_uGuideFineprint)}>These rates are illustrative. The rule saved with your log or contest submission determines its score.</p>
        <Link {...stylex.props(linkStyles.base)} to="/logs/reading-konbini">See the breakdown in log details</Link>
      </section>
      <section {...stylex.props(editorialStyles.guideSection)} id="eligibility">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Eligibility and repeats</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Official rounds include reading and listening. A community contest may accept a smaller set of activities or languages. Books, manga, games, podcasts and video describe the material; choose the activity you actually did.</p>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>For repeats or changes to playback speed, consult the applicable contest guidance before submitting. Check the score preview; do not apply a second manual adjustment when a scoring rule already accounts for it.</p>
      </section>
      <section {...stylex.props(editorialStyles.guideSection)} id="questions">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>When something looks wrong</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>If a submission seems incorrect, contact the team with the log and contest links. Include the amount recorded and the score you expected so the context is clear.</p>
        <Link {...stylex.props(linkStyles.base)} to="/contact">Ask about a rule</Link>
      </section>
    </>
  )
}

function Questions() {
  const form = usePaperForm({ defaultValues: { query: '' }, onSubmit: () => undefined })
  const query = useStore(form.store, state => state.values['query']).trim().toLocaleLowerCase()
  const visible = questions.filter(item => `${item.question} ${item.answer}`.toLocaleLowerCase().includes(query))
  const clear = () => {
    form.setFieldValue('query', '')
    document.getElementsByName('query')[0]?.focus()
  }

  return (
    <>
      <form.AppForm>
        <form.Form role="search" xstyle={[editorialStyles.guideSearch]}>
          <form.AppField name="query">{field => <field.Input label="Search these questions" type="search" autoComplete="off" aria-describedby="guide-search-results" placeholder="Try registration, language or score" />}</form.AppField>
          <div {...stylex.props(editorialStyles.guideSearchMeta)}>
            <p {...stylex.props(descendantStyles.guideSearchMetaU_uP)} id="guide-search-results" role="status">{visible.length} {visible.length === 1 ? 'question' : 'questions'}</p>
            {query && <Button variant="link" onClick={clear}>Clear search</Button>}
          </div>
        </form.Form>
      </form.AppForm>
      <div {...stylex.props(editorialStyles.guideQuestions)}>
        {visible.map(item => (
          <details {...stylex.props(descendantStyles.guideQuestionsU_uDetails)} key={item.question}>
            <summary {...stylex.props(descendantStyles.guideQuestionsU_uSummary)}>{item.question}</summary>
            <div {...stylex.props(editorialStyles.guideAnswer)}>
              <p {...stylex.props(descendantStyles.guideAnswerU_uP)}>{item.answer}</p>
              <Link {...stylex.props(linkStyles.base)} to={item.href}>{item.link}</Link>
            </div>
          </details>
        ))}
      </div>
      {visible.length === 0 && (
        <EmptyState headingLevel={2} title="No questions match that search." description="Try a broader term or clear the search." />
      )}
      <section {...stylex.props(editorialStyles.guideSection)} id="contact">
        <h2 {...stylex.props(descendantStyles.guideSectionU_uH2)}>Still looking for an answer?</h2>
        <p {...stylex.props(descendantStyles.guideSectionU_uP)}>Tell us what you were trying to do and where you got stuck.</p>
        <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to="/contact">Contact Tadoku</Link>
      </section>
    </>
  )
}

export function GuidePage({ page }: { page: GuideSection }) {
  const { ia } = usePlayground()
  const currentLabels = labels[ia]
  const pageOrder: GuideSection[] = ia === 'separate' ? ['about', 'manual', 'rules', 'faq'] : ia === 'tasks' ? ['manual', 'rules', 'faq'] : ['manual', 'rules', 'faq', 'about']
  const inPageLinks: Partial<Record<GuideSection, { id: string; label: string }[]>> = {
    about: [{ id: 'community', label: 'The community' }, { id: 'history', label: 'Our history' }, { id: 'contribute', label: 'Contribute' }],
    manual: [{ id: 'participate', label: 'Join a contest' }, { id: 'first-log', label: 'Log activity' }, { id: 'score-context', label: 'Follow your progress' }],
    rules: [{ id: 'dates', label: 'Dates' }, { id: 'fair-participation', label: 'Fair participation' }, { id: 'scores', label: 'Scores' }, { id: 'eligibility', label: 'Eligibility' }],
  }

  return (
    <div {...stylex.props(editorialStyles.guideLayout)}>
      <nav {...stylex.props(editorialStyles.guideNavigation)} aria-label={ia === 'separate' ? 'Learn' : ia === 'guide' ? 'Guide' : 'Help'}>
        {pageOrder.map(key => <Link key={key} to={routes[key]} {...stylex.props(linkStyles.base)} aria-current={key === page ? 'page' : undefined}>{currentLabels[key]}</Link>)}
      </nav>
      <article {...stylex.props(editorialStyles.guideArticle)}>
        <PageHeader xstyle={[editorialStyles.guideHeader]} title={headings[page].title} lead={headings[page].description}>
          {inPageLinks[page] && <nav {...stylex.props(editorialStyles.guideContents)} aria-label="On this page">{inPageLinks[page]?.map(link => <a key={link.id} href={`#${link.id}`} {...stylex.props(linkStyles.base, descendantStyles.guideContentsU_uA)}>{link.label}</a>)}</nav>}
        </PageHeader>
        {page === 'about' && <About />}
        {page === 'manual' && <Manual />}
        {page === 'rules' && <Rules />}
        {page === 'faq' && <Questions />}
      </article>
    </div>
  )
}
