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
import { applyAPIAuthentication, applyNSParam } from '@/api/interceptors'
import { Stale } from '@/api/queryUtils'
import ConfirmDialog from '@/mui-extends/ConfirmDialog'
import Loading from '@/mui-extends/Loading'
import { useGetCommonConfig } from '@/openapi'
import { useStoreDispatch, useStoreSelector } from '@/store'
import {
  Alert,
  Box,
  BoxProps,
  Container,
  CssBaseline,
  Divider,
  Portal,
  Snackbar,
  useMediaQuery,
  useTheme,
} from '@mui/material'
import { styled } from '@mui/material/styles'
import Cookies from 'js-cookie'
import { lazy, useEffect, useState } from 'react'
import { Outlet } from 'react-router'

import { setAlertOpen, setAuthOpen, setConfirmOpen, setNameSpace, setTokenName, setTokens } from '@/slices/globalStatus'

import { TokenFormValues } from '@/components/Token'

import insertCommonStyle from '@/lib/d3/insertCommonStyle'
import LS from '@/lib/localStorage'

import Navbar from './Navbar'
import { closedWidth, openedWidth } from './Sidebar'
import Sidebar from './Sidebar'

const Auth = lazy(() => import('./Auth'))

const Root = styled(Box, {
  shouldForwardProp: (prop) => prop !== 'open',
})<BoxProps & { open: boolean }>(({ theme, open }) => ({
  position: 'relative',
  width: `calc(100% - ${open ? openedWidth : closedWidth}px)`,
  height: '100vh',
  marginLeft: open ? openedWidth : closedWidth,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration[open ? 'enteringScreen' : 'leavingScreen'],
  }),
  [theme.breakpoints.down('sm')]: {
    minWidth: theme.breakpoints.values.md,
  },
}))

const TopContainer = () => {
  const theme = useTheme()

  const { alert, alertOpen, confirm, confirmOpen, authOpen } = useStoreSelector((state) => state.globalStatus)

  const dispatch = useStoreDispatch()
  const handleSnackClose = () => dispatch(setAlertOpen(false))
  const handleConfirmClose = () => dispatch(setConfirmOpen(false))

  // Sidebar related
  const miniSidebar = LS.get('mini-sidebar') === 'y'
  const [openDrawer, setOpenDrawer] = useState(!miniSidebar)
  const handleDrawerToggle = () => {
    setOpenDrawer(!openDrawer)
    LS.set('mini-sidebar', openDrawer ? 'y' : 'n')
  }

  const [loading, setLoading] = useState(true)

  const { data } = useGetCommonConfig({
    query: {
      staleTime: Stale.DAY,
    },
  })

  useEffect(() => {
    /**
     * Set authorization (RBAC token / GCP) for API use.
     */
    function setAuth() {
      // GCP
      const accessToken = Cookies.get('access_token')
      const expiry = Cookies.get('expiry')

      if (accessToken && expiry) {
        const token = {
          accessToken,
          expiry,
        }

        applyAPIAuthentication(token)
        dispatch(setTokenName('gcp'))

        return
      }

      const token = LS.get('token')
      const tokenName = LS.get('token-name')
      const globalNamespace = LS.get('global-namespace')

      if (token && tokenName) {
        const tokens: TokenFormValues[] = JSON.parse(token)

        applyAPIAuthentication(tokens.find(({ name }) => name === tokenName)!.token)
        dispatch(setTokens(tokens))
        dispatch(setTokenName(tokenName))
      } else {
        dispatch(setAuthOpen(true))
      }

      if (globalNamespace) {
        applyNSParam(globalNamespace)
        dispatch(setNameSpace(globalNamespace))
      }
    }

    if (data) {
      if (data.security_mode) {
        setAuth()
      }

      setLoading(false)
    }
  }, [data, dispatch])

  useEffect(() => {
    insertCommonStyle()
  }, [])

  const isTabletScreen = useMediaQuery(theme.breakpoints.down('md'))
  useEffect(() => {
    if (isTabletScreen) {
      setOpenDrawer(false)
    }
  }, [isTabletScreen])

  return (
    <>
      <CssBaseline />
      <Root open={openDrawer}>
        <Sidebar open={openDrawer} />
        <Box component="main" sx={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
          <Navbar openDrawer={openDrawer} handleDrawerToggle={handleDrawerToggle} />
          <Divider />

          <Container maxWidth="xl" disableGutters sx={{ flexGrow: 1, p: 6 }}>
            {loading ? <Loading /> : <Outlet />}
          </Container>
        </Box>
      </Root>

      <Auth open={authOpen} />

      <Portal>
        <Snackbar
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center',
          }}
          autoHideDuration={6000}
          open={alertOpen}
          onClose={handleSnackClose}
        >
          <Alert severity={alert.type} onClose={handleSnackClose}>
            {alert.message}
          </Alert>
        </Snackbar>
      </Portal>

      <Portal>
        <ConfirmDialog
          open={confirmOpen}
          close={handleConfirmClose}
          title={confirm.title}
          description={confirm.description}
          onConfirm={confirm.handle}
        />
      </Portal>
    </>
  )
}

export default TopContainer
