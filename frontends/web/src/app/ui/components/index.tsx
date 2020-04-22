import styled, { css } from 'styled-components'
import Constants from '../Constants'
import React, {
  SFC,
  ButtonHTMLAttributes,
  AnchorHTMLAttributes,
  forwardRef,
  Ref,
} from 'react'
import { IconProp } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import media from 'styled-media-query'

interface ButtonProps {
  primary?: boolean
  large?: boolean
  small?: boolean
  destructive?: boolean
  plain?: boolean
  loading?: boolean
  icon?: IconProp
  alignIconRight?: boolean
}

const buttonStyles = css`
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: transparent;
  box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
    0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  padding: 4px 12px;
  font-size: 1.1em;
  font-weight: 600;
  height: 48px;
  white-space: nowrap;
  line-height: 36px;
  border-radius: 3px;
  box-sizing: border-box;
  margin: 0 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;

  &:hover:not([disabled]),
  &:active:not([disabled]),
  &:focus:not([disabled]) {
    box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.1),
      0px 2px 3px 0px rgba(0, 0, 0, 0.15);
  }

  &:active:not([disabled]),
  &:focus:not([disabled]) {
    outline: none;
    border-color: ${Constants.colors.primary};
  }

  ${({ plain }: ButtonProps) =>
    plain &&
    `
    background-color: transparent;
    box-shadow: none;
    border: none;
    padding: 0;

    &:hover:not([disabled]),
    &:active:not([disabled]),
    &:focus:not([disabled]) {
      box-shadow: none;
      color: ${Constants.colors.primary};
    }

    &:active:not([disabled]),
    &:focus:not([disabled]) {
      text-decoration: underline;
    }
  `}

  ${({ primary }: ButtonProps) =>
    primary &&
    `
    color: ${Constants.colors.light};
    background-color: ${Constants.colors.primary};
    box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.08),
      0px 2px 3px 0px rgba(0, 0, 0, 0.24);

    &:hover:not([disabled]),
    &:active:not([disabled]),
    &:focus:not([disabled]) {
      box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.1),
        0px 2px 3px 0px rgba(0, 0, 0, 0.3),
        0px 2px 6px 2px ${Constants.colors.primaryWithAlpha(0.4)};
    }
  `}

  ${({ large }: ButtonProps) =>
    large &&
    `
    height: 56px;
    font-size: 1.4em;
    padding: 8px 24px;
    border-radius: 4px;
  `}

  ${({ small }: ButtonProps) =>
    small &&
    `
    font-size: 0.9em;
    border-radius: 2px;
  `}

  ${({ destructive }: ButtonProps) =>
    destructive &&
    `
    color: ${Constants.colors.destructive};
    border-color: ${Constants.colors.destructive};
    background-color: transparent;

    &:hover:not([disabled]),
    &:active:not([disabled]),
    &:focus:not([disabled]) {
      color: ${Constants.colors.light};
      background-color: ${Constants.colors.destructive};
      box-shadow: inset 0px 0px 2px rgba(0, 0, 0, 0.1),
        0px 2px 3px 0px rgba(0, 0, 0, 0.15),
        0px 2px 6px 2px ${Constants.colors.destructiveWithAlpha(0.4)};
    }
  `}

  ${({ loading }: ButtonProps) =>
    loading &&
    `
    div {
      width: 30px;
      text-align: center;

      > * {
        animation: fa-spin 1.4s infinite linear;
      }
    }
  `}


  &:disabled {
    opacity: 0.6;
  }
 `

export const Button: SFC<
  ButtonHTMLAttributes<HTMLButtonElement> & ButtonProps
> = ({ icon, alignIconRight, loading, children, ...props }) => (
  <StyledButton loading={loading} {...props}>
    {loading ? (
      <div>
        <FontAwesomeIcon icon="circle-notch" />
      </div>
    ) : (
      <>
        {icon && !alignIconRight && <ButtonIconLeft icon={icon} />}
        {children}
        {icon && alignIconRight && <ButtonIconRight icon={icon} />}
      </>
    )}
  </StyledButton>
)
const StyledButton = styled(
  ({
    primary,
    large,
    small,
    destructive,
    plain,
    loading,
    icon,
    alignIconRight,
    ...props
  }) => <button {...props} />,
)`
  ${buttonStyles}
`

export const ButtonLink: SFC<
  AnchorHTMLAttributes<HTMLAnchorElement> & ButtonProps
> = forwardRef(function buttonLink(
  { icon, alignIconRight, loading, children, ...props },
  ref: Ref<HTMLAnchorElement>,
) {
  return (
    <StyledButtonLink loading={loading} {...props} ref={ref}>
      {loading ? (
        <div>
          <FontAwesomeIcon icon="circle-notch" />
        </div>
      ) : (
        <>
          {icon && !alignIconRight && <ButtonIconLeft icon={icon} />}
          {children}
          {icon && alignIconRight && <ButtonIconRight icon={icon} />}
        </>
      )}
    </StyledButtonLink>
  )
})

const ForwardedStyledButtonLink: SFC<
  AnchorHTMLAttributes<HTMLAnchorElement> &
    ButtonProps & { ref: Ref<HTMLAnchorElement> }
> = forwardRef(function link(
  {
    primary,
    large,
    small,
    destructive,
    plain,
    loading,
    icon,
    alignIconRight,
    ...props
  },
  ref: Ref<HTMLAnchorElement>,
) {
  return <a {...props} ref={ref} />
})
const StyledButtonLink = styled(ForwardedStyledButtonLink)`
  ${buttonStyles}
`

const ButtonIconLeft = styled(FontAwesomeIcon)`
  margin-right: 7px;
  height: 75%;
  width: 75%;
`

const ButtonIconRight = styled(FontAwesomeIcon)`
  margin-left: 7px;
  height: 75%;
  width: 75%;
`

export const ButtonContainer = styled.div`
  display: flex;

  ${({ noMargin }: { noMargin?: boolean }) =>
    noMargin &&
    `
    margin: 0 -5px;
  `}
`

export const StackContainer = styled.div`
  display: flex;
  flex-direction: column;

  > * {
    width: 100%;
    box-sizing: border-box;
    margin-left: 0;
    margin-right: 0;
  }

  > * + * {
    margin-top: 12px;
  }
`

export const PageTitle = styled.h1`
  font-family: 'Merriweather', serif;
  font-size: 30px;

  ${media.lessThan('medium')`
    margin: 0 0 20px 0;
  `}
`

export const Logo = styled.img.attrs(() => ({
  src: '/img/logo.svg',
  alt: 'Tadoku',
}))`
  height: 29px;
  width: 158px;
`

export const LogoLight = styled.img.attrs(() => ({
  src: '/img/logo-light.svg',
  alt: 'Tadoku',
}))`
  height: 29px;
  width: 158px;
`
