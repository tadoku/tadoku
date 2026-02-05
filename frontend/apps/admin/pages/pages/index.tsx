import { routes } from '@app/common/routes'

export default function PagesRedirect() {
  return null
}

export function getServerSideProps() {
  return {
    redirect: {
      destination: routes.pages('tadoku'),
      permanent: false,
    },
  }
}
