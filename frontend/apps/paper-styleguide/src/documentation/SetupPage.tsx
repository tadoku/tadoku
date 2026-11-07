import * as stylex from '@stylexjs/stylex'
import { a11y, accentRail, linkStyles, typography } from 'paper-ui'
import { Link } from 'react-router-dom'
import { CodeBlock } from './CodeBlock'
import { TableOfContents } from './TableOfContents'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

export function SetupPage() {
  return (
    <div {...stylex.props(shellStyles.documentLayout)}>
      <article>
        <header {...stylex.props(shellStyles.documentHero, accentRail.base)}>
          <p {...stylex.props(shellStyles.eyebrow)}>Getting started</p>
          <h1 {...stylex.props(typography.page, descendantStyles.documentHeroU_uH1)}>Setup</h1>
          <p {...stylex.props(shellStyles.documentSummary)}>
            Install once, compile Paper with StyleX, then compose pages from
            Paper components and design tokens.
          </p>
        </header>
        <section id="install" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Add the workspace package</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Paper is a private package in this monorepo. Add it to your
            application’s dependencies using pnpm. Its peer dependencies are
            React and React DOM 18.2 and TanStack Form 1.33.2.
          </p>
          <CodeBlock
            language="bash"
            code={'pnpm --filter your-app add "paper-ui@workspace:*"'}
          />
        </section>
        <section id="stylex" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Compile with StyleX</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Paper ships its styles as StyleX source, so your build compiles
            Paper and your application into one atomic stylesheet. Add StyleX
            and its Vite plugin, and use the same style resolution as Paper:
            when two styles set overlapping properties, the later one wins.
          </p>
          <CodeBlock language="bash" code={'pnpm --filter your-app add @stylexjs/stylex\npnpm --filter your-app add -D @stylexjs/unplugin'} />
          <CodeBlock language="javascript" code={'// vite.config.ts\nimport stylex from "@stylexjs/unplugin";\n\nexport default defineConfig({\n  plugins: [stylex.vite({ styleResolution: "application-order", useCSSLayers: true }), react()],\n});'} />
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Add <code>paperStyleConfig</code> from <code>paper-ui/eslint</code> 
            to your ESLint config. It validates StyleX objects, sorts their
            keys, rejects raw class names and inline styles, and requires
            every length, colour and duration to come from a token.
          </p>
        </section>
        <section id="root" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Load fonts and root styles</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Import <code>paper-ui/fonts.css</code> once at your entry. It holds
            only Paper’s font faces, which StyleX cannot declare, and the
            plugin appends every compiled style to the same stylesheet. Spread 
            <code>rootProps</code> on <code>&lt;html&gt;</code> and 
            <code>&lt;body&gt;</code> for page colour and typography.
          </p>
          <CodeBlock
            code={'import "paper-ui/fonts.css";\nimport { applyDensity, applyTheme, rootProps } from "paper-ui";\n\ndocument.documentElement.className = rootProps.html().className ?? "";\ndocument.body.className = rootProps.body().className ?? "";\napplyTheme("light");\napplyDensity("comfortable");'}
          />
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            <code>applyTheme</code> and <code>applyDensity</code> swap the
            generated theme classes on <code>&lt;html&gt;</code>, so portalled
            menus and dialogs follow them. Your application owns the preference
            and its persistence. A server-rendered app puts 
            <code>themeBootScript(storageKey)</code> in an inline 
            <code>&lt;head&gt;</code> script so the stored or system theme
            applies before first paint.
          </p>
        </section>
        <section id="compose" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Style with tokens</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Write application styles with <code>stylex.create</code> and take
            every value from <code>paper-ui/tokens.stylex</code>. Colour tokens
            switch with the theme and density tokens with the density, so a
            surface needs no dark variant. Paper components accept an 
            <code>xstyle</code> prop for layout adjustments; Paper’s recipes,
            such as <code>typography</code>, <code>surfaces</code> and 
            <code>layout</code>, cover common roles.
          </p>
          <CodeBlock
            code={'import * as stylex from "@stylexjs/stylex";\nimport { Button, buttonStyles } from "paper-ui";\nimport { color, space } from "paper-ui/tokens.stylex";\n\nconst styles = stylex.create({\n  actions: {\n    alignItems: "center",\n    backgroundColor: color.surfacePaper,\n    display: "flex",\n    flexWrap: "wrap",\n    gap: space[8],\n    padding: space[16],\n  },\n});\n\nexport function ReadingActions() {\n  return (\n    <div {...stylex.props(styles.actions)}>\n      <Button onClick={() => window.print()}>Print log</Button>\n      <a {...stylex.props(...buttonStyles({ variant: "outline" }))} href="/logs">View logs</a>\n    </div>\n  );\n}'}
          />
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Start with the token references for 
            <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/spacing-and-density">
              spacing and density
            </Link>
            , <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/layout">layout</Link>, 
            <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/typography">typography</Link> and 
            <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/iconography">icons</Link>. Component pages
            list StyleX exports and React props in their API view.
          </p>
        </section>
        <section id="forms" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Connect forms</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Create a form with usePaperForm, wrap form.Form in{' '}
            <code>form.AppForm</code>, and bind named fields through
            <code>form.AppField</code>. Put validation on the field or form. Your application owns persistence
            and server errors.
          </p>
          <CodeBlock
            code={
              "import { Button, usePaperForm } from \"paper-ui\";\n\nexport function ReadingForm() {\n  const form = usePaperForm({\n    defaultValues: { title: \"\" },\n    onSubmit: ({ value }) => console.log(value),\n  });\n  return (\n    <form.AppForm><form.Form>\n      <form.AppField name=\"title\" validators={{ onChange: ({ value }) => value.trim() ? undefined : \"Enter a title.\" }}>\n        {field => <field.Input label=\"Work title\" required />}\n      </form.AppField>\n      <Button type=\"submit\">Save log</Button>\n    </form.Form></form.AppForm>\n  );\n}"
            }
          />
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            The <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/components/forms/input">Input examples</Link>{' '}
            demonstrate validation, disabled and read-only states.
          </p>
        </section>
        <section id="patterns" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Compose patterns and screens</h2>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            Paper controls provide reusable appearance and behavior. A product
            pattern describes a recognizable Tadoku task section; a screen owns
            its route, data, permissions, and actions. Start a section beside its
            screen and extract it when reuse or independent states make that
            boundary useful. A documented pattern does not need to be an
            exported component.
          </p>
          <p {...stylex.props(descendantStyles.documentSectionU_uP)}>
            See the <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/patterns/logging">logging pattern</Link>{' '}
            for a reading-log summary made from Paper controls. The consuming
            screen supplies its data and handles navigation and saving.
          </p>
        </section>
      </article>
      <TableOfContents
        items={[
          { id: 'install', label: 'Install' },
          { id: 'stylex', label: 'StyleX' },
          { id: 'root', label: 'Fonts and root styles' },
          { id: 'compose', label: 'Tokens' },
          { id: 'forms', label: 'Forms' },
          { id: 'patterns', label: 'Patterns and screens' },
        ]}
      />
    </div>
  )
}
