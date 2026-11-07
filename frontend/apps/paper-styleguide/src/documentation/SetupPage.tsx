import { Link } from 'react-router-dom'
import { CodeBlock } from './CodeBlock'
import { TableOfContents } from './TableOfContents'
import { accentRail, cn, focusRing, typography } from 'paper-ui'

const documentSection = 'document-section max-w-docs-section scroll-mt-24 [&+&]:mt-12 [&>h2]:mb-5 [&>h2]:mt-0 [&>h3]:mb-2.5 [&>h3]:mt-7 [&>h2+h3]:mt-0 [&>p]:mb-4 [&>p]:mt-0'

export function SetupPage() {
  return (
    <div className="grid grid-cols-document justify-center gap-docs-y max-lg:grid-cols-document-narrow">
      <article className="document-page">
        <header className={cn(accentRail(), 'mb-16 ps-4 [&_h1]:mb-4 [&_h1]:mt-1.2')}>
          <p className="m-0 text-metadata font-semibold tracking-caps text-link uppercase">Getting started</p>
          <h1 className={typography({ role: 'page' })}>Setup</h1>
          <p className="text-component leading-165 text-muted">
            Install once, add Tailwind CSS with Paper’s theme, then compose
            pages from token utilities and Paper components.
          </p>
        </header>
        <section id="install" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Add the workspace package</h2>
          <p>
            Paper is a private package in this monorepo. Add it to your
            application’s dependencies using pnpm. Its peer dependencies are
            React and React DOM 18.2 and TanStack Form 1.33.2.
          </p>
          <CodeBlock
            language="bash"
            code={'pnpm --filter your-app add "paper-ui@workspace:*"'}
          />
        </section>
        <section id="tailwind" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Add Tailwind CSS and Paper’s theme</h2>
          <p>
            Paper styles everything with Tailwind CSS v4 utilities generated from
            its typed tokens in <code>paper-ui/src/tokens</code>. Add Tailwind and
            its Vite plugin, then give the application one entry stylesheet.
          </p>
          <CodeBlock language="bash" code={'pnpm --filter your-app add -D tailwindcss @tailwindcss/vite'} />
          <CodeBlock language="javascript" code={'// vite.config.ts\nimport tailwindcss from "@tailwindcss/vite";\n\nexport default defineConfig({ plugins: [react(), tailwindcss()] });'} />
          <CodeBlock language="css" code={'/* src/styles.css: the only stylesheet an application writes */\n@layer theme, base, components, utilities;\n@import "tailwindcss/theme.css" layer(theme);\n@import "tailwindcss/utilities.css" layer(utilities);\n@import "paper-ui/theme.css";\n@source "./";'} />
          <p>
            <code>paper-ui/theme.css</code> replaces Tailwind’s default theme with
            Paper’s tokens, so only token-backed utilities such as{' '}
            <code>p-4</code>, <code>bg-paper</code> and <code>text-muted</code>{' '}
            exist. It also registers Paper’s components as a Tailwind source,
            declares the <code>dark:</code> and <code>compact:</code> variants and
            loads the fonts. Tailwind’s Preflight reset stays off: Paper’s
            document classes set the page defaults instead.
          </p>
          <p>
            Keep complete class names in source, such as <code>p-4</code>;
            Tailwind cannot detect a constructed string like{' '}
            <code>{'`p-${size}`'}</code>. The <code>paper/no-arbitrary-values</code>{' '}
            ESLint rule from <code>paper-ui/eslint-plugin</code> rejects values
            outside the tokens, such as <code>p-[13px]</code>.
          </p>
        </section>
        <section id="stylesheet" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Set up the document</h2>
          <p>
            Import the entry stylesheet once and give the document Paper’s
            surface, type and focus defaults. Set theme and density on{' '}
            <code>&lt;html&gt;</code> so portalled menus and dialogs inherit them.
          </p>
          <CodeBlock
            code={
              'import { bodyClassName, documentClassName } from "paper-ui";\nimport "./styles.css";\n\ndocument.documentElement.className = documentClassName;\ndocument.body.className = bodyClassName;\ndocument.documentElement.dataset.theme = "light";\ndocument.documentElement.dataset.density = "comfortable";'
            }
          />
          <p>
            Use <code>data-theme="dark"</code> for dark appearance and{' '}
            <code>data-density="compact"</code> for compact controls. Your
            application owns the preference and its persistence.
          </p>
          <p>
            Semantic colors switch with <code>data-theme</code>; a{' '}
            <code>bg-paper</code> surface needs no separate <code>dark:</code>{' '}
            utility. Fixed spacing such as <code>p-4</code> stays the same in
            both densities. Components and <code>gap-inline</code> follow density.
          </p>
        </section>
        <section id="compose" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>
            Choose classes or React components
          </h2>
          <p>
            Use token utilities for layout and spacing. Each component’s recipe
            is also exported as a function, so an element Paper does not render
            can share its appearance. React components add behavior,
            accessibility and typed props. For example, a navigation link is an
            anchor with <code>buttonClassName()</code>; an action uses{' '}
            <code>Button</code>.
          </p>
          <CodeBlock
            code={
              'import { Button, buttonClassName } from "paper-ui";\n\nexport function ReadingActions() {\n  return (\n    <div className="flex flex-wrap items-center gap-2 bg-paper p-4">\n      <Button onClick={() => window.print()}>Print log</Button>\n      <a className={buttonClassName({ variant: "outline" })} href="/logs">View logs</a>\n    </div>\n  );\n}'
            }
          />
          <p>
            Start with the class references for{' '}
            <Link className={cn(focusRing(), 'text-link')} to="/foundations/spacing-and-density">
              spacing and density
            </Link>
            , <Link className={cn(focusRing(), 'text-link')} to="/foundations/layout">layout</Link>,{' '}
            <Link className={cn(focusRing(), 'text-link')} to="/foundations/typography">typography</Link> and{' '}
            <Link className={cn(focusRing(), 'text-link')} to="/foundations/iconography">icons</Link>. Component pages
            list CSS classes and React props in their API view.
          </p>
          <p>
            Pass <code>className</code> to adjust a component: Paper merges it
            with tailwind-variants, so a later utility replaces a conflicting
            one instead of competing with it. Components with several parts take{' '}
            <code>classNames</code> keyed by part. Use <code>stack()</code> and{' '}
            <code>cluster()</code> for vertical flow and wrapping rows.
          </p>
        </section>
        <section id="forms" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Connect forms</h2>
          <p>
            Create a form with usePaperForm, wrap form.Form in{' '}
            <code>form.AppForm</code>, and bind named fields through
            <code>form.AppField</code>. Put validation on the field or form. Your application owns persistence
            and server errors.
          </p>
          <CodeBlock
            code={
              "import { Button, stack, usePaperForm } from \"paper-ui\";\n\nexport function ReadingForm() {\n  const form = usePaperForm({\n    defaultValues: { title: \"\" },\n    onSubmit: ({ value }) => console.log(value),\n  });\n  return (\n    <form.AppForm><form.Form className={stack()}>\n      <form.AppField name=\"title\" validators={{ onChange: ({ value }) => value.trim() ? undefined : \"Enter a title.\" }}>\n        {field => <field.Input label=\"Work title\" required />}\n      </form.AppField>\n      <Button type=\"submit\">Save log</Button>\n    </form.Form></form.AppForm>\n  );\n}"
            }
          />
          <p>
            The <Link className={cn(focusRing(), 'text-link')} to="/components/forms/input">Input examples</Link>{' '}
            demonstrate validation, disabled and read-only states.
          </p>
        </section>
        <section id="patterns" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Compose patterns and screens</h2>
          <p>
            Paper controls provide reusable appearance and behavior. A product
            pattern describes a recognizable Tadoku task section; a screen owns
            its route, data, permissions, and actions. Start a section beside its
            screen and extract it when reuse or independent states make that
            boundary useful. A documented pattern does not need to be an
            exported component.
          </p>
          <p>
            See the <Link className={cn(focusRing(), 'text-link')} to="/patterns/logging">logging pattern</Link>{' '}
            for a reading-log summary made from Paper controls. The consuming
            screen supplies its data and handles navigation and saving.
          </p>
        </section>
      </article>
      <TableOfContents
        items={[
          { id: 'install', label: 'Install' },
          { id: 'tailwind', label: 'Tailwind and theme' },
          { id: 'stylesheet', label: 'Document setup' },
          { id: 'compose', label: 'Classes and components' },
          { id: 'forms', label: 'Forms' },
          { id: 'patterns', label: 'Patterns and screens' },
        ]}
      />
    </div>
  )
}
