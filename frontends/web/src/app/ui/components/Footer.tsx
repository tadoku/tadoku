import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import { LogoLight } from './index'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import Link from 'next/link'
import { Contest } from '@app/contest/interfaces'

interface Props {
  contests: Contest[]
}

const Footer = (props: Props) => (
  <Container>
    <Background>
      <InnerContainer>
        <FooterContent {...props} />
      </InnerContainer>
    </Background>
  </Container>
)

export default Footer

export const FooterLanding = (props: Props) => (
  <Container>
    <Background>
      <InnerContainer wide>
        <FooterContent {...props} />
      </InnerContainer>
    </Background>
  </Container>
)

const FooterContent = ({ contests }: Props) => (
  <>
    <div>
      <LogoLight />
      <Credits>
        Built by <a href="https://antonve.be">antonve</a>
      </Credits>
      <SocialList>
        <SocialLink fixOffset>
          <a
            href="https://twitter.com/tadoku_app"
            target="_blank"
            rel="noopener noreferrer"
          >
            <FontAwesomeIcon
              icon={['fab', 'twitter-square']}
              size="3x"
              inverse
            />
          </a>
        </SocialLink>
        <SocialLink fixOffset>
          <a
            href="https://github.com/tadoku"
            target="_blank"
            rel="noopener noreferrer"
          >
            <FontAwesomeIcon
              icon={['fab', 'github-square']}
              size="3x"
              inverse
            />
          </a>
        </SocialLink>
        <SocialLink>
          <a
            href="https://discord.gg/Dd8t9WB"
            target="_blank"
            rel="noopener noreferrer"
          >
            <FontAwesomeIcon icon={['fab', 'discord']} size="3x" inverse />
          </a>
        </SocialLink>
      </SocialList>
    </div>
    <Navigation>
      <Menu>
        <MenuHeading>Get started</MenuHeading>
        <Link href="/manual" passHref>
          <a>Manual</a>
        </Link>
        <Link href="/landing" passHref>
          <a>About</a>
        </Link>
        <a href="https://forum.tadoku.app">Forum</a>
        <Link href="/blog" passHref>
          <a>Blog</a>
        </Link>
      </Menu>
      <Menu>
        <MenuHeading>Contests</MenuHeading>
        {contests.map(contest => (
          <Link
            href="/contests/[contest_id]/ranking"
            as={`/contests/${contest.id}/ranking`}
            key={contest.id}
            passHref
          >
            <a>{contest.description}</a>
          </Link>
        ))}
      </Menu>
    </Navigation>
  </>
)

const Container = styled.div`
  box-sizing: border-box;
  display: none;
  height: 250px;
  background-color: ${Constants.colors.dark2};

  ${media.greaterThan('medium')`
      display: block;
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
  `}
`

const Background = styled.div`
  background: url('/img/footer.png') no-repeat center center;
  background-size: cover;
  max-width: 1851px;
  height: 100%;
  margin: 0 auto;
`

const InnerContainer = styled.div`
  max-width: ${Constants.maxWidth};
  display: flex;
  align-items: top;
  justify-content: space-between;
  margin: 0 auto;
  padding: 40px 30px;
  box-sizing: border-box;

  ${({ wide }: { wide?: boolean }) => wide && `padding: 40px 60px;`}
`

const Navigation = styled.nav`
  display: flex;
`

const SocialList = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: top;
`

const SocialLink = styled.li`
  height: 100%;

  ${({ fixOffset }: { fixOffset?: boolean }) =>
    fixOffset && ` a svg {  margin-top: -3px; }`}

  a {
    opacity: 0.8;
    transition: 0.2s opacity;

    &:hover {
      opacity: 1;
    }
  }

  & + & {
    margin-left: 20px;
  }
`

const Credits = styled.p`
  color: ${Constants.colors.light};
  margin: 20px 0 40px;
  padding: 0;

  a {
    display: inline-block;
    border-bottom: 2px solid ${Constants.colors.primary};
    color: ${Constants.colors.light};

    &:hover {
      color: ${Constants.colors.primary};
    }
  }
`

const Menu = styled.div`
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  color: ${Constants.colors.light};

  a {
    color: ${Constants.colors.light};
    line-height: 28px;
    font-size: 16px;

    &:hover {
      color: ${Constants.colors.primary};
    }
  }

  & + & {
    margin-left: 60px;
  }
`

const MenuHeading = styled.h3`
  border-bottom: 2px solid ${Constants.colors.primary};
  font-size: 20px;
  margin: 0 0 10px;
`
