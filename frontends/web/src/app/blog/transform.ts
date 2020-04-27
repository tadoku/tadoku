import { PostOrPage } from './interfaces'
import { PostOrPage as RawPostOrPage } from '@tryghost/content-api'
import { Mapper, Mappers } from '../interfaces'
import { Serializer } from '../cache'
import { createSerializer } from '../transform'

const rawToPostOrPageMapper: Mapper<RawPostOrPage, PostOrPage> = raw => ({
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

export const postOrPageMapper: Mappers<RawPostOrPage, PostOrPage> = {
  fromRaw: rawToPostOrPageMapper,
  toRaw: postOrPageToRawMapper,
}

export const postOrPageSerializer = createSerializer(postOrPageMapper)

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
