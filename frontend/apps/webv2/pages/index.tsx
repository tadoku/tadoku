import type { NextPage } from 'next'

interface Props {}

const Index: NextPage<Props> = () => {
  return (
    <>
      <div className="w-full min-h-screen absolute -top-16 left-0 right-0 bg-[url('/img/header.jpg')] bg-no-repeat bg-top z-0"></div>
      <div className="relative h-stack space-x-8 z-10">
        <div className="w-2/6 v-stack space-y-8">
          <div className="card flex flex-col justify-center bg-sky-50">
            <h1 className="title text-xl">Get good at your second language</h1>
            <p>
              Tadoku is a friendly foreign-language immersion contest and
              tracking platform aimed at building a habit of reading and
              listening in your non-native languages.
            </p>
          </div>
          <div className="card">
            <h1 className="title text-xl">Why should I participate?</h1>
            <p>
              Extensive reading and listening of native materials is a great way
              to improve your understanding of the language you&apos;re
              learning. There are many benefits to doing so: it builds
              vocabulary, reinforces grammar patterns, and you learn about the
              culture where your language is spoken. As you track your progress
              over time you will notice that you can understand more and more of
              the language you're learning.
            </p>
          </div>
        </div>
        <div className="flex-grow">
          <div className="card">
            <h2>All-Time Leaderboard</h2>
          </div>
        </div>
      </div>
    </>
  )
}

export default Index
