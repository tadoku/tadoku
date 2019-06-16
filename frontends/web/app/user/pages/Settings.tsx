import React from 'react'
import Layout from '../../ui/components/Layout'
import { SettingsTab } from '../interfaces'

interface Props {
  tab: SettingsTab
}

const Settings = ({ tab }: Props) => {
  return <Layout title="Settings">Active tab: {tab}</Layout>
}

export default Settings
