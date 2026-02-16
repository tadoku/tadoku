import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import NewLogForm from '@examples/logging-v2/new-log-form'
import newLogFormCode from '@examples/logging-v2/new-log-form.tsx?raw'

import SubmitToContest from '@examples/logging-v2/submit-to-contest'
import submitToContestCode from '@examples/logging-v2/submit-to-contest.tsx?raw'

import LogDetails from '@examples/logging-v2/log-details'
import logDetailsCode from '@examples/logging-v2/log-details.tsx?raw'

export default function LoggingV2() {
  return (
    <>
      <h1 className="title mb-8">Logging flow v2</h1>

      <Showcase title="New log form" code={newLogFormCode}>
        <div className="w-full max-w-xl">
          <NewLogForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="Submit to contest" code={submitToContestCode}>
        <div className="w-full max-w-xl">
          <SubmitToContest />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="Log details" code={logDetailsCode}>
        <div className="w-full max-w-xl">
          <LogDetails />
        </div>
      </Showcase>

    </>
  )
}
