import * as stylex from '@stylexjs/stylex'
import { StatusTag, layout } from 'paper-ui'

export default function Example() {
  return (
    <div {...stylex.props(layout.cluster)}>
      <StatusTag>Unlisted</StatusTag>
      <StatusTag tone="information">Scheduled</StatusTag>
      <StatusTag tone="success">Live now</StatusTag>
      <StatusTag tone="warning">Draft</StatusTag>
      <StatusTag tone="danger">Unavailable</StatusTag>
    </div>
  )
}
