import styled, { keyframes } from 'styled-components'
import Constants from '../Constants'

interface Props {
  isLoading: boolean
}

const gradient = keyframes`
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
`

const ActivityIndicator = styled.div<Props>`
  background: linear-gradient(
    90deg,
    ${Constants.colors.primary},
    ${Constants.colors.primary},
    ${Constants.colors.primaryTint},
    ${Constants.colors.primary}
  );
  background-size: 400% 400%;
  animation: ${gradient} 2s ease infinite;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 10px;
  opacity: 0;
  transition: 1s opacity ease;
  z-index: 9999;

  @supports (position: -webkit-sticky) or (position: sticky) {
    position: sticky;
  }

  ${({ isLoading }) => isLoading && `opacity: 1;`}
`
export default ActivityIndicator
