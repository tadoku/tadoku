import { useState } from 'react'
import { Navbar } from 'paper-ui'

const navbarNavigation = [
  { type: 'link', id: 'home', label: 'Home', href: '#home' },
  { type: 'link', id: 'contests', label: 'Contests', href: '#contests' },
  {
    type: 'dropdown',
    id: 'guides',
    label: 'Guides',
    links: [
      { id: 'start', label: 'Start here', href: '#start' },
      { id: 'scoring', label: 'Scoring and rules', href: '#scoring' },
    ],
  },
] as const

export default function NavbarExample() {
  const [path, setPath] = useState('#home')
  return (
    <div className="paper-stack">
      <Navbar
        brand="Tadoku"
        brandHref="#home"
        currentPath={path}
        navigation={navbarNavigation}
        mobileFooter={closeMenu => (
          <a
            className="paper-navbar__mobile-link"
            href="#profile"
            aria-current={path === '#profile' ? 'page' : undefined}
            onClick={event => {
              event.preventDefault()
              setPath('#profile')
              closeMenu()
            }}
          >
            Your profile
          </a>
        )}
        renderLink={props => (
          <a
            {...props}
            onClick={event => {
              props.onClick?.(event)
              if (!event.defaultPrevented) setPath(props.href)
            }}
          />
        )}
      />
      <p role="status">
        Selected destination: {path.slice(1)}. These fragment links keep this
        example on the page.
      </p>
      <h3>Route loading</h3>
      <Navbar
        brand="Tadoku"
        brandHref="#home"
        navigation={[]}
        mobileNavigation={false}
        isLoading
      />
    </div>
  )
}
