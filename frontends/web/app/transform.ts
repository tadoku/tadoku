import { Serializer } from './cache'

export const OptionalizeSerializer = <DataType>(
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
