import { routes } from '@app/common/routes'
import Link from 'next/link'
import { LogoInverted } from 'ui'
import { DiscordIcon, GitHubIcon, TwitterIcon } from './Icons'

export default function Footer() {
  return (
    <div className="w-full md:bg-[url('/img/footer.png')] bg-no-repeat bg-top bg-[#1A1A1A]">
      <div className="max-w-7xl mx-auto p-8 flex flex-col-reverse md:flex-row space-y-8 space-y-reverse md:space-y-0 justify-between h-full">
        <div className="flex flex-col-reverse md:flex-col md:space-y-8 text-center items-center md:items-start">
          <LogoInverted className="hidden md:block" />
          <div className="mt-8 md:mt-0 text-white md:flex-grow">
            Built by{' '}
            <a
              href={routes.personalWebsite()}
              className="reset text-white hover:text-primary border-b-2 border-primary transition-all"
            >
              antonve
            </a>
          </div>

          <div className="space-x-4 flex">
            <a
              href={routes.twitter()}
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <TwitterIcon className="w-8 h-8" />
            </a>
            <a
              href={routes.github()}
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <GitHubIcon className="w-8 h-8" />
            </a>
            <a
              href={routes.discord()}
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <DiscordIcon className="w-8 h-8" />
            </a>
          </div>
        </div>
        <div className="md:space-x-16 flex flex-row justify-between">
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">
              Get started
            </h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href={routes.home()}>Homepage</Link>
              </li>
              <li>
                <Link href={routes.leaderboardLatestOfficial()}>
                  Leaderboard
                </Link>
              </li>
              <li>
                <Link href={routes.contestListOfficial()}>Contests</Link>
              </li>
              <li>
                <Link href={routes.blogList()}>Blog</Link>
              </li>
              <li>
                <Link href={routes.forum()}>Forum</Link>
              </li>
            </ul>
          </div>
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">
              Resources
            </h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href={routes.blogPage('about')}>About</Link>
              </li>
              <li>
                <Link href={routes.blogPage('manual')}>Manual</Link>
              </li>
              <li>
                <Link href={routes.blogPage('rules')}>Rules</Link>
              </li>
              <li>
                <Link href={routes.blogPage('faq')}>FAQ</Link>
              </li>
              <li>
                <Link href="#">Page counter</Link>
              </li>
            </ul>
          </div>
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">Legal</h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href={routes.blogPage('privacy')}>Privacy</Link>
              </li>
              <li>
                <Link href={routes.blogPage('contact')}>Contact</Link>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
