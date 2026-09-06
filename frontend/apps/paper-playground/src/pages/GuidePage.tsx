import { Button, Input, Table, buttonClassName } from 'paper-ui'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link } from 'react-router-dom'
import { usePlayground } from '../state'
import './editorial.css'

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
    title: 'A little more immersion. A little company along the way.',
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
      <section className="guide-section" id="community">
        <h2>A book you want to finish. A voice you want to understand.</h2>
        <p>Tadoku gives you a reason to come back tomorrow. Keep a personal record of your reading and listening, or join a friendly contest and follow the progress of other learners.</p>
        <p>Immersion can be a few pages on the train, an episode over lunch, or the book you finally feel ready to read. There is room for all of it here.</p>
        <Link className={buttonClassName()} to="/contests">Explore contests</Link>
      </section>

      <section className="guide-section" id="history">
        <h2>From a reading contest to an immersion habit</h2>
        <dl className="guide-timeline">
          <div><dt>2010</dt><dd>LordSilent founded Tadoku as a reading contest on Twitter, with an online leaderboard.</dd></div>
          <div><dt>2018</dt><dd>Changes to the Twitter API disrupted the original experience. The project later changed hands, and work began on a new website.</dd></div>
          <div><dt>2023</dt><dd>Tadoku expanded into year-round tracking, activities beyond reading, and community-run contests.</dd></div>
        </dl>
      </section>

      <section className="guide-section" id="contribute">
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
      <p className="guide-introduction">Your first log is enough to get started. You can record immersion throughout the year; a contest adds a shared period and a leaderboard.</p>
      <ol className="guide-steps">
        <li id="participate">
          <h2>Choose how you want to participate</h2>
          <p>Track on your own, or choose an official or community contest. Check its dates and permitted activities. If registration is open, choose up to three languages and join.</p>
          <Link className="text-link" to="/contests">Find a contest</Link>
        </li>
        <li id="first-log">
          <h2>Read, listen, then log activity</h2>
          <p>Choose the language and activity, enter the amount or time requested by the form, and add a description or tags if useful. Check the score preview and contest submissions before saving.</p>
          <Link className={buttonClassName()} to="/logs/new">Log activity</Link>
        </li>
        <li id="score-context">
          <h2>Check where your activity counts</h2>
          <p>Your log details show its personal score. When you can view its contest submissions, each contribution is listed separately. Different scoring rules can produce different scores from the same activity.</p>
          <Link className="text-link" to="/logs/reading-konbini">Explore a sample log</Link>
        </li>
      </ol>
      <aside className="guide-note paper-accent-rail">
        <h2>Already tracking and want to join a round?</h2>
        <p>Joining a contest and submitting a log are separate actions. Check the contest’s start, end and registration dates before assuming an activity can count.</p>
        <Link className="text-link" to="/contests/round5">Check Round 5’s dates</Link>
      </aside>
      <section className="guide-section" aria-labelledby="guide-next">
        <h2 id="guide-next">Find the detail you need</h2>
        <ul className="guide-related">
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
      <p className="guide-introduction">The competition is friendly. The record should be honest. Every contest has its own dates, permitted activities and scoring rules; check that context before submitting.</p>
      <section className="guide-section" id="dates">
        <h2>Dates use UTC</h2>
        <p>The registration deadline and contest end date are different. A round may still be running after it stops accepting new participants. Activity must fall within the contest’s eligible period.</p>
        <Link className="text-link" to="/contests/round5">See a contest’s dates and rules</Link>
      </section>
      <section className="guide-section" id="fair-participation">
        <h2>Fair participation</h2>
        <ul className="guide-prose-list">
          <li>Submit reading or listening you actually completed, as accurately as reasonably possible.</li>
          <li>Log regularly. Recording a little each day makes your progress easier to follow.</li>
          <li>Check repeat eligibility before submitting material you have read or listened to before.</li>
          <li>Use an appropriate display name and respect the other people taking part.</li>
        </ul>
      </section>
      <section className="guide-section" id="scores">
        <h2>Understand the score you see</h2>
        <p>Reading and listening can be measured in pages or time. A log’s personal score and its contest scores are separate records. Never add them together to describe that log.</p>
        <Table
          caption="One record, two scoring contexts"
          minWidth="0"
          columns={[
            { id: 'context', header: 'Context', rowHeader: true, cell: row => row.context },
            { id: 'rule', header: 'Example rule', cell: row => row.rule },
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
      <section className="guide-section" id="eligibility">
        <h2>Eligibility and repeats</h2>
        <p>Official rounds include reading and listening. A community contest may accept a smaller set of activities or languages. Books, manga, games, podcasts and video describe the material; choose the activity you actually did.</p>
        <p>For repeats or changes to playback speed, consult the applicable contest guidance before submitting. Check the score preview; do not apply a second manual adjustment when a scoring rule already accounts for it.</p>
      </section>
      <section className="guide-section" id="questions">
        <h2>When something looks wrong</h2>
        <p>If a submission seems incorrect, contact the team with the log and contest links. Include the amount recorded and the score you expected so the context is clear.</p>
        <Link className="text-link" to="/contact">Ask about a rule</Link>
      </section>
    </>
  )
}

function Questions() {
  const form = useForm({ defaultValues: { query: '' } })
  const query = useWatch({ control: form.control, name: 'query' }).trim().toLocaleLowerCase()
  const visible = questions.filter(item => `${item.question} ${item.answer}`.toLocaleLowerCase().includes(query))
  const clear = () => {
    form.setValue('query', '')
    form.setFocus('query')
  }

  return (
    <>
      <FormProvider {...form}>
        <form role="search" onSubmit={form.handleSubmit(() => undefined)} className="guide-search">
          <Input name="query" label="Search these questions" type="search" autoComplete="off" aria-describedby="guide-search-results" placeholder="Try registration, language or score" />
          <div className="guide-search-meta">
            <p id="guide-search-results" role="status">{visible.length} {visible.length === 1 ? 'question' : 'questions'}</p>
            {query && <Button variant="link" onClick={clear}>Clear search</Button>}
          </div>
        </form>
      </FormProvider>
      <div className="guide-questions">
        {visible.map(item => (
          <details key={item.question}>
            <summary>{item.question}</summary>
            <div className="guide-answer">
              <p>{item.answer}</p>
              <Link className="text-link" to={item.href}>{item.link}</Link>
            </div>
          </details>
        ))}
      </div>
      {visible.length === 0 && (
        <div className="empty-state">
          <h2>No questions match that search.</h2>
          <p>Try a broader term or clear the search.</p>
        </div>
      )}
      <section className="guide-section" id="contact">
        <h2>Still looking for an answer?</h2>
        <p>Tell us what you were trying to do and where you got stuck.</p>
        <Link className={buttonClassName({ variant: 'outline' })} to="/contact">Contact Tadoku</Link>
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
    <div className="guide-layout">
      <nav className="guide-navigation" aria-label={ia === 'separate' ? 'Learn' : ia === 'guide' ? 'Guide' : 'Help'}>
        {pageOrder.map(key => <Link key={key} to={routes[key]} className="text-link" aria-current={key === page ? 'page' : undefined}>{currentLabels[key]}</Link>)}
      </nav>
      <article className="guide-article">
        <header className="guide-header">
          <p className="guide-page-name">{currentLabels[page]}</p>
          <h1>{headings[page].title}</h1>
          <p className="page-lead">{headings[page].description}</p>
          {inPageLinks[page] && <nav className="guide-contents" aria-label="On this page">{inPageLinks[page]?.map(link => <a key={link.id} href={`#${link.id}`} className="text-link">{link.label}</a>)}</nav>}
        </header>
        {page === 'about' && <About />}
        {page === 'manual' && <Manual />}
        {page === 'rules' && <Rules />}
        {page === 'faq' && <Questions />}
      </article>
    </div>
  )
}
