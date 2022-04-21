import React, { useEffect, useState } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Contest, ContestStats } from '../../interfaces'
import Modal from '@app/ui/components/Modal'
import { Group, Label, LabelText, TextArea } from '@app/ui/components/Form'
import { Button, StackContainer } from '@app/ui/components'
import ContestApi from '../../api'
import RankingApi from '@app/ranking/api'
import { Ranking } from '@app/ranking/interfaces'
import { formatLanguageName } from '@app/ranking/transform/format'

function generateBlogPostSkeleton(stats: ContestStats, ranking: Ranking[]) {
  const winner = ranking[0]

  return (
    renderToStaticMarkup(
      <>
        <p>
          Congrats to <strong>{winner.userDisplayName}</strong> for winning the
          round with a total score of{' '}
          <strong>{winner.amount.toLocaleString()}</strong>! In total{' '}
          <strong>{stats.participants}</strong>{' '}
          {stats.participants === 1 ? 'person' : 'people'} participated in this
          round for a total score of{' '}
          <strong>{stats.totalAmount.toLocaleString()}</strong>!
        </p>

        <table>
          <tbody>
            {stats.byLanguage
              .filter(({ languageCode }) => languageCode !== 'GLO')
              .map(({ languageCode, count }) => (
                <tr key={languageCode}>
                  <th scope="row">{formatLanguageName(languageCode)}</th>
                  <td>{`${count} reader${count === 1 ? '' : 's'}`}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </>,
    )
      // posts are stored in a JSON file, the quotes need to be either esacaped or single quotes
      .replace(/"/g, '\\"')
  )
}

const ViewStatsModal = ({
  contest,
  onCancel,
}: {
  contest: Contest | undefined
  onCancel: () => void
}) => {
  const [statsHtml, setStatsHtml] = useState('')

  useEffect(() => {
    if (!contest) {
      return
    }
    Promise.all([
      ContestApi.getStats(contest.id),
      RankingApi.get(contest.id),
    ]).then(([stats, ranking]) => {
      if (stats && ranking.length) {
        setStatsHtml(generateBlogPostSkeleton(stats, ranking))
      } else {
        setStatsHtml('No stats found for this contest')
      }
    })
  }, [contest])

  return (
    <Modal
      isOpen={!!contest}
      onRequestClose={() => onCancel()}
      contentLabel={`Contest ${contest?.description ?? ''} Stats`}
    >
      <Group>
        <Label>
          <LabelText>Blog Post Skeleton</LabelText>
          <TextArea readOnly value={statsHtml} />
        </Label>
      </Group>
      <Group>
        <StackContainer>
          <Button type="button" onClick={onCancel}>
            Cancel
          </Button>
        </StackContainer>
      </Group>
    </Modal>
  )
}

export default ViewStatsModal
