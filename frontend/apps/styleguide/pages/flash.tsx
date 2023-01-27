import { CodeBlock, Preview, Separator, Title } from '@components/example'
import {
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  FireIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { Flash } from 'ui'

export default function FlashPreview() {
  return (
    <>
      <h1 className="title mb-8">Flash messages</h1>
      <Title>Standard</Title>
      <Preview>
        <div className="v-stack spaced">
          <Flash style="success">This is a success!</Flash>
          <Flash style="info">This is some info!</Flash>
          <Flash style="warning">This is a warning!</Flash>
          <Flash style="error">This is an error!</Flash>
        </div>
      </Preview>
      <CodeBlock
        language="typescript"
        code={`<div className="v-stack spaced">
  <Flash style="success">This is a success!</Flash>
  <Flash style="info">This is some info!</Flash>
  <Flash style="warning">This is a warning!</Flash>
  <Flash style="error">This is an error!</Flash>
</div>`}
      />

      <Separator />

      <Title>Link variation</Title>
      <Preview>
        <div className="v-stack spaced">
          <Flash style="success" href="#">
            This is a success!
          </Flash>
          <Flash style="info" href="#">
            This is some info!
          </Flash>
          <Flash style="warning" href="#">
            This is a warning!
          </Flash>
          <Flash style="error" href="#">
            This is an error!
          </Flash>
        </div>
      </Preview>
      <CodeBlock
        language="typescript"
        code={`<div className="v-stack spaced">
  <Flash style="success" href="#">
    This is a success!
  </Flash>
  <Flash style="info" href="#">
    This is some info!
  </Flash>
  <Flash style="warning" href="#">
    This is a warning!
  </Flash>
  <Flash style="error" href="#">
    This is an error!
  </Flash>
</div>`}
      />

      <Separator />

      <Title>Icon variation</Title>
      <Preview>
        <div className="v-stack spaced">
          <Flash style="success" IconComponent={FireIcon}>
            This is a success!
          </Flash>
          <Flash style="info" IconComponent={InformationCircleIcon}>
            This is some info!
          </Flash>
          <Flash style="warning" IconComponent={ExclamationCircleIcon}>
            This is a warning!
          </Flash>
          <Flash style="error" IconComponent={ExclamationTriangleIcon}>
            This is an error!
          </Flash>
        </div>
      </Preview>
      <CodeBlock
        language="typescript"
        code={`<Flash style="success" IconComponent={FireIcon}>
  This is a success!
</Flash>
<Flash style="info" IconComponent={InformationCircleIcon}>
  This is some info!
</Flash>
<Flash style="warning" IconComponent={ExclamationCircleIcon}>
  This is a warning!
</Flash>
<Flash style="error" IconComponent={ExclamationTriangleIcon}>
  This is an error!
</Flash>
</div>`}
      />
    </>
  )
}
