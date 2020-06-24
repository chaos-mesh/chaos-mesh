import { useEffect, useRef } from 'react'

export function usePrevious<T>(value: T) {
  const ref = useRef<T>()

  useEffect(() => {
    ref.current = value
  })

  return ref.current
}
