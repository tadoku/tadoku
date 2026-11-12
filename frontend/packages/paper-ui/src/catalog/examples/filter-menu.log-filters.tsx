import { FilterMenu, type FilterMenuSection } from 'paper-ui'
import { useState } from 'react'

const activities = ['Reading', 'Listening', 'Writing', 'Speaking']
const tags = [
  'audiobook', 'novel', 'manga', 'podcast', 'news', 'graded reader',
  'drama', 'dense', 'light novel', 'game', 'textbook', 'blog',
]
// 127 deterministic logs, one activity and one tag each.
const logs = Array.from({ length: 127 }, (_, index) => ({
  activity: activities[index % 3 === 0 ? 1 : index % 7 === 0 ? 2 : 0],
  tag: tags[(index * 5 + Math.floor(index / 12)) % tags.length],
  points: 4 + ((index * 37) % 46),
}))

type Dimension = 'activity' | 'tag'
type Selection = Record<Dimension, readonly string[]>

const allows = (selection: Selection, log: (typeof logs)[number], except?: Dimension) =>
  (['activity', 'tag'] as const).every(
    id => id === except || selection[id].length === 0 || selection[id].includes(log[id]),
  )

function section(id: Dimension, label: string, values: readonly string[], selection: Selection): FilterMenuSection {
  return {
    id,
    label,
    selected: selection[id],
    values: values.map(value => ({
      value,
      label: value,
      points: logs
        .filter(log => log[id] === value && allows(selection, log, id))
        .reduce((total, log) => total + log.points, 0),
    })),
  }
}

export default function FilterMenuExample() {
  const [selection, setSelection] = useState<Selection>({ activity: ['Reading'], tag: [] })
  const count = logs.filter(log => allows(selection, log)).length

  return (
    <FilterMenu
      sections={[
        section('activity', 'Activity', activities, selection),
        section('tag', 'Tags', tags, selection),
      ]}
      onChange={(id, values) => setSelection({ ...selection, [id]: values })}
      onClear={() => setSelection({ activity: [], tag: [] })}
      resultLabel={`${count} ${count === 1 ? 'log' : 'logs'}`}
    />
  )
}
