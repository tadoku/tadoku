import { useEffect, useState } from 'react'

export enum ApiFetchStatus {
  Initialized,
  Stale,
  Loading,
  Completed,
}

export const useCachedApiState = (
  cacheKey: string,
  defaultValue: any,
  fetchData: () => Promise<any>,
  dependencies: any[],
) => {
  const [status, setStatus] = useState(ApiFetchStatus.Initialized)
  const [data, setData] = useState(defaultValue)

  const reloadData = async () => {
    const cachedValue = localStorage.getItem(cacheKey)
    if (cachedValue) {
      setStatus(ApiFetchStatus.Stale)
      setData(JSON.parse(cachedValue))
    } else {
      // We don't want to set loading state when we don't have a cached version
      setStatus(ApiFetchStatus.Loading)
    }

    const fetchedData = await fetchData()
    if (fetchedData !== data) {
      setData(fetchedData)
      localStorage.setItem(cacheKey, JSON.stringify(fetchedData))
    }

    setStatus(ApiFetchStatus.Completed)
  }

  useEffect(() => {
    const update = async () => await reloadData()
    update()
  }, dependencies)

  return [data, status, reloadData]
}
