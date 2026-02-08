import { Tabbar } from 'ui'

export default function TabbarExample() {
  return (
    <Tabbar
      links={[
        {
          href: '/contests/official',
          label: 'Official contests',
          active: false,
        },
        {
          href: '/contests/user-contests',
          label: 'User contests',
          active: false,
        },
        {
          href: '/contests/my-contests',
          label: 'My contests',
          active: true,
        },
      ]}
    />
  )
}
