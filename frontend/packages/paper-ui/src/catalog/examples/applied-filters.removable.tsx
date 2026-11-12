import { AppliedFilters, Button } from 'paper-ui'
import { useRef, useState } from 'react'

const initial = ['Japanese', 'Listening', 'audiobook']
// Logs left with 0, 1, 2 or 3 filters applied; a screen derives this from its own results.
const counts = [127, 108, 96, 83]

export default function RemovableFiltersExample() {
  const [applied, setApplied] = useState(initial)
  const filters = useRef<HTMLButtonElement>(null)

  return (
    <div className="grid gap-3">
      <Button ref={filters} variant="ghost" className="justify-self-start" onClick={() => setApplied(initial)}>
        Filters ({applied.length})
      </Button>
      <AppliedFilters
        variant="removable"
        filters={applied.map(label => ({
          id: label,
          label,
          onRemove: () => {
            setApplied(applied.filter(item => item !== label))
            if (applied.length === 1) filters.current?.focus()
          },
        }))}
        onClear={() => {
          setApplied([])
          filters.current?.focus()
        }}
        resultLabel={`${counts[applied.length]} logs of 127`}
      />
    </div>
  )
}
