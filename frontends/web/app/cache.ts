import { useEffect, useState } from 'react'

export interface Serializer<DataType> {
  serialize: (data: DataType) => string
  deserialize: (rawData: string) => DataType
}

const jsonSerializer: Serializer<any> = {
  serialize: JSON.stringify,
  deserialize: JSON.parse,
}

export enum ApiFetchStatus {
  Initialized,
  Stale,
  Loading,
  Completed,
}

export const isReady = (status: ApiFetchStatus[]) =>
  !status.includes(ApiFetchStatus.Loading)

interface useCachedApiStateParameters<DataType> {
  cacheKey: string
  defaultValue: DataType
  fetchData: () => Promise<DataType>
  onChange?: (data: DataType) => void
  dependencies?: any[]
  serializer?: Serializer<DataType> | Serializer<Exclude<DataType, undefined>>
}

export const useCachedApiState = <DataType>({
  cacheKey,
  defaultValue,
  fetchData,
  onChange,
  dependencies: originalDependencies,
  serializer: originalSerializer,
}: useCachedApiStateParameters<DataType>) => {
  const [status, setStatus] = useState(ApiFetchStatus.Initialized)
  const [data, setData] = useState(defaultValue)
  const [apiEffectCounter, setApiEffectCounter] = useState(0)

  const dependencies = [...(originalDependencies || []), apiEffectCounter]
  const serializer = originalSerializer ? originalSerializer : jsonSerializer

  const observedSetData = (newData: DataType) => {
    setData(newData)

    if (onChange) {
      onChange(newData)
    }
  }

  useEffect(() => {
    let isSubscribed = true

    const cachedValue = localStorage.getItem(cacheKey)
    if (cachedValue) {
      const parsedCacheData = serializer.deserialize(cachedValue) as DataType

      if (parsedCacheData !== data) {
        setStatus(ApiFetchStatus.Stale)
        observedSetData(parsedCacheData)
      }
    } else {
      // We don't want to set loading state when we don't have a cached version
      setStatus(ApiFetchStatus.Loading)
    }

    fetchData().then(fetchedData => {
      if (!isSubscribed || fetchedData === data) {
        return
      }

      observedSetData(fetchedData)

      if (fetchedData !== undefined) {
        localStorage.setItem(
          cacheKey,
          serializer.serialize(fetchedData as Exclude<DataType, undefined>),
        )
      } else {
        localStorage.setItem(cacheKey, '')
      }
    })

    setStatus(ApiFetchStatus.Completed)

    return () => {
      isSubscribed = false
    }
  }, dependencies)

  return {
    data,
    status,
    reload: () => setApiEffectCounter(apiEffectCounter + 1),
  }
}
