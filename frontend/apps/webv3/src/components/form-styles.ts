/** Webv3's form layouts on the paper surface. */
export const formControls = 'm-0 grid min-w-0 gap-6 border-0 p-0'
export const formRow = 'grid grid-cols-halves items-start gap-5 max-sm:grid-cols-single'
export const formNote = '-mt-3 mb-0 text-metadata text-muted'
export const formActions = 'mt-1 flex flex-wrap items-center gap-x-5 gap-y-3'
/** A titled group of fields, ruled off from the fields above it. */
export const formGroup = [
  'mx-0 mb-0 mt-2 grid min-w-0 gap-4 border-0 border-t border-rule px-0 pb-0 pt-6 [&>p]:m-0',
  '[&>legend]:float-start [&>legend]:w-full [&>legend]:p-0 [&>legend]:font-display [&>legend]:text-heading',
  '[&>legend]:leading-section [&>legend]:font-bold [&>legend]:text-ink',
].join(' ')
