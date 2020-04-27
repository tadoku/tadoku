import { PostOrPage } from './interfaces'
import { PostOrPage as RawPostOrPage } from '@tryghost/content-api'
import { Mapper } from '../interfaces'
import { Serializer } from '../cache'

export const rawToPostOrPageMapper: Mapper<
  RawPostOrPage,
  PostOrPage
> = raw => ({
  id: raw.id,
  slug: raw.slug,
  title: raw.title ?? '',
  html: raw.html ?? '',
  publishedAt: raw.published_at ? new Date(raw.published_at) : new Date(),
})

const postOrPageToRawMapper: Mapper<
  PostOrPage,
  RawPostOrPage
> = postOrPage => ({
  id: postOrPage.id,
  slug: postOrPage.slug,
  title: postOrPage.title,
  html: postOrPage.html,
  published_at: postOrPage.publishedAt.toISOString(),
})

export const postOrPageSerializer: Serializer<PostOrPage> = {
  serialize: postOrPage => {
    let raw = postOrPageToRawMapper(postOrPage)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return rawToPostOrPageMapper(raw)
  },
}

export const postOrPagesSerializer: Serializer<PostOrPage[]> = {
  serialize: postOrPages => {
    const raw = postOrPages.map(postOrPageToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(rawToPostOrPageMapper)
  },
}
