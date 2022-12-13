import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'

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
