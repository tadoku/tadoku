import { Flash } from 'ui'

export default function FlashLinks() {
  return (
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
  )
}
