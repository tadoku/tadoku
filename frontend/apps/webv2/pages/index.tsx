import type { NextPage } from 'next'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import { HomeIcon } from '@heroicons/react/20/solid'

interface Props {}

const Index: NextPage<Props> = () => {
  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[{ label: 'Home', href: '/', IconComponent: HomeIcon }]}
        />
      </div>
      <div className="h-stack">
        <div className="card flex flex-col justify-center bg-sky-50">
          <h1 className="title text-xl">Get good at your second language</h1>
          <p>
            Tadoku is a friendly foreign-language reading contest aimed at
            building a habit of reading in your non-native languages.
          </p>
        </div>
        <div className="card">
          <h1 className="title text-xl">Why should I participate?</h1>
          <p>
            Extensive reading of native materials is a great way to improve your
            understanding of the language you&apos;re learning. There are many
            benefits to doing so: it builds vocabulary, reinforces grammar
            patterns, and you learn about the culture where your language is
            spoken. As you participate in more rounds you will notice that you
            can read more and more as you improve.
          </p>
          <p>
            That said, it&apos;s not for everyone. Not everyone enjoys the
            process of immersing themselves. Tadoku isn&apos;t a magical pill
            that will make you fluent. It only covers extensive reading, and not
            extensive listening. While Tadoku is here to promote reading, a
            balanced approach to learning is still recommended.
          </p>
        </div>
      </div>
    </>
  )
}

export default Index
