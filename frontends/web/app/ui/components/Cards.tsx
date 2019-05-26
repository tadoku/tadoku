import React from 'react'
import styled from 'styled-components'

export const Card = styled.div`
  flex: 1;
  padding: 20px 30px;
  border-radius: 2px;
  box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  margin: 0 20px;
`

export const CardContent = styled.div`
  font-size: 3em;
  text-align: center;
`

export const CardLabel = styled.span`
  text-transform: uppercase;
  opacity: 0.6;
  font-size: 0.8em;
  display: block;
  text-align: center;
`

const Container = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;

  > :first-child {
    margin-left: 0;
  }

  > :last-child {
    margin-right: 0;
  }
`

const Cards: React.SFC<{}> = ({ children }) => <Container>{children}</Container>

export default Cards
