import { Alert, Box, CssBaseline, Paper, Portal, Snackbar, useMediaQuery, useTheme } from '@material-ui/core'
import { Redirect, Route, Switch } from 'react-router-dom'
import { drawerCloseWidth, drawerWidth } from './Sidebar'
import { setAlertOpen, setConfig, setConfirmOpen, setNameSpace, setTokenName, setTokens } from 'slices/globalStatus'
import { useEffect, useMemo, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import ContentContainer from 'components-mui/ContentContainer'
import { IntlProvider } from 'react-intl'
import LS from 'lib/localStorage'
import Loading from 'components-mui/Loading'
import Navbar from './Navbar'
import Sidebar from './Sidebar'
import api from 'api'
import flat from 'flat'
import insertCommonStyle from 'lib/d3/insertCommonStyle'
import loadable from '@loadable/component'
import { makeStyles } from '@material-ui/styles'
import messages from 'i18n/messages'
import routes from 'routes'
import { setNavigationBreadcrumbs } from 'slices/navigation'
import { useLocation } from 'react-router-dom'

const Auth = loadable(() => import('./Auth'))

const useStyles = makeStyles((theme) => ({
  root: {
    display: 'flex',
    flexDirection: 'column',
    marginLeft: drawerCloseWidth,
    width: `calc(100% - ${drawerCloseWidth})`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    [theme.breakpoints.down('sm')]: {
      minWidth: theme.breakpoints.values.md,
    },
  },
  rootShift: {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth})`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  main: {
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh',
    zIndex: 1,
  },
  switchContent: {
    display: 'flex',
    flex: 1,
  },
}))

const TopContainer = () => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('md'))
  const classes = useStyles()

  const { pathname } = useLocation()

  const { settings, globalStatus, navigation } = useStoreSelector((state) => state)
  const { lang } = settings
  const { alert, alertOpen, confirm, confirmOpen } = globalStatus
  const { breadcrumbs } = navigation

  const intlMessages = useMemo<Record<string, string>>(() => flat(messages[lang]), [lang])

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

  /**
   * Render different components according to server configuration.
   *
   */
  function fetchServerConfig() {
    api.common
      .config()
      .then(({ data }) => {
        if (data.security_mode) {
          setAuth()
        }

        dispatch(setConfig(data))
      })
      .finally(() => setLoading(false))
  }

  const [loading, setLoading] = useState(true)
  const [authOpen, setAuthOpen] = useState(false)

  /**
   * Set authorization (RBAC token) for API use.
   *
   */
  function setAuth() {
    const token = LS.get('token')
    const tokenName = LS.get('token-name')
    const globalNamespace = LS.get('global-namespace')

    if (token && tokenName) {
      const tokens = JSON.parse(token)

      api.auth.token(tokens.filter(({ name }: { name: string }) => name === tokenName)[0].token)
      dispatch(setTokens(tokens))
      dispatch(setTokenName(tokenName))
    } else {
      setAuthOpen(true)
    }

    if (globalNamespace) {
      api.auth.namespace(globalNamespace)
      dispatch(setNameSpace(globalNamespace))
    }
  }

  useEffect(() => {
    fetchServerConfig()
    insertCommonStyle()
    // eslint-disable-next-line
  }, [])

  useEffect(() => {
    dispatch(setNavigationBreadcrumbs(pathname))
  }, [dispatch, pathname])

  useEffect(() => {
    if (isTabletScreen) {
      setOpenDrawer(false)
    }
  }, [isTabletScreen])

  return (
    <IntlProvider messages={intlMessages} locale={lang} defaultLocale="en">
      <CssBaseline />

      <Box className={openDrawer ? classes.rootShift : classes.root}>
        <Sidebar open={openDrawer} />
        <Paper className={classes.main} component="main" elevation={0}>
          <Box className={classes.switchContent}>
            <ContentContainer>
              <Navbar openDrawer={openDrawer} handleDrawerToggle={handleDrawerToggle} breadcrumbs={breadcrumbs} />

              {loading ? (
                <Loading />
              ) : (
                <Switch>
                  <Redirect path="/" to="/dashboard" exact />
                  {!authOpen && routes.map((route) => <Route key={route.path as string} {...route} />)}
                  <Redirect to="/dashboard" />
                </Switch>
              )}
            </ContentContainer>
          </Box>
        </Paper>
      </Box>

      <Auth open={authOpen} setOpen={setAuthOpen} />

      <Portal>
        <Snackbar
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center',
          }}
          autoHideDuration={3000}
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
    </IntlProvider>
  )
}

export default TopContainer
