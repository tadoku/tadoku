import { DateTime } from 'luxon'
import { useRouter } from 'next/router'
import { useEffect, useRef, useState } from 'react'

export const useCurrentLocation = () => {
  const router = useRouter()
  const [location, setLocation] = useState('')

  useEffect(() => {
    const origin =
      typeof window !== 'undefined' && window.location.origin
        ? window.location.origin
        : ''
    setLocation(origin + router.asPath)
  }, [router.isReady])

  return location
}

// See https://overreacted.io/making-setinterval-declarative-with-react-hooks/
export const useInterval = (callback: () => void, delay: number | 'pause') => {
  const savedCallback = useRef<() => void>()

  // Keep reference to the callback
  useEffect(() => {
    savedCallback.current = callback
  }, [callback])

  // Set up interval
  useEffect(() => {
    const tick = () => savedCallback.current?.()

    if (delay !== 'pause') {
      const id = setInterval(tick, delay)
      return () => clearInterval(id)
    }
  }, [delay])
}

export const useCurrentDateTime = () => {
  const [now, setNow] = useState(() => DateTime.utc())
  useInterval(() => setNow(DateTime.utc()), 1000)

  return now
}
