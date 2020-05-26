import { NavigateAction, NavigationBreadCrumbProps } from './navigation.type'

import { createSlice } from '@reduxjs/toolkit'

function pathnameToBreadCrumbs(pathname: string) {
  const nameArray = pathname.slice(1).split('/')

  return nameArray.map((name, i) => {
    const b: NavigationBreadCrumbProps = {
      name: i === 0 ? name.charAt(0).toUpperCase() + name.slice(1) : name,
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
    setNavigationBreadcrumbs(state, action: NavigateAction) {
      state.breadcrumbs = pathnameToBreadCrumbs(action.payload)
    },
  },
})

export const { setNavigationBreadcrumbs } = navigationSlice.actions

export default navigationSlice.reducer
