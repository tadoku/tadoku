import { routes } from '@app/common/routes'

export default function PostsRedirect() {
  return null
}

export function getServerSideProps() {
  return {
    redirect: {
      destination: routes.posts('tadoku'),
      permanent: false,
    },
  }
}
