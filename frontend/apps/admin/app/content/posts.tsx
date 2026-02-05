import { DocumentTextIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ContentConfig } from './types'
import { MarkdownPreview } from './MarkdownPreview'
import { CodeEditor } from './CodeEditor'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'

const mdExtensions = [markdown({ codeLanguages: languages })]

export const postsConfig: ContentConfig = {
  type: 'posts',
  label: 'Post',
  labelPlural: 'Posts',
  bodyField: 'content',
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
