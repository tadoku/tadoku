import { useEffect, useState } from 'react'

export enum ApiFetchStatus {
  Initialized,
  Stale,
  Loading,
  Completed,
}

interface useCachedApiStateParameters<DataType> {
  cacheKey: string
  defaultValue: DataType
  fetchData: () => Promise<DataType>
  onChange?: (data: DataType) => void
  dependencies?: any[]
}

export const useCachedApiState = <DataType>({
  cacheKey,
  defaultValue,
  fetchData,
  onChange,
  dependencies: originalDependencies,
}: useCachedApiStateParameters<DataType>) => {
  const [status, setStatus] = useState(ApiFetchStatus.Initialized)
  const [data, setData] = useState(defaultValue)
  const [apiEffectCounter, setApiEffectCounter] = useState(0)

  const dependencies = [...(originalDependencies || []), apiEffectCounter]

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
      const parsedCacheData = JSON.parse(cachedValue)

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
      localStorage.setItem(cacheKey, JSON.stringify(fetchedData))
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
