import {
  ArrowDownTrayIcon,
  CodeBracketIcon,
  TableCellsIcon,
} from '@heroicons/react/20/solid'
import { ActionMenu } from 'ui'
import { useSession } from '@app/common/session'
import { logExportUrl } from '@app/immersion/api'

export const LogExportMenu = ({ userId }: { userId: string }) => {
  const [session] = useSession()

  if (session?.id !== userId) {
    return null
  }

  return (
    <ActionMenu
      orientation="right"
      links={[
        {
          label: 'Download CSV',
          href: logExportUrl('csv'),
          IconComponent: TableCellsIcon,
        },
        {
          label: 'Download JSON',
          href: logExportUrl('json'),
          IconComponent: CodeBracketIcon,
        },
      ]}
    >
      <ArrowDownTrayIcon className="w-5 h-5" />
      Export
    </ActionMenu>
  )
}
