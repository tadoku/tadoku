import styled from 'styled-components'
import Constants from '../../../ui/Constants'

const HintContainer = styled.div`
  background: ${Constants.colors.darkWithAlpha(0.9)};
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  color: ${Constants.colors.light};
  padding: 8px 12px;
  border-radius: 0;
`

export default HintContainer
