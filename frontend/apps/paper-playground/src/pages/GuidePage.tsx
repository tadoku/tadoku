import { useStore } from '@tanstack/react-form'
import { accentRail, Button, buttonClassName, cn, EmptyState, PageHeader, Table, usePaperForm } from 'paper-ui'
import { Link } from 'react-router-dom'
import { usePlayground } from '../state'

const guideSection = 'mt-10 scroll-mt-6 max-tablet:mt-8 [.guide-header+&]:mt-0 [&>h2]:mb-3 [&>h2]:mt-0 [&>h2]:font-display [&>h2]:text-xl [&>h2]:leading-150 [&_p]:mb-4 [&_p]:mt-0 [&_p]:max-w-measure-prose [&_p]:leading-180 [&>p:last-child]:mb-0 [&_.guide-fineprint]:mt-3 [&_.guide-fineprint]:text-fine [&_.guide-fineprint]:leading-170 [&_.guide-fineprint]:text-muted'

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
      <section className={guideSection} id="community">
        <h2>A book you want to finish. A voice you want to understand.</h2>
        <p>Tadoku gives you a reason to come back tomorrow. Keep a personal record of your reading and listening, or join a friendly contest and follow the progress of other learners.</p>
        <p>Immersion can be a few pages on the train, an episode over lunch, or the book you finally feel ready to read. There is room for all of it here.</p>
        <Link className={buttonClassName()} to="/contests">Explore contests</Link>
      </section>

      <section className={guideSection} id="history">
        <h2>From a reading contest to an immersion habit</h2>
        <dl className="mb-0 mt-5 [&>div]:grid [&>div]:grid-cols-timeline [&>div]:gap-5 [&>div]:border-t [&>div]:border-rule [&>div]:py-4.5 max-xs:[&>div]:gap-3 [&_dt]:font-bold [&_dt]:tabular-nums [&_dd]:m-0 [&_dd]:leading-180">
          <div><dt>2010</dt><dd>LordSilent founded Tadoku as a reading contest on Twitter, with an online leaderboard.</dd></div>
          <div><dt>2018</dt><dd>Changes to the Twitter API disrupted the original experience. The project later changed hands, and work began on a new website.</dd></div>
          <div><dt>2023</dt><dd>Tadoku expanded into year-round tracking, activities beyond reading, and community-run contests.</dd></div>
        </dl>
      </section>

      <section className={guideSection} id="contribute">
        <h2>Made by people who use it</h2>
        <p>antonve took over Tadoku after the contest became an important part of his own Japanese learning journey. The project continues as an open-source effort.</p>
        <p>Development, contest moderation and thoughtful feedback all help. Discuss a contribution with the team before starting a substantial change.</p>
        <div className="flex flex-wrap gap-4">
          <a className="text-link" href="https://github.com/tadoku/tadoku">View the project on GitHub</a>
          <Link className="text-link" to="/contact">Contact the team</Link>
        </div>
      </section>
    </>
  )
}

function Manual() {
  return (
    <>
      <p className="mb-7 mt-0 max-w-measure-detail text-base leading-180">You can log immersion anytime. Join a contest when you’d like a little company and a shared goal.</p>
      <ol className="mb-9 mt-0 grid list-none gap-7 p-0 [&_h2]:mb-3 [&_h2]:mt-0 [&_h2]:font-display [&_h2]:text-xl [&_h2]:leading-150 [&_p]:mb-4 [&_p]:mt-0 [&_p]:max-w-measure-prose [&_p]:leading-180 [&_li]:relative [&_li]:scroll-mt-6 [&_li]:ps-10 max-tablet:[&_li]:ps-8">
        <li id="participate">
          <span aria-hidden="true" className={stepNumber}>1</span>
          <h2>Choose how you want to participate</h2>
          <p>Track on your own, or choose an official or community contest. Check its dates and permitted activities. If registration is open, choose up to three languages and join.</p>
          <Link className="text-link" to="/contests">Find a contest</Link>
        </li>
        <li id="first-log">
          <span aria-hidden="true" className={stepNumber}>2</span>
          <h2>Read, listen, then log activity</h2>
          <p>Choose the language and activity, enter the amount or time requested by the form, and add a description or tags if useful. Check the score preview and contest submissions before saving.</p>
          <Link className={buttonClassName()} to="/logs/new">Log activity</Link>
        </li>
        <li id="score-context">
          <span aria-hidden="true" className={stepNumber}>3</span>
          <h2>Check where your activity counts</h2>
          <p>Your log details show its personal score. When you can view its contest submissions, each contribution is listed separately. Different scoring rules can produce different scores from the same activity.</p>
          <Link className="text-link" to="/logs/reading-konbini">Explore a sample log</Link>
        </li>
      </ol>
      <aside className={cn(accentRail(), 'py-1 pe-0 ps-5 [&_h2]:mb-2 [&_h2]:mt-0 [&_h2]:font-body [&_h2]:text-base [&_h2]:leading-160 [&_p]:mb-3 [&_p]:mt-0 [&_p]:leading-180')}>
        <h2>Already tracking and want to join a round?</h2>
        <p>Joining a contest and submitting a log are separate actions. Check the contest’s start, end and registration dates before assuming an activity can count.</p>
        <Link className="text-link" to="/contests/round5">Check Round 5’s dates</Link>
      </aside>
      <section className={guideSection} aria-labelledby="guide-next">
        <h2 id="guide-next">Find the detail you need</h2>
        <ul className="m-0 grid list-none gap-3 p-0">
          <li><Link className="text-link" to="/guide/scoring">Units and contest rules</Link></li>
          <li><Link className="text-link" to="/guide/questions">Help with registration and languages</Link></li>
          <li><Link className="text-link" to="/users/anton">Explore your profile</Link></li>
        </ul>
      </section>
    </>
  )
}

function Rules() {
  return (
    <>
      <p className="mb-7 mt-0 max-w-measure-detail text-base leading-180">The competition is friendly. The record should be honest. Every contest has its own dates, permitted activities and scoring rules; check that context before submitting.</p>
      <section className={guideSection} id="dates">
        <h2>Contest times are local to you</h2>
        <p>Contest start times and deadlines are shown in your device’s time zone. Activity dates stay on the day you recorded. The registration deadline and contest end date are different. A round may still be running after it stops accepting new participants. Activity must fall within the contest’s eligible period.</p>
        <Link className="text-link" to="/contests/round5">See a contest’s dates and rules</Link>
      </section>
      <section className={guideSection} id="fair-participation">
        <h2>Fair participation</h2>
        <ul className="m-0 grid gap-3 ps-5.5 leading-180">
          <li>Submit reading or listening you actually completed, as accurately as reasonably possible.</li>
          <li>Log regularly. Recording a little each day makes your progress easier to follow.</li>
          <li>Check repeat eligibility before submitting material you have read or listened to before.</li>
          <li>Use an appropriate display name and respect the other people taking part.</li>
        </ul>
      </section>
      <section className={guideSection} id="scores">
        <h2>Understand the score you see</h2>
        <p>Record reading in pages, sentences or characters, and listening in minutes. A log’s personal score and its contest scores are separate records. Never add them together to describe that log.</p>
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
        <p>For pages, choose at most one modifier: Manga or Comic multiplies the score by 0.2; Two column by 1.6. Passive listening multiplies the listening score by 0.5. The reading and listening circle contests use their own fixed minute rates without these personal modifiers.</p>
        <Table
          caption="One record, two scoring contexts"
          captionVisibility="screen-reader"
          minWidth="0"
          columns={[
            { id: 'context', header: 'Context and example rule', rowHeader: true, cell: row => <span className="block leading-150 [&_small]:mt-1 [&_small]:block [&_small]:text-fine [&_small]:font-regular [&_small]:text-muted">{row.context}<small>{row.rule}</small></span> },
            { id: 'score', header: 'Score', align: 'end', cell: row => row.score },
          ]}
          rows={[
            { context: 'Personal record', rule: '20 pages × 1 point', score: '20' },
            { context: 'A community contest', rule: '20 pages × 0.5 points', score: '10' },
          ]}
        />
        <p className="guide-fineprint">These rates are illustrative. The rule saved with your log or contest submission determines its score.</p>
        <Link className="text-link" to="/logs/reading-konbini">See the breakdown in log details</Link>
      </section>
      <section className={guideSection} id="eligibility">
        <h2>Eligibility and repeats</h2>
        <p>Official rounds include reading and listening. A community contest may accept a smaller set of activities or languages. Books, manga, games, podcasts and video describe the material; choose the activity you actually did.</p>
        <p>For repeats or changes to playback speed, consult the applicable contest guidance before submitting. Check the score preview; do not apply a second manual adjustment when a scoring rule already accounts for it.</p>
      </section>
      <section className={guideSection} id="questions">
        <h2>When something looks wrong</h2>
        <p>If a submission seems incorrect, contact the team with the log and contest links. Include the amount recorded and the score you expected so the context is clear.</p>
        <Link className="text-link" to="/contact">Ask about a rule</Link>
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
        <form.Form role="search" className="mb-2">
          <form.AppField name="query">{field => <field.Input label="Search these questions" type="search" autoComplete="off" aria-describedby="guide-search-results" placeholder="Try registration, language or score" />}</form.AppField>
          <div className="flex min-h-target items-center justify-between gap-4 [&_p]:my-2 [&_p]:text-fine [&_p]:text-muted">
            <p id="guide-search-results" role="status">{visible.length} {visible.length === 1 ? 'question' : 'questions'}</p>
            {query && <Button variant="link" onClick={clear}>Clear search</Button>}
          </div>
        </form.Form>
      </form.AppForm>
      <div className="[&_details]:border-t [&_details]:border-rule [&_details:last-child]:border-b [&_summary]:flex [&_summary]:cursor-pointer [&_summary]:list-none [&_summary]:items-baseline [&_summary]:justify-between [&_summary]:gap-5 [&_summary]:py-4.5 [&_summary]:font-semibold [&_summary]:leading-165 [&_summary::-webkit-details-marker]:hidden">
        {visible.map(item => (
          <details key={item.question}>
            <summary>
              {item.question}
              <span
                aria-hidden="true"
                className="me-1 size-1.75 flex-none rotate-eighth border-b-chevron border-e-chevron border-muted [details[open]_&]:rotate-chevron-up"
              />
            </summary>
            <div className="pb-6 [&_p]:mb-4 [&_p]:mt-0 [&_p]:max-w-measure-answer [&_p]:leading-180">
              <p>{item.answer}</p>
              <Link className="text-link" to={item.href}>{item.link}</Link>
            </div>
          </details>
        ))}
      </div>
      {visible.length === 0 && (
        <EmptyState headingLevel={2} title="No questions match that search." description="Try a broader term or clear the search." />
      )}
      <section className={guideSection} id="contact">
        <h2>Still looking for an answer?</h2>
        <p>Tell us what you were trying to do and where you got stuck.</p>
        <Link className={buttonClassName({ variant: 'outline' })} to="/contact">Contact Tadoku</Link>
      </section>
    </>
  )
}

const stepNumber = 'absolute start-0 top-0 font-display text-xl leading-150 font-bold text-muted'

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
    <div className="grid grid-cols-guide items-start gap-14 max-lg:grid-cols-guide-narrow max-lg:gap-9 max-tablet:grid-cols-single max-tablet:gap-7">
      <nav className="sticky top-7 flex flex-col gap-1 max-tablet:static max-tablet:grid max-tablet:grid-cols-halves max-tablet:gap-1 max-tablet:border-b max-tablet:border-rule max-tablet:pb-4 [&_a]:flex [&_a]:min-h-target [&_a]:items-center [&_a]:border-s-edge [&_a]:border-transparent [&_a]:px-3 [&_a]:py-2.5 [&_a]:text-sm [&_a]:leading-150 [&_a]:text-muted [&_a]:no-underline [&_a:hover]:text-link [&_a:hover]:underline [&_a[aria-current=page]]:border-link [&_a[aria-current=page]]:bg-paper [&_a[aria-current=page]]:font-bold [&_a[aria-current=page]]:text-ink max-tablet:[&_a]:px-2.5 max-tablet:[&_a]:py-2 max-xs:[&_a]:text-fine" aria-label={ia === 'separate' ? 'Learn' : ia === 'guide' ? 'Guide' : 'Help'}>
        {pageOrder.map(key => <Link key={key} to={routes[key]} className="text-link" aria-current={key === page ? 'page' : undefined}>{currentLabels[key]}</Link>)}
      </nav>
      <article className="min-w-0">
        <PageHeader className="guide-header [&_h1]:max-w-measure-guide-title [&_h1]:text-balance [&_h1]:font-display [&_h1]:text-guide-title [&_h1]:leading-125 [&_h1]:tracking-tight" title={headings[page].title} lead={headings[page].description}>
          {inPageLinks[page] && <nav className="mt-7 flex flex-wrap gap-x-6 gap-y-1 border-y border-rule py-3 text-fine [&_a]:py-1 max-tablet:mt-5 max-tablet:gap-x-5 max-tablet:gap-y-0" aria-label="On this page">{inPageLinks[page]?.map(link => <a key={link.id} href={`#${link.id}`} className="text-link">{link.label}</a>)}</nav>}
        </PageHeader>
        {page === 'about' && <About />}
        {page === 'manual' && <Manual />}
        {page === 'rules' && <Rules />}
        {page === 'faq' && <Questions />}
      </article>
    </div>
  )
}
