import { Button, Drawer } from 'paper-ui'
import { useState } from 'react'

const languages = ['Japanese', 'Spanish', 'Korean', 'French', 'German', 'Mandarin', 'Italian', 'Portuguese']

export default function FullScreenDrawerExample() {
  const [open, setOpen] = useState(false)

  return (
    <Drawer
      size="full"
      open={open}
      onOpenChange={setOpen}
      trigger={<Button variant="outline">Choose a language</Button>}
      title="Language"
      closeLabel="Close language"
      footer={<Button onClick={() => setOpen(false)}>Done</Button>}
    >
      <ul className="m-0 list-none p-0">
        {languages.map(language => (
          <li key={language} className="border-b border-rule-subtle py-3">{language}</li>
        ))}
      </ul>
    </Drawer>
  )
}
