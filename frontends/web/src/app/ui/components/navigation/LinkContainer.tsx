import styled from 'styled-components'
import media from 'styled-media-query'
import Constants from '../../Constants'

const LinkContainer = styled.div`
  display: flex;
  padding-right: 20px;

  > * + * {
    margin-left: 20px;

    ${media.lessThan('medium')`
      margin-left: 0;
    `}
  }

  > * {
    ${media.lessThan('medium')`
      margin-left: 0;
      padding-left: 30px;
      border-top: 1px solid ${Constants.colors.lightGray};
      display: block;
      text-align: left;
      align-items: center;
    `}
  }

  > a {
    ${media.lessThan('medium')`
      line-height: 48px;
    `}
  }

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;
  `}
`

export default LinkContainer
