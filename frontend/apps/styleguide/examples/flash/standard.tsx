import { Flash } from 'ui'

export default function FlashStandard() {
  return (
    <div className="v-stack spaced">
      <Flash style="success">This is a success!</Flash>
      <Flash style="info">This is some info!</Flash>
      <Flash style="warning">This is a warning!</Flash>
      <Flash style="error">This is an error!</Flash>
    </div>
  )
}
