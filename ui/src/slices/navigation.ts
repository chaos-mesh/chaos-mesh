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
import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import { toTitleCase } from 'lib/utils'

export interface NavigationBreadCrumbProps {
  name: string
  path?: string
}

function pathnameToBreadCrumbs(pathname: string) {
  const nameArray = pathname.slice(1).split('/')

  return nameArray.map((name, i) => {
    const b: NavigationBreadCrumbProps = {
      name,
    }

    if (i < nameArray.length - 1) {
      b.path = '/' + nameArray.slice(0, i + 1).join('/')
    }

    return b
  })
}

const navigationSlice = createSlice({
  name: 'navigation',
  initialState: {
    breadcrumbs: [] as NavigationBreadCrumbProps[],
  },
  reducers: {
    setNavigationBreadcrumbs(state, action: PayloadAction<string>) {
      const breadcrumbs = pathnameToBreadCrumbs(action.payload)

      if (breadcrumbs[0].name) {
        state.breadcrumbs = breadcrumbs

        document.title = toTitleCase(breadcrumbs.map((b) => b.name).join(' | ') + ' | Chaos Mesh Dashboard')
      }
    },
  },
})

export const { setNavigationBreadcrumbs } = navigationSlice.actions

export default navigationSlice.reducer
