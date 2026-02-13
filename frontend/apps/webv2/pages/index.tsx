import { routes } from '@app/common/routes'
import { useSession } from '@app/common/session'
import { usePostList } from '@app/content/api'
import { PostBody } from '@app/content/Post'
import { useYearlyLeaderboard } from '@app/immersion/api'
import { Leaderboard } from '@app/immersion/Leaderboard'
import { ArrowLongRightIcon } from '@heroicons/react/20/solid'
import { BookOpenIcon } from '@heroicons/react/24/solid'
import { DateTime, Interval } from 'luxon'
import type { NextPage } from 'next'
import Link from 'next/link'

interface Props { }

const Index: NextPage<Props> = () => (
  <>
    <div className="w-full min-h-screen absolute top-0 left-0 right-0 bg-[url('/img/header.jpg')] bg-no-repeat bg-top z-0 pointer-events-none"></div>
    <div className="relative flex flex-col lg:h-stack space-y-4 lg:space-y-0 lg:space-x-8 z-10">
      <div className="lg:w-2/6 v-stack space-y-4 lg:space-y-8">
        <div className="card flex-grow flex flex-col justify-center bg-sky-50">
          <h1 className="title text-xl">Get good at your second language</h1>
          <p>
            Tadoku is a friendly foreign-language immersion contest and tracking
            platform aimed at building a habit of reading and listening in your
            non-native languages.
          </p>
        </div>
        <AboutSection />
      </div>
      <div className="flex-grow">
        <Top10Leaderboard />
      </div>
    </div>
    <div className="mt-4 lg:mt-8 flex flex-col lg:flex-row w-full space-y-4 lg:space-y-0 lg:space-x-8">
      <div className="v-stack flex-grow space-y-4 lg:space-y-8">
        <ContestSchedule />
        <FAQ />
      </div>
      <LatestBlogPost />
    </div>
  </>
)

function AboutSection() {
  const [session] = useSession()

  return (
    <div className="card">
      <h1 className="title text-xl">Why should I participate?</h1>
      <p>
        Extensive reading and listening of native materials is a great way to
        improve your understanding of the language you&apos;re learning. There
        are many benefits to doing so: it builds vocabulary, reinforces grammar
        patterns, and you learn about the culture where your language is spoken.
      </p>
      <p>
        As you track your progress over time you will notice that you can
        understand more and more of the language you&apos;re learning.
      </p>
      {session ? null : (
        <Link
          href={routes.authSignup()}
          className="mt-4 btn primary block !w-full"
        >
          Join Tadoku
          <BookOpenIcon className="ml-2" />
        </Link>
      )}
    </div>
  )
}

function LatestBlogPost() {
  const posts = usePostList({ pageSize: 1, page: 0 })

  if (!posts.data?.posts[0]) {
    return null
  }

  return (
    <div className="lg:w-3/6 flex-shrink-0">
      <div className="card">
        <div className="h-stack w-full items-center justify-between space-x-4 mb-4">
          <h2 className="title text-xl">
            Latest blog post:{' '}
            <Link href={routes.blogPost(posts.data.posts[0].slug)}>
              {posts.data.posts[0].title}
            </Link>
          </h2>
          <Link href={routes.blogList()} className="btn flex-shrink-0">
            <span className="hidden md:inline">More posts</span>
            <span className="inline md:hidden">More</span>
            <ArrowLongRightIcon className="ml-2" />
          </Link>
        </div>
        <div className="auto-format">
          <PostBody post={posts.data.posts[0]} />
        </div>
      </div>
    </div>
  )
}

function Top10Leaderboard() {
  const now = DateTime.utc()
  const year = now.year

  const leaderboard = useYearlyLeaderboard({
    year,
    pageSize: 10,
    page: 1,
  })

  return (
    <div className="card p-0">
      <div className="h-stack w-full items-center justify-between p-4 pb-2 lg:p-7 lg:pb-4">
        <h2 className="title text-xl">{year} Leaderboard Top 10</h2>
        <Link href={routes.leaderboardYearlyOfficial()} className="btn">
          <span className="hidden md:inline">See more</span>
          <span className="inline md:hidden">More</span>
          <ArrowLongRightIcon className="ml-2" />
        </Link>
      </div>
      <Leaderboard
        leaderboard={leaderboard}
        urlForRow={userId => routes.userProfileStatistics(userId)}
        embedded={true}
        emptyMessage={`No participants yet in ${year}, be the first to submit!`}
      />
    </div>
  )
}

const questions = [
  {
    question: "I can't find my language in the list, what do I do?",
    answer:
      "While most popular languages have been added, we haven't added every language. If your desired language is missing you can request your language to be added through our Discord server or leave us a message on the contact page.",
  },
  {
    question: 'Can I join a contest which has already started?',
    answer:
      'Contests have a deadline for registration. You can join a contest as long as the registration period has not ended. For official contests registrations will generally close one week before it ends.',
  },
  {
    question: 'How do I register for a contest?',
    answer:
      'You can find the button to register for a contest in the top right corner of the contest leaderboard. This will only be visible during the contest registration period.',
  },
]

function FAQ() {
  return (
    <div className="card">
      <div className="h-stack w-full items-center justify-between mb-4">
        <h2 className="title text-xl">
          <Link href={routes.blogPage('faq')}>Frequently Asked Questions</Link>
        </h2>
        <Link href={routes.blogPage('faq')} className="btn">
          <span className="hidden md:inline">More answers</span>
          <span className="inline md:hidden">More</span>
          <ArrowLongRightIcon className="ml-2" />
        </Link>
      </div>
      <div>
        <ul className="space-y-4">
          {questions.map((it, i) => (
            <li key={i}>
              <div className="font-bold text-lg">{it.question}</div>
              <div className="mt-2">{it.answer}</div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

const scheduledContests = [
  { name: 'Round 1', startMonth: 1, endDay: 31 },
  { name: 'Round 2', startMonth: 3, endDay: 31 },
  { name: 'Round 3', startMonth: 5, endDay: 14 },
  { name: 'Round 4', startMonth: 7, endDay: 31 },
  { name: 'Round 5', startMonth: 9, endDay: 30 },
  { name: 'Round 6', startMonth: 11, endDay: 14 },
]

function ContestSchedule() {
  const now = DateTime.utc()
  const year = now.year

  return (
    <div className="card p-0">
      <h2 className="title text-xl p-4 pb-2 lg:p-7 lg:pb-4">
        Contest Schedule for {year}
      </h2>
      <table className="default w-full shadow-none">
        <thead>
          <tr>
            <th className="default !pl-4">Round</th>
            <th className="default">Start</th>
            <th className="default">End</th>
            <th className="default">Status</th>
          </tr>
        </thead>
        <tbody>
          {scheduledContests.map(c => {
            const start = DateTime.utc(year, c.startMonth, 1)
            const end = DateTime.utc(year, c.startMonth, c.endDay).endOf('day')
            const interval = Interval.fromDateTimes(start, end)
            return (
              <tr key={c.name}>
                <td className="default !pl-7">
                  <strong>{c.name}</strong>
                </td>
                <td className="default">{start.toFormat('MMM d')}</td>
                <td className="default">{end.toFormat('MMM d')}</td>
                <td className="default">
                  {interval.isAfter(now) ? (
                    <span>Scheduled</span>
                  ) : interval.contains(now) ? (
                    <strong className="text-lime-700">Ongoing</strong>
                  ) : (
                    <span className="text-red-700">Ended</span>
                  )}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

export default Index
