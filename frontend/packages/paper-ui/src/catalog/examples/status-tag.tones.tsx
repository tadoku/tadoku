import { StatusTag } from 'paper-ui'

export default function Example() {
  return (
    <div className="paper-cluster">
      <StatusTag>Unlisted</StatusTag>
      <StatusTag tone="information">Scheduled</StatusTag>
      <StatusTag tone="success">Live now</StatusTag>
      <StatusTag tone="warning">Draft</StatusTag>
      <StatusTag tone="danger">Unavailable</StatusTag>
    </div>
  )
}
