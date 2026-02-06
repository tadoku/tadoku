import { routes } from '@app/common/routes'
import { DEFAULT_NAMESPACE } from '@app/content/NamespaceSelector'

export default function AnnouncementsRedirect() {
  return null
}

export function getServerSideProps() {
  return {
    redirect: {
      destination: routes.announcements(DEFAULT_NAMESPACE),
      permanent: false,
    },
  }
}
