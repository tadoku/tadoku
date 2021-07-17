import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../Constants'
import LinkContainer from '@app/ui/components/navigation/LinkContainer'
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
    <BrandingContainer>
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
    </BrandingContainer>
    <Navigation>
      <Menu>
        <MenuHeading>Get started</MenuHeading>
        <LinkContainer dark>
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
        </LinkContainer>
      </Menu>
      <Menu>
        <MenuHeading>Contests</MenuHeading>
        <LinkContainer dark>
          {contests.map(contest => (
            <Link
              href={`/contests/${contest.id}/ranking`}
              key={contest.id}
              passHref
            >
              <a>{contest.description}</a>
            </Link>
          ))}
          <Link href="/contests" passHref>
            <a>Archive</a>
          </Link>
        </LinkContainer>
      </Menu>
    </Navigation>
  </>
)

const BrandingContainer = styled.div`
  ${media.lessThan('medium')`
    text-align: center;
  `}
`

const Container = styled.div`
  box-sizing: border-box;
  background-color: ${Constants.colors.dark2};
  height: 250px;
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;

  ${media.lessThan('medium')`
    height: inherit;
    position: inherit;
    left: inherit;
    position: inherit;
    position: inherit;
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
  flex-direction: row;
  align-items: top;
  justify-content: space-between;
  margin: 0 auto;
  box-sizing: border-box;
  padding: 40px 30px;
  ${({ wide }: { wide?: boolean }) => wide && `padding: 40px 60px;`}

  ${media.lessThan('medium')`
    flex-direction: column;
    padding: 20px 0;
  `}
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

  ${media.lessThan('medium')`
    display: none;
  `}
`

const Navigation = styled.nav`
  display: flex;
  flex-direction: row;

  ${media.lessThan('medium')`
    flex-direction: column;
  `}
`

const SocialList = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: top;

  ${media.lessThan('medium')`
    display: none;
  `}
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

const Menu = styled.div`
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  color: ${Constants.colors.light};
  flex: 1;

  & + & {
    margin-left: 60px;
  }

  & > div {
    flex-direction: column;

    & > a {
      margin: 0;
      font-weight: normal;
      font-size: 16px;
      line-height: 26px;
      color: ${Constants.colors.light};
    }
  }

  ${media.lessThan('medium')`
    margin-top: 20px;

    & + & {
      margin-left: 0;
    }

    & > div > a {
      font-size: 16px;
      line-height: 48px;
    }
 `}
`

const MenuHeading = styled.h3`
  font-size: 20px;
  margin: 0 0 10px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.primary};

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 10px 30px;
    background-color: ${Constants.colors.lightWithAlpha(0.04)};
    width: 100%;
  `}
`
