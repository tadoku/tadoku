import { DocumentTextIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ContentConfig } from './types'
import { MarkdownPreview } from './MarkdownPreview'
import { CodeEditor } from './CodeEditor'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'
import prettier from 'prettier/standalone'
import parserMarkdown from 'prettier/parser-markdown'

const mdExtensions = [markdown({ codeLanguages: languages })]

export const postsConfig: ContentConfig = {
  type: 'posts',
  label: 'Post',
  labelPlural: 'Posts',
  bodyField: 'content',
  formatBody: (body: string) =>
    prettier.format(body, {
      parser: 'markdown',
      plugins: [parserMarkdown],
      tabWidth: 2,
      proseWrap: 'always',
      printWidth: 80,
    }),
  renderBody: (body: string) => <MarkdownPreview content={body} />,
  renderEditor: (props) => (
    <CodeEditor {...props} extensions={mdExtensions} />
  ),
  routes: {
    list: routes.posts,
    preview: routes.postPreview,
    edit: routes.postEdit,
    new: routes.postNew,
  },
  icon: DocumentTextIcon,
  sidebarKey: 'posts',
}
