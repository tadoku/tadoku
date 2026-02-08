import {
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  FireIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { Flash } from 'ui'

export default function FlashIcons() {
  return (
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
  )
}
