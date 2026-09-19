import {
  LoginFlow,
  RecoveryFlow,
  RegistrationFlow,
  SettingsFlow,
  VerificationFlow,
  UiNode,
} from '@ory/kratos-client'
import { FormProvider, set, useForm } from 'react-hook-form'
import { useMemo } from 'react'
import MessagesList from './MessagesList'
import Node from './Node'

export type SelfServiceFlow =
  | LoginFlow
  | RegistrationFlow
  | SettingsFlow
  | VerificationFlow
  | RecoveryFlow

export type Method =
  | 'oidc'
  | 'password'
  | 'profile'
  | 'totp'
  | 'webauthn'
  | 'link'
  | 'lookup_secret'

interface FlowProps {
  flow: SelfServiceFlow | undefined
  method?: Method
  onSubmit: (data: any) => void
  hideGlobalMessages?: boolean
}

const filterNodes = (
  flow: SelfServiceFlow | undefined,
  targetGroup: Method | undefined,
): UiNode[] => {
  if (!flow) {
    return []
  }

  return flow.ui.nodes.filter(({ group }) => {
    if (targetGroup === undefined) {
      return true
    }

    return group === 'default' || group === targetGroup
  })
}

const defaultValuesFromNodes = (nodes: UiNode[]): { [key: string]: any } => {
  return nodes.reduce((values, { attributes }) => {
    if (
      attributes.node_type === 'input' &&
      attributes.type !== 'button' &&
      attributes.type !== 'submit'
    ) {
      set(values, attributes.name, attributes.value)
    }
    return values
  }, {} as { [key: string]: any })
}

const Flow = ({ flow, method, onSubmit, hideGlobalMessages }: FlowProps) => {
  const nodes = filterNodes(flow, method)
  const defaultValues = useMemo(
    () => defaultValuesFromNodes(filterNodes(flow, method)),
    [flow, method],
  )

  const methods = useForm({
    values: defaultValues,
  })

  if (!flow) {
    return null
  }

  const disabled = methods.formState.isSubmitting

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="kratos-form relative"
      >
        {!hideGlobalMessages && <MessagesList messages={flow.ui.messages} />}
        {nodes.map((node, k) => {
          return (
            <Node
              key={`${node.group}-${node.type}-${k}`}
              disabled={disabled}
              node={node}
              dispatchSubmit={methods.handleSubmit(onSubmit)}
            />
          )
        })}
      </form>
    </FormProvider>
  )
}

export default Flow
