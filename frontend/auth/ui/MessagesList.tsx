import { UiText } from '@ory/client'

interface MessagesListProps {
  messages: UiText[] | undefined
}

const MessagesList = ({ messages }: MessagesListProps) => {
  if (!messages) {
    return null
  }

  return (
    <ul>
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
    <li
      style={{
        backgroundColor: message.type === 'error' ? 'lightred' : 'lightblue',
      }}
    >
      {message.text}
    </li>
  )
}

export default MessagesList
