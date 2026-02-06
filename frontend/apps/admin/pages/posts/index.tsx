import { routes } from '@app/common/routes'
import { DEFAULT_NAMESPACE } from '@app/content/NamespaceSelector'

export default function PostsRedirect() {
  return null
}

export function getServerSideProps() {
  return {
    redirect: {
      destination: routes.posts(DEFAULT_NAMESPACE),
      permanent: false,
    },
  }
}
