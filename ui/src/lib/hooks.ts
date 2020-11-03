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
