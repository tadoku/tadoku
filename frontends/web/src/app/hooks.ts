import { useRef, useEffect } from 'react'

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
