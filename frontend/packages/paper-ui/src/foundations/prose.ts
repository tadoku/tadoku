import { tv } from './variants'

/**
 * Long-form HTML that Paper does not render itself: CMS pages, the manual and Markdown posts. It styles the elements
 * the sanitizers keep, so the HTML needs no classes.
 */
export const prose = tv({
  base: [
    'paper-prose text-lead leading-180 wrap-break-word [&>:first-child]:mt-0',
    '[&_:is(p,ul,ol,blockquote,pre)]:mb-4 [&_:is(p,ul,ol,blockquote,pre)]:mt-0',
    '[&_:is(p,ul,ol,blockquote,pre)]:max-w-measure-prose',
    '[&_:is(h2,h3,h4)]:scroll-mt-20 [&_:is(h2,h3,h4)]:font-display [&_:is(h2,h3,h4)]:leading-150',
    '[&_h2]:mb-3 [&_h2]:mt-12 [&_h2]:text-heading [&_h2]:leading-section',
    '[&_h3]:mb-2 [&_h3]:mt-8 [&_h3]:text-component',
    '[&_h4]:mb-2 [&_h4]:mt-6 [&_h4]:text-lead',
    '[&_:is(ul,ol)]:ps-5.5 [&_li+li]:mt-2 [&_a]:text-link [&_sup]:text-link',
    '[&_blockquote]:border-s-edge [&_blockquote]:border-rule [&_blockquote]:ps-5 [&_blockquote]:text-muted',
    '[&_code]:font-mono [&_code]:text-code',
    '[&_pre]:overflow-x-auto [&_pre]:border [&_pre]:border-rule [&_pre]:bg-paper [&_pre]:p-4',
    '[&_img]:h-auto [&_img]:max-w-full',
    '[&_hr]:my-8 [&_hr]:border-0 [&_hr]:border-t [&_hr]:border-rule',
    '[&_table]:w-full [&_table]:border-collapse [&_table]:text-sm [&_table]:leading-150 [&_table]:tabular-nums',
    '[&_:is(th,td)]:border-b [&_:is(th,td)]:border-rule [&_:is(th,td)]:py-2.5 [&_:is(th,td)]:pe-4 [&_:is(th,td)]:ps-0',
    '[&_:is(th,td)]:text-left [&_:is(th,td)]:align-top',
    '[&_th]:border-b-rule-strong [&_th]:font-semibold [&_th]:text-ink',
  ],
})
