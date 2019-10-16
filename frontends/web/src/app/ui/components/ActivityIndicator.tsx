import React from 'react'
import { Button } from './index'

interface Props {
  isLoading: boolean
}

const ActivityIndicator = ({ isLoading }: Props) =>
  isLoading ? <Button plain loading disabled /> : null

export default ActivityIndicator
