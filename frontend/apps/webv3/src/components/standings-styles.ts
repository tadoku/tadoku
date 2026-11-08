import { listFooter } from 'paper-ui'
import { cardTable } from './HomePage'

/** The standings screens: a sheet holding period tabs, the board and its pagination, beside an aside. */
export const standingsHeader = 'standings-header mb-5'
export const workspace = [
  'mt-5 grid grid-cols-workspace items-start gap-8 *:min-w-0',
  'max-lg:grid-cols-workspace-narrow max-lg:gap-6 max-tablet:grid-cols-single',
].join(' ')
export const sheet = 'min-w-0 border border-rule bg-paper'
/** Inset the tabs, not the tab bar, so its rule runs edge to edge across the sheet. */
export const sheetTabs = { list: 'px-5 max-md:px-4' }
export const sheetEmpty = 'm-0 border-t border-rule px-5 py-6 max-md:px-4'
export const standingsTable = [
  `standings-table tabular-nums ${cardTable}`,
  '[&_td]:px-5 [&_th]:px-5 max-md:[&_td]:px-4 max-md:[&_th]:px-4',
].join(' ')
export const standingsPagination = listFooter().root({
  className: 'standings-pagination border-t border-rule px-5 py-3 max-md:px-4',
})
export const paginationSummary = listFooter().summary()
export const standingsAside = [
  'min-w-0 pt-1 text-sm leading-170 wrap-anywhere [&_p]:mb-4 [&_p]:mt-0 [&>p]:text-muted',
  '[&>section+section]:mt-7 [&>section+section]:border-t [&>section+section]:border-rule [&>section+section]:pt-6',
  '[&_h2]:mb-3 [&_h2]:mt-0 [&_h2]:font-display [&_h2]:text-component [&_h2]:leading-component [&_h2]:font-bold',
  '[&_.paper-description-list+.text-link]:mt-4 [&_.paper-description-list+.text-link]:inline-block',
].join(' ')
export const standingsNotice = 'mt-5'
export const standingsContext = 'm-0 gap-y-0 p-5 max-md:px-4 max-md:pb-4 max-md:pt-5'
export const standingsContextTitle = 'text-heading max-md:text-lg'
export const description = 'whitespace-pre-line'
/** A form or record on the paper surface, one column wide; it fills the board column beside an aside. */
export const recordSurface = 'max-w-record p-record-inset wrap-anywhere'
export const muted = 'mb-0 mt-2 text-metadata leading-160 text-muted'
