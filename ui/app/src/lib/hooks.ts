/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
