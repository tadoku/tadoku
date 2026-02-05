import { DocumentDuplicateIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ContentConfig } from './types'

export const pagesConfig: ContentConfig = {
  type: 'pages',
  label: 'Page',
  labelPlural: 'Pages',
  bodyField: 'html',
  renderBody: (body: string) => (
    <div className="auto-format" dangerouslySetInnerHTML={{ __html: body }} />
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
