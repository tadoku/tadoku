import React from 'react'

interface Props {
  isLoading: boolean
}

const ActivityIndicator = ({ isLoading }: Props) =>
  isLoading ? <span>Loading...</span> : null

export default ActivityIndicator
