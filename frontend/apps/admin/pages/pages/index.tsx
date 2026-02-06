import { routes } from '@app/common/routes'
import { DEFAULT_NAMESPACE } from '@app/content/NamespaceSelector'

export default function PagesRedirect() {
  return null
}

export function getServerSideProps() {
  return {
    redirect: {
      destination: routes.pages(DEFAULT_NAMESPACE),
      permanent: false,
    },
  }
}
