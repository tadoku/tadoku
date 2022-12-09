import { UiText } from '@ory/client'

interface MessagesListProps {
  messages: UiText[] | undefined
}

const MessagesList = ({ messages }: MessagesListProps) => {
  if (!messages) {
    return null
  }

  return (
    <ul className={`mb-4 space-y-2`}>
      {messages.map(m => (
        <Message key={m.id} message={m} />
      ))}
    </ul>
  )
}

interface MessageProps {
  message: UiText
}

export const Message = ({ message }: MessageProps) => {
  return (
    <li className={`flash ${message.type === 'error' ? 'error' : 'info'}`}>
      {message.text}
    </li>
  )
}

export default MessagesList
