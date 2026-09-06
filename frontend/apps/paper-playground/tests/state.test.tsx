import { fireEvent, render, screen } from '@testing-library/react'
import { Link, MemoryRouter } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { PlaygroundProvider, usePlayground } from '../src/state'

function Journey(){const app=usePlayground();return <><output aria-label="Viewer">{app.viewer}</output><Link to="/contests/round5/registration">Join the round</Link><button onClick={()=>app.joinContest('round5',['Japanese'])}>Complete registration</button><Link to="/">Home</Link><output aria-label="Joined">{app.joinedContests.join(',')}</output></>}
beforeEach(()=>localStorage.clear())
it('keeps the selected sample account when a scenario follows its registration link',()=>{
  render(<MemoryRouter initialEntries={['/?scenario=member-open']}><PlaygroundProvider><Journey/></PlaygroundProvider></MemoryRouter>)
  expect(screen.getByLabelText('Viewer')).toHaveTextContent('member')
  fireEvent.click(screen.getByRole('link',{name:'Join the round'}))
  expect(screen.getByLabelText('Viewer')).toHaveTextContent('member')
  fireEvent.click(screen.getByRole('button',{name:'Complete registration'}))
  fireEvent.click(screen.getByRole('link',{name:'Home'}))
  expect(screen.getByLabelText('Viewer')).toHaveTextContent('participant')
  expect(screen.getByLabelText('Joined')).toHaveTextContent('round5')
})
