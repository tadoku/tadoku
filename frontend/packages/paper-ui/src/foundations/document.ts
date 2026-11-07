/**
 * Classes for a Paper application's `<html>` and `<body>`. Paper ships no reset stylesheet: these zero-specificity
 * utilities set the document's surface, type and focus ring, so any component utility overrides them.
 */
export const documentClassName = [
  'bg-canvas text-ink font-body leading-body overflow-y-scroll [text-size-adjust:100%]',
  '[:where(&,&_*,&_::before,&_::after)]:box-border',
  '[:where(&_:is(button,input,select,textarea))]:text-inherit',
  '[:where(&_:is(button,input,select,textarea))]:[font:inherit]',
  '[:where(&_:is(img,svg))]:max-w-full',
  '[:where(&_[hidden])]:hidden!',
  '[:where(&_a)]:underline-offset-link',
  '[:where(&_a:any-link:hover)]:text-link-hover',
  '[:where(&_a:any-link:hover)]:decoration-hover',
  '[:where(&_:focus-visible)]:outline-ring',
  '[:where(&_:focus-visible)]:outline-focus',
  '[:where(&_:focus-visible)]:outline-offset-ring',
  "[:where(&_:is(main,h1,h2)[tabindex='-1']:focus-visible)]:outline-none",
].join(' ')

export const bodyClassName = 'm-0 min-h-full bg-canvas text-ink text-body'
