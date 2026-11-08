import { PageHeader, prose } from 'paper-ui'

const contentTable = [
  'content-table mb-4 mt-0 max-w-content-table overflow-x-auto [background:var(--effect-canvas-table-scroll)]',
  'focus-visible:outline-ring focus-visible:outline-link focus-visible:outline-offset-ring',
  '[&_table:has(th:nth-child(4))]:min-w-wide-table',
  '[&_:is(th,td):not(:first-child)]:text-end [&_:is(th,td):last-child]:pe-0 max-sm:[&_:is(th,td)]:pe-3',
].join(' ')

/** A unit table: rows on wide screens; on a phone, one block per language, with later blocks showing only what differs. */
export function ManualTable({ id, headers, rows, onPaper = false }: { id: string; headers: string[]; rows: string[][]; onPaper?: boolean }) {
  const merged = rows.reduce<string[][]>((groups, row) => {
    const same = groups.find(group => group.slice(1).join() === row.slice(1).join())
    if (same) same[0] += `, ${row[0]}`
    else groups.push([...row])
    return groups
  }, [])
  return (
    <>
      <div className={`${contentTable} max-sm:hidden ${onPaper ? 'max-w-full [background:var(--effect-table-scroll)]' : ''}`} role="region" tabIndex={0} aria-labelledby={id}>
        <table>
          <thead>
            <tr>{headers.map(header => <th key={header}>{header}</th>)}</tr>
          </thead>
          <tbody>
            {rows.map(row => <tr key={row[0]}>{row.map((cell, index) => <td key={headers[index]}>{cell}</td>)}</tr>)}
          </tbody>
        </table>
      </div>
      <div className="mb-4 grid gap-5 sm:hidden">
        {merged.map(row => (
          <section key={row[0]} aria-label={`${row[0]} units`}>
            <h5 className="m-0 mb-1 text-ui font-semibold">{row[0]}</h5>
            {row === merged[0] ? null : <p className="mb-1 mt-0 text-metadata text-muted">Same as {merged[0][0]}, except:</p>}
            <dl className="m-0">
              {headers.slice(1).map((header, index) => row === merged[0] || row[index + 1] !== merged[0][index + 1] ? (
                <div key={header} className="flex justify-between gap-4 border-b border-rule py-1.5 text-metadata tabular-nums">
                  <dt>{header}</dt>
                  <dd className="m-0">{row[index + 1]}</dd>
                </div>
              ) : null)}
            </dl>
          </section>
        ))}
      </div>
    </>
  )
}

const sections = [
  ['signing-up-for-a-contest', 'Signing up for a contest'],
  ['submitting-updates', 'Submitting updates'],
  ['frequency-of-submitting-updates', 'Frequency of submitting updates'],
  ['submitting-re-reads', 'Submitting re-reads'],
  ['media', 'Tracking activities and media'],
  ['units-reading', 'Reading units'],
  ['units-writing', 'Writing units'],
  ['units-listening', 'Listening units'],
  ['units-speaking', 'Speaking units'],
  ['units-study', 'Study units'],
  ['speeding-up-listening', 'Speeding up listening material'],
  ['ebooks', 'Ebooks'],
] as const

/** Webv2's code-owned manual, ported as written. CMS pages render through `CmsPage` instead. */
export function ManualPage() {
  return (
    <div className="grid items-start lg:grid-cols-workspace lg:gap-12">
      <nav aria-label="On this page" className="sticky top-6 col-start-2 row-start-1 max-lg:hidden">
        <ol className="m-0 list-none gap-x-8 p-0 text-ui leading-175 ">
          {sections.map(([id, label]) => <li key={id} className="break-inside-avoid"><a className="inline-block py-1 text-link no-underline hover:text-link-hover hover:underline" href={`#${id}`}>{label}</a></li>)}
        </ol>
      </nav>
    <article className="min-w-0 max-w-article lg:col-start-1 lg:row-start-1">
      <PageHeader title="Guide" />
      <details className="sticky top-0 z-lift mb-8 border-y border-rule bg-canvas py-3 lg:hidden">
        <summary className="cursor-pointer text-ui font-semibold">On this page</summary>
        <ol className="mb-0 mt-3 list-none columns-1 gap-x-8 p-0 text-ui leading-175 tablet:columns-2">
          {sections.map(([id, label]) => <li key={id} className="break-inside-avoid"><a className="inline-block py-1 text-link no-underline hover:text-link-hover hover:underline" href={`#${id}`}>{label}</a></li>)}
        </ol>
      </details>
      <div className={prose({ className: 'content-prose' })}>
        <h2 id="signing-up-for-a-contest">Signing up for a contest</h2>
        <p>
          A contest is open for registration before it starts. A contest also has
          a registration deadline, generally one week before the contest ends for
          official contests. During this time you can sign up for the contest with
          a maximum of three languages. Submitting updates won&apos;t be possible
          until the contests starts.
        </p>
        <h2 id="submitting-updates">Submitting updates</h2>
        <p>
          You can submit pages once the contest has started. The contest runs
          based on UTC time. You can submit pages that you&apos;ve read within the
          contest timeframe, after the contest has ended you won&apos;t be able to
          add, delete, or change pages anymore.
        </p>
        <h3 id="frequency-of-submitting-updates">
          Frequency of submitting updates
        </h3>
        <p>
          You should be posting updates in regular intervals, preferably daily.
          Submitting everything you&apos;ve read at the end of the contest
          isn&apos;t in line with the spirit of the contest. Tadoku is supposed to
          motivate people to read more, and in order to keep things fair updates
          should be somewhat regular. Participants who keep posting irregular
          (large) updates in batch will be removed from the contest. Repeat
          offenders may be banned from future rounds.
        </p>
        <h3 id="submitting-re-reads">Submitting re-reads (repeats)</h3>
        <p>
          If you have previously read something, but would still like to read it
          again for the contest you can do so (as long as it&apos;s not within the
          same contest). Cut the total amount of pages you&apos;ve read in half
          and submit them as-is. As much as possible I would like to discourage
          re-reading the same material, but I also understand that sometimes it
          makes sense to do so.
        </p>
        <h3 id="media">Tracking different kinds of activities and media</h3>
        <p>
          There are a couple different activities that can be tracked in Tadoku:
          reading, listening, writing, speaking, and generic study sessions. For
          official contests only reading and listening updates will be accepted.
          The score of an update is calculated by the entered amount and chosen
          unit. An estimated score is also visible at the bottom of the form when
          submitting an update.
        </p>
        <p>
          Some languages with different writing systems also have different
          modifiers for certain units. You can find an overview of all the
          activity and unit configurations per language in the table below.
        </p>

        <h4 id="units-reading">Reading</h4>
        <ManualTable id="units-reading" headers={['Language', 'Page', '2 Column page ¹', 'Comic page', 'Sentence', 'Character ²']} rows={[
          ['Japanese', '1', '1.6', '0.2', '0.05', '0.0025'],
          ['Chinese', '1', 'N/A', '0.2', '0.05', '0.0025'],
          ['Korean', '1', 'N/A', '0.2', '0.05', '0.0025'],
          ['All other languages', '1', 'N/A', '0.2', '0.05', '0.0008333'],
        ]} />
        <ol className="mb-6! mt-0! text-metadata leading-metadata text-muted">
          <li>Only available for Japanese</li>
          <li>
            400 characters per page for Japanese, Chinese, and Korean. <br />
            1200 characters per page for all other languages.
          </li>
        </ol>

        <h4 id="units-writing">Writing</h4>
        <ManualTable id="units-writing" headers={['Language', 'Page', 'Sentence', 'Character']} rows={[
          ['Japanese', '10', '0.5', '0.025'],
          ['Chinese', '10', '0.5', '0.025'],
          ['Korean', '10', '0.5', '0.025'],
          ['All other languages', '10', '0.5', '0.0083333'],
        ]} />

        <h4 id="units-listening">Listening</h4>

        <p>
          While listening make sure that it&apos;s comprehensible. Passive
          listening is allowed in official contests as long as you&apos;re paying
          attention to it and can understand what&apos;s going on. You could track
          listening while doing the dishes, but tracking audio you&apos;re
          &quot;listening&quot; to while sleeping would not count.
        </p>

        <ManualTable id="units-listening" headers={['Language', 'Minute', 'Minute (high density) ¹']} rows={[
          ['All languages', '0.4', '0.6'],
        ]} />
        <ol className="mb-6! mt-0! text-metadata leading-metadata text-muted">
          <li>Media such as audiobooks which have few breaks</li>
        </ol>

        <h4 id="units-speaking">Speaking</h4>
        <ManualTable id="units-speaking" headers={['Language', 'Minute', 'Minute (high density) ¹']} rows={[
          ['All languages', '0.5', '0.7'],
        ]} />
        <ol className="mb-6! mt-0! text-metadata leading-metadata text-muted">
          <li>
            When you&apos;re the only one speaking; such as in speeches or
            presentations
          </li>
        </ol>

        <h4 id="units-study">Study</h4>
        <ManualTable id="units-study" headers={['Language', 'Minute ¹']} rows={[
          ['All languages', '0.5'],
        ]} />
        <ol className="mb-6! mt-0! text-metadata leading-metadata text-muted">
          <li>
            Can be used for any kind of generic studying such as flashcards,
            textbooks, tutoring sessions, etc...
          </li>
        </ol>

        <h3 id="speeding-up-listening">Speeding up listening material</h3>
        <p>
          It&apos;s fairly common for advanced learners to change the playback
          speed of audio while listening. In such cases you are allowed to track
          the original audio track length, with a limit of 2x playback speed.
          Anything over 2x playback speed is not eligible for official contests
          and should not be submitted to contests. Make sure that the audio is
          still comprehensible. If it&apos;s not comprehensible it would not be
          eligible for tracking in official contests.
        </p>
        <p>
          For example: you&apos;re listening at <strong>1.5x speed</strong> to a{' '}
          <strong>60 minute</strong> audiobook chapter. This will take you{' '}
          <strong>40 minutes</strong> to finish but can be tracked as{' '}
          <strong>60 minutes</strong> of input.
        </p>

        <h3 id="ebooks">Ebooks</h3>

        <p>
          Tracking pages read of ebooks can be quite tricky. Most of the times
          they are counted in lines, and every book will be different depending on
          the device/software you&apos;re using. Here is one way to calculate your
          page count using the book <strong>こころ</strong> (Kokoro) by Natsume
          Soseki:
        </p>
        <ol>
          <li>
            Get the total page count from the paper copy, a total of{' '}
            <strong>384</strong> pages according to Amazon Japan
          </li>
          <li>
            The kindle shows me there are a total of <strong>4513</strong> lines
            in this book. This means that 1 page from the paper copy is equal to{' '}
            <strong>11.75</strong> lines in the ebook.
          </li>
          <li>
            Let&apos;s say I&apos;ve read from line <strong>400</strong> to{' '}
            <strong>640</strong>, <strong>+240</strong> lines since tracking last
            time. <strong>240 / 11.75 = 20.425531915</strong>
          </li>
          <li>
            I usually round down/up the number when tracking, and correct the
            discrepancy when I finish the book. In this case I&apos;d track this
            as <strong>20.4</strong> book pages read.
          </li>
        </ol>
      </div>
    </article>
    </div>
  )
}
