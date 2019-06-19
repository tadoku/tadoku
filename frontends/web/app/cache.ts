import { useEffect, useState } from 'react'

export interface Serializer<DataType> {
  serialize: (data: DataType) => string
  deserialize: (rawData: string) => DataType
}

export const DefaultSerializer: Serializer<any> = {
  serialize: JSON.stringify,
  deserialize: JSON.parse,
}

export enum ApiFetchStatus {
  Initialized,
  Stale,
  Loading,
  Completed,
}

interface CachedData<DataType> {
  maxAge: number
  fetchedAt: Date
  data: DataType
}

const generateCachedDataSerializer = <DataType>(
  dataSerializer: Serializer<DataType>,
): Serializer<CachedData<DataType>> => {
  return {
    serialize: cache => {
      const raw = {
        maxAge: cache.maxAge,
        fetchedAt: cache.fetchedAt.toISOString(),
        data: dataSerializer.serialize(cache.data),
      }
      return JSON.stringify(raw)
    },
    deserialize: serializedData => {
      let raw = JSON.parse(serializedData)
      return {
        maxAge: raw.maxAge,
        fetchedAt: new Date(raw.fetchedAt),
        data: dataSerializer.deserialize(raw.data),
      }
    },
  }
}

export const isReady = (status: ApiFetchStatus[]) =>
  !(
    status.includes(ApiFetchStatus.Loading) ||
    status.includes(ApiFetchStatus.Initialized)
  )

interface UseCachedApiStateParameters<DataType> {
  cacheKey: string
  defaultValue: DataType
  fetchData: () => Promise<DataType>
  onChange?: (data: DataType) => void
  dependencies?: any[]
  serializer?: Serializer<DataType>
}

const isCacheValid = <DataType>(cache: CachedData<DataType>): boolean => {
  const expirationDate = new Date(new Date(cache.fetchedAt))
  expirationDate.setSeconds(expirationDate.getSeconds() + cache.maxAge)

  const today = new Date()

  return today < expirationDate
}

export const useCachedApiState = <DataType>({
  cacheKey,
  defaultValue,
  fetchData,
  onChange,
  dependencies: originalDependencies,
  serializer: originalSerializer,
}: UseCachedApiStateParameters<DataType>) => {
  const [data, setData] = useState({
    body: defaultValue,
    status: ApiFetchStatus.Initialized,
  })
  const [apiEffectCounter, setApiEffectCounter] = useState(0)

  const dependencies = [...(originalDependencies || []), apiEffectCounter]
  const serializer = originalSerializer ? originalSerializer : DefaultSerializer

  const observedSetData = (newData: DataType, status: ApiFetchStatus) => {
    setData({ body: newData, status })

    if (onChange) {
      onChange(newData)
    }
  }

  useEffect(() => {
    let isSubscribed = true

    const cachedValue = localStorage.getItem(cacheKey)
    if (cachedValue) {
      const parsedCacheData = serializer.deserialize(cachedValue) as DataType

      if (parsedCacheData !== data.body) {
        observedSetData(parsedCacheData, ApiFetchStatus.Stale)
      }
    } else {
      setData({ ...data, status: ApiFetchStatus.Loading })
    }

    fetchData().then(fetchedData => {
      if (!isSubscribed || fetchedData === data.body) {
        return
      }

      observedSetData(fetchedData, ApiFetchStatus.Completed)
      localStorage.setItem(cacheKey, serializer.serialize(fetchedData))
    })

    return () => {
      isSubscribed = false
    }
  }, dependencies)

  return {
    data: data.body,
    status: data.status,
    reload: () => setApiEffectCounter(apiEffectCounter + 1),
  }
}
