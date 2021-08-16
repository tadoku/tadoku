import styled from 'styled-components'

import Constants from '@app/ui/Constants'
import media from 'styled-media-query'

export const Table = styled.table`
  padding: 0;
  width: 100%;
  background: ${Constants.colors.light};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  border-collapse: collapse;
`

export const TableHeading = styled.tr`
  height: 55px;
  font-size: 16px;
  font-weight: bold;
  text-transform: uppercase;
  color: ${Constants.colors.nonFocusText};
`

export const TableHeadingCell = styled.td`
  padding: 0 30px;
  box-sizing: border-box;
  border-bottom: 2px solid ${Constants.colors.nonFocusTextWithAlpha(0.2)};

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`

export const Row = styled.tr<{ fontSize?: string }>`
  height: 55px;
  padding: 0;
  font-size: ${({ fontSize }) => (fontSize ? fontSize : '20px')};
  font-weight: bold;
  transition: background 0.1s ease;

  &:nth-child(2n + 1) {
    background-color: ${Constants.colors.nonFocusTextWithAlpha(0.05)};
  }
`

export const ClickableRow = styled(Row)`
  &:hover,
  &:active,
  &:focus {
    background: ${Constants.colors.primary};

    a {
      color: ${Constants.colors.light};
      transition: none;
    }
  }
`

export const RowAnchor = styled.a`
  display: block;
  padding: 0;
  height: 55px;
  line-height: 55px;

  &:hover,
  &:active,
  &:focus {
    color: inherit;
  }
`

export const Cell = styled.td`
  height: 55px;
  padding: 0 30px;

  ${media.lessThan('large')`
    padding: 0 20px;
  `}
`
