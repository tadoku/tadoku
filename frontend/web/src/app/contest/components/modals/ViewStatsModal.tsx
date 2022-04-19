import React, { useEffect, useState } from 'react'
import { Contest, ContestStats } from '../../interfaces'
import Modal from '@app/ui/components/Modal'
import { Group, Label, LabelText, TextArea } from '@app/ui/components/Form'
import { Button, StackContainer } from '@app/ui/components'
import ContestApi from '../../api'
import RankingApi from '@app/ranking/api'
import { Ranking } from '@app/ranking/interfaces'
import { formatLanguageName } from '@app/ranking/transform/format'

function generateBlogPostSkeleton(stats: ContestStats, ranking: Ranking[]) {
  const text = (textContent: string) => document.createTextNode(textContent)
  const el = (
    tagName: string,
    textContent: string | number,
    children: (HTMLElement | Text)[] = [],
    attributes?: Record<string, string>,
  ) => {
    const el = document.createElement(tagName)
    el.textContent = '' + textContent
    for (const child of children) {
      el.appendChild(child)
    }
    if (attributes) {
      for (const attr in attributes) {
        if (attributes.hasOwnProperty(attr)) {
          el.setAttribute(attr, attributes[attr])
        }
      }
    }
    return el
  }

  const winner = ranking[0]

  return el('div', '', [
    el('p', '', [
      text('Congrats to '),
      el('strong', winner.userDisplayName),
      text(' for winning the round with a total score of '),
      el('strong', winner.amount),
      text('! In total '),
      el('strong', stats.participants),
      text(
        ` ${
          stats.participants === 1 ? 'person' : 'people'
        } participated in this round for a total score of `,
      ),
      el('strong', stats.totalPages),
      text('!'),
    ]),
    el('table', '', [
      el(
        'tbody',
        '',
        stats.byLanguage
          .filter(({ languageCode }) => languageCode !== 'GLO')
          .map(({ count, languageCode }) => {
            return el('tr', '', [
              el('th', formatLanguageName(languageCode), [], { scope: 'row' }),
              el('td', `${count} reader${count === 1 ? '' : 's'}`, []),
            ])
          }),
      ),
    ]),
  ]).innerHTML
}

const EditContestFormModal = ({
  contest,
  onCancel,
}: {
  contest: Contest | undefined
  onSuccess: () => void
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

export default EditContestFormModal
