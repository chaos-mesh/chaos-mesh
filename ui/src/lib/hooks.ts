import { useEffect, useRef } from 'react'

import { useLocation } from 'react-router-dom'

export function usePrevious<T>(value: T) {
  const ref = useRef<T>()

  useEffect(() => {
    ref.current = value
  }, [value])

  return ref.current
}

export function useQuery() {
  return new URLSearchParams(useLocation().search)
}

export function useIntervalFetch(fetch: (intervalID: number) => void, timeout: number = 6000) {
  const id = useRef(0)

  useEffect(() => {
    id.current = window.setInterval(() => fetch(id.current), timeout)

    fetch(id.current)

    return () => clearInterval(id.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
}
