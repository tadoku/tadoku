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
  active?: boolean
}

const buttonStyles = css`
  border: 1px solid ${Constants.colors.darkWithAlpha(0.2)};
  border-bottom-width: 3px;
  background: transparent;
  padding: 4px 12px;
  font-size: 1.1em;
  font-weight: 600;
  height: 48px;
  white-space: nowrap;
  line-height: 36px;
  box-sizing: border-box;
  margin: 0 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;

  &:hover:not([disabled]),
  &:active:not([disabled]),
  &:focus:not([disabled]) {
    border-color: ${Constants.colors.darkWithAlpha(0.4)};
  }

  &:active:not([disabled]),
  &:focus:not([disabled]) {
    outline: none;
    border-color: ${Constants.colors.primary};
  }

  ${({ plain, active }: ButtonProps) =>
    plain &&
    `
    background-color: transparent;
    box-shadow: none;
    border: none;
    padding: 0;

    &:hover:not([disabled]),
    &:active:not([disabled]),
    &:focus:not([disabled]) {
      color: ${Constants.colors.primary};
      position: relative;
      box-shadow: none;

      &:after {
        content: '';
        position: absolute;
        bottom: 10px;
        left: 0;
        right: 0;
        border-bottom: 2px solid ${Constants.colors.primary};
      }
    }

    ${
      active &&
      `
      color: ${Constants.colors.primary};
      position: relative;

      &:after {
        content: '';
        position: absolute;
        bottom: 10px;
        left: 0;
        right: 0;
        border-bottom: 2px solid ${Constants.colors.primary};
      }

      &:hover {
        opacity: 0.7;
      }
    `
    }
  `}

  ${({ primary }: ButtonProps) =>
    primary &&
    `
    color: ${Constants.colors.light};
    background-color: ${Constants.colors.primary};

    &:hover:not([disabled]),
    &:active:not([disabled]),
    &:focus:not([disabled]) {
    }
  `}

  ${({ large }: ButtonProps) =>
    large &&
    `
    height: 56px;
    font-size: 1.4em;
    padding: 8px 24px;
  `}

  ${({ small }: ButtonProps) =>
    small &&
    `
    font-size: 0.9em;
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
    active,
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
  font-size: 30px;

  ${media.lessThan('medium')`
    margin: 0 0 20px 0;
  `}
`

export const SubHeading = styled.h2`
  font-family: ${Constants.fonts.sansSerif};
  color: ${Constants.colors.nonFocusText};
  font-size: 17px;
  text-transform: uppercase;
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
