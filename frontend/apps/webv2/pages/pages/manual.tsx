import { routes } from '@app/common/routes'
import { HomeIcon } from '@heroicons/react/20/solid'
import type { NextPage } from 'next'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
interface Props {}

const Page: NextPage<Props> = () => (
  <>
    <Head>
      <title>Manual - Tadoku</title>
    </Head>
    <div className="pb-4">
      <Breadcrumb
        links={[
          { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
          { label: 'Manual', href: routes.blogPage('manual') },
        ]}
      />
    </div>

    <h1 className="title my-4">Manual</h1>
    <div className="max-w-3xl auto-format">
      <p>This document is a work in progress and will be updated over time.</p>
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

      <h4>Reading</h4>
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Language</th>
              <th>Page</th>
              <th>
                2 Column page{' '}
                <sup
                  className="text-primary"
                  title="Only available for Japanese"
                >
                  1
                </sup>
              </th>
              <th>Comic page</th>
              <th>Sentence</th>
              <th>
                Character{' '}
                <sup
                  className="text-primary"
                  title="400 characters per page for Japanese, Chinese, and Korean. 1200 characters per page for all other languages."
                >
                  2
                </sup>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-bold">Japanese</td>
              <td>1</td>
              <td>1.6</td>
              <td>0.2</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">Chinese</td>
              <td>1</td>
              <td className="disabled">N/A</td>
              <td>0.2</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">Korean</td>
              <td>1</td>
              <td className="disabled">N/A</td>
              <td>0.2</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">All other languages</td>
              <td>1</td>
              <td className="disabled">N/A</td>
              <td>0.2</td>
              <td>0.05</td>
              <td>0.0008333</td>
            </tr>
          </tbody>
        </table>
      </div>
      <ol className="list-decimal ml-5 my-4 text-sm text-gray-800 italic">
        <li>Only available for Japanese</li>
        <li>
          400 characters per page for Japanese, Chinese, and Korean. <br />
          1800 characters per page for all other languages.
        </li>
      </ol>

      <h4>Writing</h4>
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Language</th>
              <th>Page</th>
              <th>Sentence</th>
              <th>Character</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-bold">Japanese</td>
              <td>1</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">Chinese</td>
              <td>1</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">Korean</td>
              <td>1</td>
              <td>0.05</td>
              <td>0.0025</td>
            </tr>
            <tr>
              <td className="font-bold">All other languages</td>
              <td>1</td>
              <td>0.05</td>
              <td>0.0008333</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h4>Listening</h4>

      <p>
        While listening make sure that it&apos;s comprehensible. Passive
        listening is allowed in official contests as long as you&apos;re paying
        attention to it and can understand what&apos;s going on. You could track
        listening while doing the dishes, but tracking audio you&apos;re
        &quot;listening&quot; to while sleeping would not count.
      </p>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Language</th>
              <th>Minute</th>
              <th>
                Minute (high density){' '}
                <sup
                  className="text-primary"
                  title="Media such as audiobooks which have few breaks"
                >
                  1
                </sup>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-bold">All languages</td>
              <td>0.4</td>
              <td>0.6</td>
            </tr>
          </tbody>
        </table>
      </div>
      <ol className="list-decimal ml-5 my-4 text-sm text-gray-800 italic">
        <li>Media such as audiobooks which have few breaks</li>
      </ol>

      <h4>Speaking</h4>
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Language</th>
              <th>Minute</th>
              <th>
                Minute (high density){' '}
                <sup
                  className="text-primary"
                  title="When you're the only one speaking; such as in speeches or presentations"
                >
                  1
                </sup>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-bold">All languages</td>
              <td>0.5</td>
              <td>0.7</td>
            </tr>
          </tbody>
        </table>
      </div>
      <ol className="list-decimal ml-5 my-4 text-sm text-gray-800 italic">
        <li>
          When you&apos;re the only one speaking; such as in speeches or
          presentations
        </li>
      </ol>

      <h4>Study</h4>
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Language</th>
              <th>
                Minute{' '}
                <sup
                  className="text-primary"
                  title="Can be used for any kind of generic studying such as flashcards, textbooks, tutoring sessions, etc..."
                >
                  1
                </sup>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-bold">All languages</td>
              <td>0.5</td>
            </tr>
          </tbody>
        </table>
      </div>
      <ol className="list-decimal ml-5 my-4 text-sm text-gray-800 italic">
        <li>
          Can be used for any kind of generic studying such as flashcards,
          textbooks, tutoring sessions, etc...
        </li>
      </ol>

      <h3>Speeding up listening material</h3>
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

      <h3>Ebooks</h3>

      <p>
        Tracking pages read of ebooks can be quite tricky. Most of the times
        they are counted in lines, and every book will be different depending on
        the device/software you&apos;re using. Here is one way to calculate your
        page count using the book <strong>こころ</strong> (Kokoro) by Natsume
        Soseki:
      </p>
      <ol className="list-decimal ml-5 space-y-2">
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
  </>
)

export default Page
