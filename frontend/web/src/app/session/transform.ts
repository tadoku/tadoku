import { User, RawUser } from './interfaces'
import { Mapper, Mappers } from '../interfaces'
import { Serializer } from '../cache'
import {
  createSerializer,
  createMappers,
  createCollectionSerializer,
} from '../transform'

const rawToUserMapper: Mapper<RawUser, User> = raw => ({
  id: raw.id,
  email: raw.email,
  displayName: raw.display_name,
  role: raw.role,
})

const userToRawMapper: Mapper<User, RawUser> = user => ({
  id: user.id,
  email: user.email,
  display_name: user.displayName,
  role: user.role,
})

export const userMapper: Mappers<RawUser, User> = createMappers({
  fromRaw: rawToUserMapper,
  toRaw: userToRawMapper,
})

export const contestSerializer: Serializer<User> = createSerializer(userMapper)

export const userCollectionSerializer: Serializer<User[]> =
  createCollectionSerializer(userMapper)
