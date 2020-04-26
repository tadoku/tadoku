import { PostOrPage } from './interfaces'
import { PostOrPage as RawPostOrPage } from '@tryghost/content-api'
import { Mapper } from '../interfaces'
import { Serializer } from '../cache'

export const RawToPostOrPageMapper: Mapper<
  RawPostOrPage,
  PostOrPage
> = raw => ({
  id: raw.id,
  slug: raw.slug,
  title: raw.title ?? '',
  html: raw.html ?? '',
  publishedAt: raw.published_at ? new Date(raw.published_at) : new Date(),
})

const PostOrPageToRawMapper: Mapper<
  PostOrPage,
  RawPostOrPage
> = postOrPage => ({
  id: postOrPage.id,
  slug: postOrPage.slug,
  title: postOrPage.title,
  html: postOrPage.html,
  publishedAt: postOrPage.publishedAt.toISOString(),
})

export const PostOrPageSerializer: Serializer<PostOrPage> = {
  serialize: postOrPage => {
    let raw = PostOrPageToRawMapper(postOrPage)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return RawToPostOrPageMapper(raw)
  },
}

export const PostOrPagesSerializer: Serializer<PostOrPage[]> = {
  serialize: postOrPages => {
    const raw = postOrPages.map(PostOrPageToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(RawToPostOrPageMapper)
  },
}
