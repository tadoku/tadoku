import { UiNode, UiNodeInputAttributes } from '@ory/client'
import { FormEvent } from 'react'
import { UseFormRegister, UseFormSetValue } from 'react-hook-form'

export type ValueSetter = (
  value: string | number | boolean | undefined,
) => Promise<void>

export type FormDispatcher = (e: MouseEvent | FormEvent) => Promise<void>

export interface NodeInputProps {
  node: UiNode
  attributes: UiNodeInputAttributes
  disabled: boolean
  register: UseFormRegister<any>
}
