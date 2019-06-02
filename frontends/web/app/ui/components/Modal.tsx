import React, { SFC } from 'react'
import ReactModal from 'react-modal'
import styled from 'styled-components'

const modalStyles = {
  content: {
    top: '50%',
    left: '50%',
    right: 'auto',
    bottom: 'auto',
    marginRight: '-50%',
    transform: 'translate(-50%, -50%)',
    border: 0,
    boxShadow: '4px 15px 20px 1px rgba(0, 0, 0, 0.28)',
    padding: '40px',
  },
  overlay: {
    backgroundColor: 'rgba(0, 0, 0, 0.4)',
  },
}

const ModalHeading = styled.h2`
  margin-top: 0;
  text-transform: uppercase;
  text-align: center;
`

const Modal: SFC<ReactModal.Props> = ({ children, style, ...props }) => (
  <ReactModal style={style ? style : modalStyles} {...props}>
    {props.contentLabel && <ModalHeading>{props.contentLabel}</ModalHeading>}
    {children}
  </ReactModal>
)

export default Modal
