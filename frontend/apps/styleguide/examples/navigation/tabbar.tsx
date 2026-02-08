import { Tabbar } from 'ui'

export default function TabbarExample() {
  return (
    <Tabbar
      links={[
        {
          href: '#official',
          label: 'Official contests',
          active: false,
        },
        {
          href: '#user-contests',
          label: 'User contests',
          active: false,
        },
        {
          href: '#my-contests',
          label: 'My contests',
          active: true,
        },
      ]}
    />
  )
}
