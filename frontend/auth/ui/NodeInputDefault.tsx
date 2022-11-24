import { NodeInputProps } from './helpers'

export function NodeInputDefault<T>(props: NodeInputProps) {
  const { node, attributes, disabled, register } = props

  // Some attributes have dynamic JavaScript - this is for example required for WebAuthn.
  const onClick = () => {
    // This section is only used for WebAuthn. The script is loaded via a <script> node
    // and the functions are available on the global window level. Unfortunately, there
    // is currently no better way than executing eval / function here at this moment.
    if (attributes.onclick) {
      const run = new Function(attributes.onclick)
      run()
    }
  }

  // Render a generic text input field.
  return (
    <>
      <label htmlFor={attributes.name}>{node.meta.label?.text}</label>
      <input
        {...register(attributes.name, {
          required: attributes.required,
        })}
        id={attributes.name}
        onClick={onClick}
        type={attributes.type}
        disabled={attributes.disabled || disabled}
      />
      <p>
        {node.messages.map(({ text, id }, k) => (
          <span key={`${id}-${k}`}>{text}</span>
        ))}
      </p>
    </>
  )
}
