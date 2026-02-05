import { DocumentDuplicateIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ContentConfig } from './types'
import { CodeEditor } from './CodeEditor'
import { html } from '@codemirror/lang-html'

const htmlExtensions = [html()]

export const pagesConfig: ContentConfig = {
  type: 'pages',
  label: 'Page',
  labelPlural: 'Pages',
  bodyField: 'html',
  renderBody: (body: string) => (
    <div className="auto-format" dangerouslySetInnerHTML={{ __html: body }} />
  ),
  renderEditor: (props) => (
    <CodeEditor {...props} extensions={htmlExtensions} />
  ),
  routes: {
    list: routes.pages,
    preview: routes.pagePreview,
    edit: routes.pageEdit,
    new: routes.pageNew,
  },
  icon: DocumentDuplicateIcon,
  sidebarKey: 'pages',
}
