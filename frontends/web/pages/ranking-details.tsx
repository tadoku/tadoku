import { ExpressNextContext } from '../app/interfaces'
import RankingProfile from '../app/ranking/components/RankingProfile'
interface Props {
  contestId: number | undefined
  userId: number | undefined
}

const RankingDetails = (props: Props) => <RankingProfile {...props} />

RankingDetails.getInitialProps = async ({ req, query }: ExpressNextContext) => {
  if (req && req.params) {
    const { contest_id, user_id } = req.params

    return {
      contestId: parseInt(contest_id),
      userId: parseInt(user_id),
    }
  }

  if (query.contest_id && query.user_id) {
    const { contest_id, user_id } = query

    return {
      contestId: parseInt(contest_id as string),
      userId: parseInt(user_id as string),
    }
  }

  return {}
}

export default RankingDetails
