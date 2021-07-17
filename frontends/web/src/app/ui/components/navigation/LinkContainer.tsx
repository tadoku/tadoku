import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../../Constants'

interface Props {
  dark?: boolean
}

const LinkContainer = styled.div<Props>`
  display: flex;
  padding-right: 20px;
  width: 100%;

  > * + * {
    margin-left: 20px;

    ${media.lessThan('medium')`
      margin-left: 0;
    `}
  }

  > a,
  > button {
    font-weight: bold;
    font-family: ${Constants.fonts.sansSerif};
  }

  ${media.lessThan<Props>('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;

    > a,
    > button {
      margin-left: 0;
      padding-left: 30px;
      border-top: 1px solid ${({ dark }) =>
        dark
          ? Constants.colors.lightWithAlpha(0.08)
          : Constants.colors.lightGray};
      display: block;
      text-align: left;
      align-items: flex-start;
      line-height: 48px;

      &:focus,
      &:active,
      &:hover {
        background-color: ${Constants.colors.lightWithAlpha(0.05)};
        border-color: ${({ dark }) =>
          dark
            ? Constants.colors.lightWithAlpha(0.08)
            : Constants.colors.lightGray} !important;
      }

      &:after {
        bottom: 0 !important;
        left: 10px !important;
        right: inherit !important;
        top: 0 !important;
        width: 7px;
        border: none;
        background-color: ${Constants.colors.primary};
      }
    }
  `}
`

export default LinkContainer
