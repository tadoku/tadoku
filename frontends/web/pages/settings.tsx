import { ExpressNextContext } from '../app/interfaces'
import { SettingsTab } from '../app/user/interfaces'
import Settings from '../app/user/pages/Settings'

interface Props {
  tab: SettingsTab
}

const SettingsPage = ({ tab }: Props) => {
  return <Settings tab={tab} />
}

SettingsPage.getInitialProps = async ({ req, query }: ExpressNextContext) => {
  if (req && req.params) {
    const { tab } = req.params
    return { tab }
  }

  if (query.tab) {
    const { tab } = query
    return { tab }
  }

  return { tab: SettingsTab.ChangePassword }
}

export default SettingsPage
