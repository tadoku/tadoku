import { PostOrPage } from './interfaces'
import { PostOrPage as RawPostOrPage } from '@tryghost/content-api'
import { Mapper, Mappers } from '../interfaces'
import { Serializer } from '../cache'
import {
  createSerializer,
  createMappers,
  createCollectionSerializer,
} from '../transform'

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

export const postOrPageMapper: Mappers<
  RawPostOrPage,
  PostOrPage
> = createMappers({
  fromRaw: rawToPostOrPageMapper,
  toRaw: postOrPageToRawMapper,
})

export const postOrPageSerializer: Serializer<PostOrPage> = createSerializer(
  postOrPageMapper,
)

export const postOrPagesSerializer: Serializer<
  PostOrPage[]
> = createCollectionSerializer(postOrPageMapper)
