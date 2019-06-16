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

export const isReady = (status: ApiFetchStatus[]) =>
  !(
    status.includes(ApiFetchStatus.Loading) ||
    status.includes(ApiFetchStatus.Initialized)
  )

interface useCachedApiStateParameters<DataType> {
  cacheKey: string
  defaultValue: DataType
  fetchData: () => Promise<DataType>
  onChange?: (data: DataType) => void
  dependencies?: any[]
  serializer?: Serializer<DataType>
}

export const useCachedApiState = <DataType>({
  cacheKey,
  defaultValue,
  fetchData,
  onChange,
  dependencies: originalDependencies,
  serializer: originalSerializer,
}: useCachedApiStateParameters<DataType>) => {
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
