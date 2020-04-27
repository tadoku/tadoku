import { Serializer } from './cache'
import { Mappers, MappersWithOptional } from './interfaces'

export const optionalizeSerializer = <DataType>(
  serializer: Serializer<DataType>,
): Serializer<DataType | undefined> => ({
  serialize: data => {
    if (!data) {
      return ''
    }

    return serializer.serialize(data)
  },
  deserialize: serializedData => {
    if (serializedData === '') {
      return undefined
    }

    return serializer.deserialize(serializedData)
  },
})

export const withOptional = <Raw, Original>(
  mappers: Mappers<Raw, Original>,
): MappersWithOptional<Raw, Original> => ({
  ...mappers,
  optional: {
    toRaw: (original: Original | undefined): Raw | undefined =>
      original ? mappers.toRaw(original) : undefined,
    fromRaw: (raw: Raw | undefined): Original | undefined =>
      raw ? mappers.fromRaw(raw) : undefined,
  },
})

export function createSerializer<Raw, Target>(
  mapper: Mappers<Raw, Target>,
): Serializer<Target> {
  return {
    serialize: data => {
      const raw = mapper.toRaw(data)
      return JSON.stringify(raw)
    },
    deserialize: data => {
      let raw = JSON.parse(data)
      return mapper.fromRaw(raw)
    },
  }
}
