import { PageHeader, StatusTag, buttonClassName } from 'paper-ui'

export default function Example() {
  return (
    <PageHeader
      backLink={{ href: '#official-contests', label: 'Back to official contests' }}
      title="2026 Round 5"
      status={
        <>
          <StatusTag tone="success">Live now</StatusTag>
          <StatusTag>Registration open</StatusTag>
        </>
      }
      lead="Two weeks of reading and listening with the Tadoku community. Join with up to three languages."
      actions={
        <>
          <a className={buttonClassName()} href="#join">
            Join contest
          </a>
          <a className={buttonClassName({ variant: 'outline' })} href="#standings">
            View standings
          </a>
        </>
      }
    />
  )
}
