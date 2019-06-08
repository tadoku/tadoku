import React, { SFC } from 'react'
import ReactModal from 'react-modal'
import styled from 'styled-components'

const Modal: SFC<ReactModal.Props> = ({ children, ...props }) => (
  <StyledModal
    style={{
      overlay: {
        backgroundColor: 'rgba(0, 0, 0, 0.4)',
      },
    }}
    {...props}
  >
    {props.contentLabel && <ModalHeading>{props.contentLabel}</ModalHeading>}
    {children}
  </StyledModal>
)

export default Modal

const ModalHeading = styled.h2`
  margin-top: 0;
  text-transform: uppercase;
  text-align: center;
`

const StyledModal = styled(ReactModal)`
  // positioning
  position: absolute;
  top: 50%;
  left: 50%;
  right: auto;
  bottom: auto;
  margin-right: -50%;
  transform: translate(-50%, -50%);

  // responsiveness
  overflow: auto;
  min-width: 200px;
  max-width: 100vh;
  max-height: calc(100vh + 4px);
  box-sizing: border-box;
  -webkit-overflow-scrolling: touch;

  // styling
  background: #fff;
  outline: none;
  border: 0;
  box-shadow: 4px 15px 20px 1px rgba(0, 0, 0, 0.28);
  padding: 40px;
  border-radius: 4px;
`
