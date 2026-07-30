import { formatScore } from '@app/common/format'
import {
  ContestRegistrationView,
  ScorePreview,
} from '@app/immersion/api'

interface Props {
  enabled: boolean
  preview?: ScorePreview
  registrations?: ContestRegistrationView[]
}

export const ScorePreviewEstimate = ({
  enabled,
  preview,
  registrations = [],
}: Props) => {
  const registrationsById = new Map(
    registrations.map(registration => [registration.id, registration]),
  )

  return (
    <>
      <div>
        Estimated score:{' '}
        <strong>
          {enabled && preview
            ? formatScore(preview.platform.score)
            : '-'}
        </strong>
      </div>
      {enabled &&
        preview?.contests.map(contest => (
          <div key={contest.registration_id} className="text-sm">
            {registrationsById.get(contest.registration_id)?.contest?.title ??
              'Contest'}
            : <strong>{formatScore(contest.estimate.score)}</strong>
          </div>
        ))}
    </>
  )
}
