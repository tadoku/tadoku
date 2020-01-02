import { RawPostOrPage, PostOrPage } from './interfaces'
import { Mapper } from '../interfaces'

export const RawToPostOrPageMapper: Mapper<
  RawPostOrPage,
  PostOrPage
> = raw => ({
  slug: raw.slug,
  title: raw.title ?? '',
  html: raw.html ?? '',
})
