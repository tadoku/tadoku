import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const navbarSource = await readFile(
  new URL('../components/Navbar.tsx', import.meta.url),
  'utf8',
)
const actionMenuSource = await readFile(
  new URL('../components/ActionMenu.tsx', import.meta.url),
  'utf8',
)

const mobilePanelStart = navbarSource.indexOf(
  '<DisclosurePanel className="sm:hidden">',
)
const mobilePanelEnd = navbarSource.indexOf(
  '</DisclosurePanel>',
  mobilePanelStart,
)
const mobilePanelSource = navbarSource.slice(mobilePanelStart, mobilePanelEnd)

test('mobile navigation exposes dropdown links directly in the drawer', () => {
  assert.ok(mobilePanelStart >= 0, 'expected a mobile disclosure panel')
  assert.ok(mobilePanelEnd > mobilePanelStart, 'expected the panel to close')
  assert.doesNotMatch(mobilePanelSource, /<DropDownMobile/)
  assert.match(mobilePanelSource, /item\.links\.map/)
})

test('mobile navigation supports primary and bottom account actions', () => {
  assert.match(navbarSource, /mobilePrimary/)
  assert.match(navbarSource, /mobileBottom/)
  assert.match(mobilePanelSource, /mt-auto/)
})

test('top mobile action uses the drawer link style instead of a primary button', () => {
  assert.doesNotMatch(mobilePanelSource, /btn primary/)
})

test('mobile sections have consistent rows and a highlighted top action', () => {
  assert.match(mobilePanelSource, />\s*Navigation\s*</)
  assert.match(navbarSource, /const MobileDropDownLink/)
  assert.match(mobilePanelSource, /highlighted/)
  assert.match(navbarSource, /bg-secondary\/5/)
  assert.doesNotMatch(
    navbarSource.slice(navbarSource.indexOf('const MobileDropDownLink')),
    /text-primary/,
  )
  assert.doesNotMatch(navbarSource, /dividerBefore/)
  assert.doesNotMatch(navbarSource, /dividerAfter/)
  assert.doesNotMatch(mobilePanelSource, /!!divider/)
  assert.match(mobilePanelSource, /bg-white p-2 shadow-lg/)
  assert.doesNotMatch(mobilePanelSource, /pb-6/)
  assert.doesNotMatch(
    mobilePanelSource,
    /mt-auto border-t border-slate-500\/20 pt-3/,
  )

  const sharedRowUses = mobilePanelSource.match(/<MobileDropDownLink/g) ?? []
  assert.ok(
    sharedRowUses.length >= 3,
    'expected top, account, and bottom actions to share one row component',
  )
})

test('mobile drawer avoids hydration and containing-block boundaries', () => {
  assert.doesNotMatch(navbarSource, /<Portal>/)
  assert.doesNotMatch(navbarSource, /backdrop-blur/)
})

test('opening the mobile drawer locks and restores page scrolling', () => {
  assert.match(navbarSource, /<MobileScrollLock active={open} \/>/)
  assert.match(navbarSource, /const MobileScrollLock/)
  assert.match(navbarSource, /document\.body\.style\.overflow/)
  assert.match(navbarSource, /window\.scrollTo/)
})

test('logged-out navigation does not render an empty bottom divider', () => {
  assert.match(navbarSource, /const hasMobileBottomLinks/)
  assert.match(mobilePanelSource, /{hasMobileBottomLinks \? \(/)
})

test('shared navigation uses light hover and focus states', () => {
  assert.doesNotMatch(
    navbarSource,
    /text-secondary hover:bg-secondary hover:text-white/,
  )
  assert.match(navbarSource, /text-secondary hover:bg-secondary\/5/)
  assert.match(navbarSource, /data-\[focus\]:bg-secondary\/5/)
  assert.doesNotMatch(
    navbarSource,
    /data-\[focus\]:bg-secondary data-\[focus\]:text-white/,
  )

  assert.match(actionMenuSource, /data-\[focus\]:bg-secondary\/5/)
  assert.doesNotMatch(
    actionMenuSource,
    /'data-\[focus\]:bg-secondary': type === 'normal'/,
  )
  assert.match(actionMenuSource, /data-\[focus\]:bg-red-700\/80/)
})
