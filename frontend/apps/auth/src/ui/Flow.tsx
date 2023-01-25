import {
  SelfServiceLoginFlow,
  SelfServiceRecoveryFlow,
  SelfServiceRegistrationFlow,
  SelfServiceSettingsFlow,
  SelfServiceVerificationFlow,
  UiNode,
  UiNodeInputAttributes,
} from '@ory/client'
import { isUiNodeInputAttributes, getNodeId } from '@ory/integrations/ui'
import { FormProvider, useForm } from 'react-hook-form'
import MessagesList from './MessagesList'
import Node from './Node'

export type SelfServiceFlow =
  | SelfServiceLoginFlow
  | SelfServiceRegistrationFlow
  | SelfServiceSettingsFlow
  | SelfServiceVerificationFlow
  | SelfServiceRecoveryFlow

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
  const ignoredNodeTypes = ['button', 'submit']
  return nodes
    .filter(node => isUiNodeInputAttributes(node.attributes))
    .filter(
      node =>
        !ignoredNodeTypes.includes(
          (node.attributes as UiNodeInputAttributes).type,
        ),
    )
    .reduce((acc, node) => {
      const attr = node.attributes as UiNodeInputAttributes
      acc[attr.name] = attr.value
      return acc
    }, {} as { [key: string]: any })
}

const Flow = ({ flow, method, onSubmit, hideGlobalMessages }: FlowProps) => {
  const nodes = filterNodes(flow, method)
  const defaultValues = defaultValuesFromNodes(nodes)

  const methods = useForm({
    defaultValues,
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
          const id = getNodeId(node)
          return (
            <Node
              key={`${id}-${k}`}
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
