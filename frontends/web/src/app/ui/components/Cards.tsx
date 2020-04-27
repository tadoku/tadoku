import React from 'react'
import styled from 'styled-components'
import Constants from '../Constants'

const Cards: React.SFC<{}> = ({ children }) => <Container>{children}</Container>

export default Cards

export const Card = styled.div`
  padding: 20px 30px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
`

export const LargeCard = styled.div`
  padding: 30px;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  background: ${Constants.colors.light};
`

export const CardContent = styled.div`
  font-size: 3em;
  text-align: center;
`

const Container = styled.div`
  display: flex;
  align-items: center;
  box-sizing: border-box;
  margin: -20px;
  flex-wrap: wrap;
`
