import { Spoiler } from 'paper-ui'
import type { ImmersionLog } from 'api-client'
import { isNsfw } from '../lib/format'

/** A log's description; one tagged NSFW stays blurred until the reader reveals it, as in Webv2. */
export function LogDescription({ log, className }: { log: Pick<ImmersionLog, 'tags'> & { description: string }; className?: string }) {
  return isNsfw(log)
    ? <Spoiler label="NSFW description" badge="NSFW" className={className}>{log.description}</Spoiler>
    : <span className={className}>{log.description}</span>
}
