import { Serializer } from './cache'
import { Mappers } from './interfaces'

export const createMappers = <Raw, Original>({
  toRaw,
  fromRaw,
}: {
  toRaw: (original: Original) => Raw
  fromRaw: (raw: Raw) => Original
}): Mappers<Raw, Original> => ({
  toRaw,
  fromRaw,
  optional: {
    toRaw: (original: Original | undefined): Raw | undefined =>
      original ? toRaw(original) : undefined,
    fromRaw: (raw: Raw | undefined): Original | undefined =>
      raw ? fromRaw(raw) : undefined,
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

export function createCollectionSerializer<Raw, Target>(
  mapper: Mappers<Raw, Target>,
): Serializer<Target[]> {
  return {
    serialize: data => {
      const raw = data.map(mapper.toRaw)
      return JSON.stringify(raw)
    },
    deserialize: data => {
      let raw = JSON.parse(data)
      return raw.map(mapper.fromRaw)
    },
  }
}

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
