import { DocumentTextIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ContentConfig } from './types'
import { MarkdownPreview } from './MarkdownPreview'

export const postsConfig: ContentConfig = {
  type: 'posts',
  label: 'Post',
  labelPlural: 'Posts',
  bodyField: 'content',
  renderBody: (body: string) => <MarkdownPreview content={body} />,
  routes: {
    list: routes.posts,
    preview: routes.postPreview,
    edit: routes.postEdit,
    new: routes.postNew,
  },
  icon: DocumentTextIcon,
  sidebarKey: 'posts',
}
