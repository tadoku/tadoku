import React from 'react'
import styled from 'styled-components'

import Constants from '../../ui/Constants'
import media from 'styled-media-query'

const ContestSchedule = () => (
  <Container>
    <Title>Contest schedule</Title>
    <BackgroundContainer>
      <Schedule>
        <Contest
          round="1"
          description="January 1st - February 14th (1.5 months)"
        />
        <Contest round="2" description="March 24th - April 7th (2 weeks)" />
        <Contest round="3" description="May 1st - 31st (1 month)" />
        <Contest round="4" description="June 23rd - July 7th (2 weeks)" />
        <Contest round="5" description="July 24th - August 7th (2 weeks)" />
        <Contest round="6" description="September 1st - 30th (1 month)" />
        <Contest
          round="7"
          description="October 24th - November 7th (2 weeks)"
        />
        <Contest
          round="8"
          description="November 23rd - December 7th (2 weeks)"
        />
      </Schedule>
      <Background />
    </BackgroundContainer>
  </Container>
)

export default ContestSchedule

const Container = styled.div`
  width: 100%;
  padding: 60px 60px 0 60px;
  box-sizing: border-box;
`

const BackgroundContainer = styled.div`
  position: relative;
`

const Background = styled.div`
  position: absolute;
  background: #bbb;
  right: 0px;
  bottom: 60px;
  top: 60px;
  padding-left: 520px;
  width: 100%;
  max-width: 680px;
  z-index: -1;
  background: bottom right no-repeat url('/img/background-contests.jpg');

  ${media.lessThan('large')`display: none;`}
`

const Title = styled.h2`
  font-family: ${Constants.fonts.serif};
  margin: 0 0 30px 0;
  font-size: 24px;
  line-height: 31px;
  font-weight: 700;
  letter-spacing: 1.05;

  ${media.lessThan('large')`
    text-align: center;
  `}
`

const Schedule = styled.ul`
  margin: 0;
  padding: 0;
  list-style-type: none;
  position: relative;
  width: 490px;
  box-sizing: border-box;

  &:after {
    content: '';
    z-index: -1;
    position: absolute;
    top: 60px;
    bottom: 60px;
    width: 2px;
    left: calc(50% - 1px);
    background: ${Constants.colors.lightGray};
  }

  ${media.lessThan('large')`
    margin: 0 auto;
  `}

  ${media.lessThan('medium')`
    width: 100%;
  `}
`

interface Props {
  round: string
  description: string
}

const Contest = ({ round, description }: Props) => (
  <ContestRow>
    <h2>Round {round}</h2>
    <p>{description}</p>
  </ContestRow>
)

const ContestRow = styled.li`
  background: #fff;
  box-shadow: 0px 2px 3px 0px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  padding: 60px;
  margin: 0;

  & + & {
    margin-top: 60px;
  }

  h2 {
    margin-bottom: 10px;
    margin-top: 0;
  }

  p {
    margin: 0;
  }

  ${media.lessThan('medium')`
    padding: 30px;

    & + & {
      margin-top: 20px;
    }
  `}
`
