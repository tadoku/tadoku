import * as stylex from '@stylexjs/stylex'
import { PageHeader, Prose, proseElements as P } from 'paper-ui'
import { breakpoints, border, color, ratio, size, space, type } from 'paper-ui/tokens.stylex'
import { contentStyles } from './page-styles'

const shade = color.proseScrollShade

/** The manual's rate tables: a scroll region whose edge shadows appear only while more columns sit off-screen. */
const styles = stylex.create({
  region: {
    background: `linear-gradient(to right, ${color.surfaceCanvas} ${ratio.scrollCover}, transparent) left / ${space[32]} 100% no-repeat local, linear-gradient(to left, ${color.surfaceCanvas} ${ratio.scrollCover}, transparent) right / ${space[32]} 100% no-repeat local, radial-gradient(farthest-side at 0 50%, ${shade}, transparent) left / ${space[12]} 100% no-repeat scroll, radial-gradient(farthest-side at 100% 50%, ${shade}, transparent) right / ${space[12]} 100% no-repeat scroll`,
    marginInline: 0,
    marginBlockEnd: space[16],
    marginBlockStart: 0,
    maxInlineSize: `min(100%, ${size[704]})`,
    outlineColor: { default: null, ':focus-visible': color.textLink },
    outlineOffset: { default: null, ':focus-visible': border.focusOffset },
    outlineStyle: { default: null, ':focus-visible': 'solid' },
    outlineWidth: { default: null, ':focus-visible': border.focusRing },
    overflowX: 'auto',
  },
  /** Only wide rate tables keep a minimum width and scroll; two- and three-column tables fit the column. */
  wide: { minInlineSize: size[448] },
  cell: {
    paddingInlineEnd: { default: space[16], ':last-child': 0, [breakpoints.phone]: { default: space[12], ':last-child': 0 } },
    textAlign: { default: 'end', ':first-child': 'left' },
  },
  dataCell: { whiteSpace: { default: null, ':first-child': 'nowrap' } },
  unavailable: { color: color.textMuted },
  footnotes: {
    marginInline: 0,
    color: color.textMuted,
    fontSize: type.metadataSize,
    lineHeight: type.metadataLineHeight,
    marginBlockEnd: space[24],
    marginBlockStart: 0,
  },
})

/** Webv2's code-owned manual, ported as written. CMS pages render through `CmsPage` instead. */
export function ManualPage() {
  return (
    <article {...stylex.props(contentStyles.article)}>
      <PageHeader title="Manual" />
      <Prose>
        <P.h2 id="signing-up-for-a-contest">Signing up for a contest</P.h2>
        <P.p>
          A contest is open for registration before it starts. A contest also has
          a registration deadline, generally one week before the contest ends for
          official contests. During this time you can sign up for the contest with
          a maximum of three languages. Submitting updates won&apos;t be possible
          until the contests starts.
        </P.p>
        <P.h2 id="submitting-updates">Submitting updates</P.h2>
        <P.p>
          You can submit pages once the contest has started. The contest runs
          based on UTC time. You can submit pages that you&apos;ve read within the
          contest timeframe, after the contest has ended you won&apos;t be able to
          add, delete, or change pages anymore.
        </P.p>
        <P.h3 id="frequency-of-submitting-updates">
          Frequency of submitting updates
        </P.h3>
        <P.p>
          You should be posting updates in regular intervals, preferably daily.
          Submitting everything you&apos;ve read at the end of the contest
          isn&apos;t in line with the spirit of the contest. Tadoku is supposed to
          motivate people to read more, and in order to keep things fair updates
          should be somewhat regular. Participants who keep posting irregular
          (large) updates in batch will be removed from the contest. Repeat
          offenders may be banned from future rounds.
        </P.p>
        <P.h3 id="submitting-re-reads">Submitting re-reads (repeats)</P.h3>
        <P.p>
          If you have previously read something, but would still like to read it
          again for the contest you can do so (as long as it&apos;s not within the
          same contest). Cut the total amount of pages you&apos;ve read in half
          and submit them as-is. As much as possible I would like to discourage
          re-reading the same material, but I also understand that sometimes it
          makes sense to do so.
        </P.p>
        <P.h3 id="media">Tracking different kinds of activities and media</P.h3>
        <P.p>
          There are a couple different activities that can be tracked in Tadoku:
          reading, listening, writing, speaking, and generic study sessions. For
          official contests only reading and listening updates will be accepted.
          The score of an update is calculated by the entered amount and chosen
          unit. An estimated score is also visible at the bottom of the form when
          submitting an update.
        </P.p>
        <P.p>
          Some languages with different writing systems also have different
          modifiers for certain units. You can find an overview of all the
          activity and unit configurations per language in the table below.
        </P.p>

        <P.h4 id="units-reading">Reading</P.h4>
        <div
          {...stylex.props(styles.region)}
          role="region"
          tabIndex={0}
          aria-labelledby="units-reading"
        >
          <P.table xstyle={styles.wide}>
            <thead>
              <tr>
                <P.th xstyle={styles.cell}>Language</P.th>
                <P.th xstyle={styles.cell}>Page</P.th>
                <P.th xstyle={styles.cell}>
                  2 Column page{' '}
                  <P.sup
                   
                    title="Only available for Japanese"
                  >
                    1
                  </P.sup>
                </P.th>
                <P.th xstyle={styles.cell}>Comic page</P.th>
                <P.th xstyle={styles.cell}>Sentence</P.th>
                <P.th xstyle={styles.cell}>
                  Character{' '}
                  <P.sup
                   
                    title="400 characters per page for Japanese, Chinese, and Korean. 1200 characters per page for all other languages."
                  >
                    2
                  </P.sup>
                </P.th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Japanese</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>1</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>1.6</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.2</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.05</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.0025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Chinese</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>1</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell, styles.unavailable]}>N/A</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.2</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.05</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.0025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Korean</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>1</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell, styles.unavailable]}>N/A</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.2</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.05</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.0025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>All other languages</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>1</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell, styles.unavailable]}>N/A</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.2</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.05</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.0008333</P.td>
              </tr>
            </tbody>
          </P.table>
        </div>
        <P.ol xstyle={styles.footnotes}>
          <P.li>Only available for Japanese</P.li>
          <P.li>
            400 characters per page for Japanese, Chinese, and Korean. <br />
            1200 characters per page for all other languages.
          </P.li>
        </P.ol>

        <P.h4 id="units-writing">Writing</P.h4>
        <div
          {...stylex.props(styles.region)}
          role="region"
          tabIndex={0}
          aria-labelledby="units-writing"
        >
          <P.table xstyle={styles.wide}>
            <thead>
              <tr>
                <P.th xstyle={styles.cell}>Language</P.th>
                <P.th xstyle={styles.cell}>Page</P.th>
                <P.th xstyle={styles.cell}>Sentence</P.th>
                <P.th xstyle={styles.cell}>Character</P.th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Japanese</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>10</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Chinese</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>10</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>Korean</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>10</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.025</P.td>
              </tr>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>All other languages</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>10</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.0083333</P.td>
              </tr>
            </tbody>
          </P.table>
        </div>

        <P.h4 id="units-listening">Listening</P.h4>

        <P.p>
          While listening make sure that it&apos;s comprehensible. Passive
          listening is allowed in official contests as long as you&apos;re paying
          attention to it and can understand what&apos;s going on. You could track
          listening while doing the dishes, but tracking audio you&apos;re
          &quot;listening&quot; to while sleeping would not count.
        </P.p>

        <div
          {...stylex.props(styles.region)}
          role="region"
          tabIndex={0}
          aria-labelledby="units-listening"
        >
          <P.table>
            <thead>
              <tr>
                <P.th xstyle={styles.cell}>Language</P.th>
                <P.th xstyle={styles.cell}>Minute</P.th>
                <P.th xstyle={styles.cell}>
                  Minute (high density){' '}
                  <P.sup
                   
                    title="Media such as audiobooks which have few breaks"
                  >
                    1
                  </P.sup>
                </P.th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>All languages</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.4</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.6</P.td>
              </tr>
            </tbody>
          </P.table>
        </div>
        <P.ol xstyle={styles.footnotes}>
          <P.li>Media such as audiobooks which have few breaks</P.li>
        </P.ol>

        <P.h4 id="units-speaking">Speaking</P.h4>
        <div
          {...stylex.props(styles.region)}
          role="region"
          tabIndex={0}
          aria-labelledby="units-speaking"
        >
          <P.table>
            <thead>
              <tr>
                <P.th xstyle={styles.cell}>Language</P.th>
                <P.th xstyle={styles.cell}>Minute</P.th>
                <P.th xstyle={styles.cell}>
                  Minute (high density){' '}
                  <P.sup
                   
                    title="When you're the only one speaking; such as in speeches or presentations"
                  >
                    1
                  </P.sup>
                </P.th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>All languages</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.7</P.td>
              </tr>
            </tbody>
          </P.table>
        </div>
        <P.ol xstyle={styles.footnotes}>
          <P.li>
            When you&apos;re the only one speaking; such as in speeches or
            presentations
          </P.li>
        </P.ol>

        <P.h4 id="units-study">Study</P.h4>
        <div
          {...stylex.props(styles.region)}
          role="region"
          tabIndex={0}
          aria-labelledby="units-study"
        >
          <P.table>
            <thead>
              <tr>
                <P.th xstyle={styles.cell}>Language</P.th>
                <P.th xstyle={styles.cell}>
                  Minute{' '}
                  <P.sup
                   
                    title="Can be used for any kind of generic studying such as flashcards, textbooks, tutoring sessions, etc..."
                  >
                    1
                  </P.sup>
                </P.th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <P.td xstyle={[styles.cell, styles.dataCell]}>All languages</P.td>
                <P.td xstyle={[styles.cell, styles.dataCell]}>0.5</P.td>
              </tr>
            </tbody>
          </P.table>
        </div>
        <P.ol xstyle={styles.footnotes}>
          <P.li>
            Can be used for any kind of generic studying such as flashcards,
            textbooks, tutoring sessions, etc...
          </P.li>
        </P.ol>

        <P.h3>Speeding up listening material</P.h3>
        <P.p>
          It&apos;s fairly common for advanced learners to change the playback
          speed of audio while listening. In such cases you are allowed to track
          the original audio track length, with a limit of 2x playback speed.
          Anything over 2x playback speed is not eligible for official contests
          and should not be submitted to contests. Make sure that the audio is
          still comprehensible. If it&apos;s not comprehensible it would not be
          eligible for tracking in official contests.
        </P.p>
        <P.p>
          For example: you&apos;re listening at <strong>1.5x speed</strong> to a{' '}
          <strong>60 minute</strong> audiobook chapter. This will take you{' '}
          <strong>40 minutes</strong> to finish but can be tracked as{' '}
          <strong>60 minutes</strong> of input.
        </P.p>

        <P.h3>Ebooks</P.h3>

        <P.p>
          Tracking pages read of ebooks can be quite tricky. Most of the times
          they are counted in lines, and every book will be different depending on
          the device/software you&apos;re using. Here is one way to calculate your
          page count using the book <strong>こころ</strong> (Kokoro) by Natsume
          Soseki:
        </P.p>
        <P.ol>
          <P.li>
            Get the total page count from the paper copy, a total of{' '}
            <strong>384</strong> pages according to Amazon Japan
          </P.li>
          <P.li>
            The kindle shows me there are a total of <strong>4513</strong> lines
            in this book. This means that 1 page from the paper copy is equal to{' '}
            <strong>11.75</strong> lines in the ebook.
          </P.li>
          <P.li>
            Let&apos;s say I&apos;ve read from line <strong>400</strong> to{' '}
            <strong>640</strong>, <strong>+240</strong> lines since tracking last
            time. <strong>240 / 11.75 = 20.425531915</strong>
          </P.li>
          <P.li>
            I usually round down/up the number when tracking, and correct the
            discrepancy when I finish the book. In this case I&apos;d track this
            as <strong>20.4</strong> book pages read.
          </P.li>
        </P.ol>
      </Prose>
    </article>
  )
}
