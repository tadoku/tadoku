import { Link } from 'react-router-dom'
import { CodeBlock } from './CodeBlock'
import { TableOfContents } from './TableOfContents'

export function SetupPage() {
  return (
    <div className="document-layout">
      <article className="document-page">
        <header className="document-hero paper-accent-rail">
          <p className="eyebrow">Getting started</p>
          <h1 className="paper-type-page">Setup</h1>
          <p className="document-summary">
            Install once, connect Paper’s Tailwind preset, then compose pages
            with familiar utilities and Paper components.
          </p>
        </header>
        <section id="install" className="document-section">
          <h2 className="paper-type-section">Add the workspace package</h2>
          <p>
            Paper is a private package in this monorepo. Add it to your
            application’s dependencies using pnpm. Its peer dependencies are
            React and React DOM 18.2 and React Hook Form 7.
          </p>
          <CodeBlock
            language="bash"
            code={'pnpm --filter your-app add "paper-ui@workspace:*"'}
          />
        </section>
        <section id="tailwind" className="document-section">
          <h2 className="paper-type-section">Use the Tailwind preset</h2>
          <p>
            Paper uses Tailwind 3 for page composition. Use <code>p-4</code>,{' '}
            <code>gap-2</code>, <code>bg-paper</code> and <code>text-ink</code>.
            The preset maps these utilities to Paper’s semantic values and keeps
            Tailwind’s other sizes, breakpoints and variants available.
          </p>
          <CodeBlock language="bash" code={'pnpm --filter your-app add -D tailwindcss@3 postcss autoprefixer'} />
          <CodeBlock language="javascript" code={'// tailwind.config.cjs\nmodule.exports = {\n  presets: [require("paper-ui/tailwind-preset")],\n  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],\n};\n\n// postcss.config.cjs\nmodule.exports = {\n  plugins: { tailwindcss: {}, autoprefixer: {} },\n};'} />
          <p>
            Adjust <code>content</code> to your app’s source directories. Keep
            complete class names in source, such as <code>p-4</code>; Tailwind
            cannot detect a constructed string like <code>{'`p-${size}`'}</code>.
          </p>
          <CodeBlock language="css" code={'/* src/tailwind.css */\n@tailwind base;\n@tailwind utilities;'} />
          <p>
            The preset disables Tailwind’s Preflight reset because Paper supplies
            its own base and component styles. For a custom border, specify its
            style too: <code>border border-solid border-rule</code>. Use Paper
            components for buttons and fields, which already own their appearance.
          </p>
        </section>
        <section id="stylesheet" className="document-section">
          <h2 className="paper-type-section">Load styles once</h2>
          <p>
            Import the stylesheet at your application entry. It includes fonts,
            tokens, base styles, utilities and components; it also sets page
            typography and box sizing. There is no need to import fonts
            separately.
          </p>
          <CodeBlock
            code={
              'import "paper-ui/styles.css";\nimport "./tailwind.css";\n\n// Set these on <html> so portalled menus and dialogs inherit them.\ndocument.documentElement.dataset.theme = "light";\ndocument.documentElement.dataset.density = "comfortable";'
            }
          />
          <p>
            Use <code>data-theme="dark"</code> for dark appearance and{' '}
            <code>data-density="compact"</code> for compact controls. Your
            application owns the preference and its persistence. The stylesheet
            does not automatically follow system appearance.
          </p>
          <p>
            Semantic colors switch with <code>data-theme</code>; a{' '}
            <code>bg-paper</code> surface needs no separate <code>dark:</code>{' '}
            utility. Fixed spacing such as <code>p-4</code> stays the same in
            both densities. Components and <code>gap-inline</code> follow density.
          </p>
        </section>
        <section id="compose" className="document-section">
          <h2 className="paper-type-section">
            Choose classes or React components
          </h2>
          <p>
            Use Tailwind for layout and spacing. Paper’s documented component
            classes also work directly in HTML or JSX. React components
            provide the same appearance together with behavior, accessibility
            and typed props. For example, a navigation link is an anchor with
            button classes; an action uses <code>Button</code>.
          </p>
          <CodeBlock
            code={
              'import { Button } from "paper-ui";\n\nexport function ReadingActions() {\n  return (\n    <div className="flex flex-wrap items-center gap-2 bg-paper p-4">\n      <Button onClick={() => window.print()}>Print log</Button>\n      <a className="paper-button paper-button--outline" href="/logs">View logs</a>\n    </div>\n  );\n}'
            }
          />
          <p>
            Start with the class references for{' '}
            <Link className="text-link paper-focus-ring" to="/foundations/spacing-and-density">
              spacing and density
            </Link>
            , <Link className="text-link paper-focus-ring" to="/foundations/layout">layout</Link>,{' '}
            <Link className="text-link paper-focus-ring" to="/foundations/typography">typography</Link> and{' '}
            <Link className="text-link paper-focus-ring" to="/foundations/iconography">icons</Link>. Component pages
            list CSS classes and React props in their API view.
          </p>
          <p>
            Without Tailwind, importing <code>paper-ui/styles.css</code> still
            supplies all Paper components. Use <code>paper-stack</code>,{' '}
            <code>paper-cluster</code> or your own CSS with Paper variables for
            composition. Tailwind is a build-time convenience, not a component
            runtime dependency.
          </p>
        </section>
        <section id="forms" className="document-section">
          <h2 className="paper-type-section">Connect forms</h2>
          <p>
            Paper fields read React Hook Form context. Wrap a form in{' '}
            <code>FormProvider</code>, give each field a name, and submit
            through <code>handleSubmit</code>. Your application owns persistence
            and server errors.
          </p>
          <CodeBlock
            code={
              'import { FormProvider, useForm } from "react-hook-form";\nimport { Button, Input } from "paper-ui";\n\nexport function ReadingForm() {\n  const methods = useForm({ defaultValues: { title: "" } });\n  return (\n    <FormProvider {...methods}>\n      <form className="paper-stack" onSubmit={methods.handleSubmit(values => console.log(values))}>\n        <Input name="title" label="Work title" required />\n        <Button type="submit">Save log</Button>\n      </form>\n    </FormProvider>\n  );\n}'
            }
          />
          <p>
            The <Link className="text-link paper-focus-ring" to="/components/forms/input">Input examples</Link>{' '}
            demonstrate validation, disabled and read-only states.
          </p>
        </section>
      </article>
      <TableOfContents
        items={[
          { id: 'install', label: 'Install' },
          { id: 'tailwind', label: 'Tailwind preset' },
          { id: 'stylesheet', label: 'Stylesheet' },
          { id: 'compose', label: 'Classes and components' },
          { id: 'forms', label: 'Forms' },
        ]}
      />
    </div>
  )
}
