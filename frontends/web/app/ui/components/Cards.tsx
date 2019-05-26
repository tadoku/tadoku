import React from 'react'
import styled from 'styled-components'

export const Card = styled.div`
  flex: 1 1 auto;
  padding: 20px 30px;
  border-radius: 2px;
  box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  margin: 20px;
  max-width: 300px;
  height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
`

export const LargeCard = styled.div`
  flex: 1 1 100%;
  padding: 20px 30px;
  border-radius: 2px;
  box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  margin: 20px;
  display: flex;
  flex-direction: column;
  justify-content: center;
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
  box-sizing: border-box;
  margin: -20px;
  flex-wrap: wrap;
`

const Cards: React.SFC<{}> = ({ children }) => <Container>{children}</Container>

export default Cards
