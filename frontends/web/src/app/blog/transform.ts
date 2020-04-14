import { PostOrPage } from './interfaces'
import { PostOrPage as RawPostOrPage } from '@tryghost/content-api'
import { Mapper } from '../interfaces'

export const RawToPostOrPageMapper: Mapper<
  RawPostOrPage,
  PostOrPage
> = raw => ({
  slug: raw.slug,
  title: raw.title ?? '',
  html: raw.html ?? '',
})
