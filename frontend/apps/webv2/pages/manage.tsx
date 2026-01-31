import { GetServerSideProps } from 'next'
import { routes } from '@app/common/routes'

export const getServerSideProps: GetServerSideProps = async () => {
  return {
    redirect: {
      destination: routes.managePosts(),
      permanent: false,
    },
  }
}

const Page = () => null

export default Page
