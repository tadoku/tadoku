import Link from 'next/link'
import { LogoInverted } from 'ui'
import { DiscordIcon, GitHubIcon, TwitterIcon } from './Icons'

export default function Footer() {
  return (
    <div className="w-full h-60 bg-[url('/img/footer.png')] bg-no-repeat bg-top bg-[#1A1A1A]">
      <div className="max-w-7xl mx-auto p-4 md:p-8 flex justify-between h-full">
        <div className="flex flex-col space-y-4">
          <LogoInverted />
          <div className="text-white flex-grow">
            Built by{' '}
            <a
              href="https://antonve.be"
              className="reset text-white hover:text-primary border-b-2 border-primary transition-all"
            >
              antonve
            </a>
          </div>

          <div className="space-x-4 flex">
            <a
              href="https://twitter.com/tadoku_app"
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <TwitterIcon className="w-8 h-8" />
            </a>
            <a
              href="https://github.com/tadoku"
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <GitHubIcon className="w-8 h-8" />
            </a>
            <a
              href="https://discord.gg/Dd8t9WB"
              target="_blank"
              rel="noopener noreferrer"
              className="reset bg-white/80 hover:bg-primary/80 p-2"
            >
              <DiscordIcon className="w-8 h-8" />
            </a>
          </div>
        </div>
        <div className="flex space-x-16">
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">
              Get started
            </h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href="/">Homepage</Link>
              </li>
              <li>
                <Link href="/leaderboard">Leaderboard</Link>
              </li>
              <li>
                <Link href="/contests">Contests</Link>
              </li>
              <li>
                <Link href="/blog">Blog</Link>
              </li>
              <li>
                <Link href="https://forum.tadoku.app/">Forum</Link>
              </li>
            </ul>
          </div>
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">
              Resources
            </h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href="/pages/manual">Manual</Link>
              </li>
              <li>
                <Link href="/pages/rules">Rules</Link>
              </li>
              <li>
                <Link href="/pages/faq">FAQ</Link>
              </li>
              <li>
                <Link href="#">Page counter</Link>
              </li>
              <li>
                <Link href="#">Text Reader</Link>
              </li>
            </ul>
          </div>
          <div className="flex flex-col items-start">
            <h2 className="text-white border-b-2 border-primary mb-2">Legal</h2>
            <ul className="[&>li>a]:text-white space-y-1">
              <li>
                <Link href="/pages/privacy">Privacy</Link>
              </li>
              <li>
                <Link href="/pages/contact">Contact</Link>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
