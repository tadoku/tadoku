import { UiNode, UiNodeInputAttributes } from '@ory/client'
import { BaseSyntheticEvent } from 'react'

export type ValueSetter = (
  value: string | number | boolean | undefined,
) => Promise<void>

export type FormDispatcher = (e: BaseSyntheticEvent) => Promise<void>

export interface NodeInputProps {
  node: UiNode
  attributes: UiNodeInputAttributes
  disabled: boolean
  dispatchSubmit: FormDispatcher
}
