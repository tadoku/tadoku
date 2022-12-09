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
      <label
        htmlFor={attributes.name}
        className={`label ${node.messages.length > 0 ? 'error' : ''}`}
      >
        <span className="label-text">{node.meta.label?.text}</span>
        <input
          {...register(attributes.name, {
            required: attributes.required,
          })}
          id={attributes.name}
          onClick={onClick}
          type={attributes.type}
          disabled={attributes.disabled || disabled}
        />

        {node.messages.map(({ text, id }, k) => (
          <span className="error" key={`${id}-${k}`}>
            {text}
          </span>
        ))}
      </label>
      <p></p>
    </>
  )
}
