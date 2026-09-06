import { Button, ToastProvider, useToast } from 'paper-ui'

function ToastFixture() {
  const toast = useToast()
  return (
    <div className="paper-stack">
      <p>
        Notifications remain visible in this preview until dismissed.
        Applications normally use a finite timeout.
      </p>
      <div className="paper-cluster">
        <Button
          onClick={() =>
            toast.add({
              title: 'Entry saved',
              description: '48 pages added to August Japanese.',
            })
          }
        >
          Show notification
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.add({
              title: 'Could not sync entry',
              description:
                'Your local entry is safe. Retry from the reading log.',
              priority: 'high',
            })
          }
        >
          Show failure
        </Button>
        <Button variant="ghost" onClick={() => toast.close()}>
          Dismiss all
        </Button>
      </div>
    </div>
  )
}

export default function Example() {
  return (
    <ToastProvider timeout={0}>
      <ToastFixture />
    </ToastProvider>
  )
}
