import { Box, CssBaseline, Paper, Portal, Snackbar, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect, useMemo, useState } from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'
import { ThemeProvider, makeStyles } from '@material-ui/core/styles'
import customTheme, { darkTheme as customDarkTheme } from 'theme'
import { drawerCloseWidth, drawerWidth } from './Sidebar'
import { setAlertOpen, setConfig, setNameSpace, setTokenName, setTokens } from 'slices/globalStatus'
import { useStoreDispatch, useStoreSelector } from 'store'

import Alert from '@material-ui/lab/Alert'
import Auth from './Auth'
import ContentContainer from 'components-mui/ContentContainer'
import { IntlProvider } from 'react-intl'
import LS from 'lib/localStorage'
import Loading from 'components-mui/Loading'
import MobileNavigation from './MobileNavigation'
import Navbar from './Navbar'
import Sidebar from './Sidebar'
import api from 'api'
import flat from 'flat'
import insertCommonStyle from 'lib/d3/insertCommonStyle'
import messages from 'i18n/messages'
import routes from 'routes'
import { setNavigationBreadcrumbs } from 'slices/navigation'
import { useLocation } from 'react-router-dom'

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
    [theme.breakpoints.down('xs')]: {
      width: '100%',
      marginLeft: 0,
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
    position: 'relative',
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh',
    zIndex: 1,
  },
  toolbar: theme.mixins.toolbar,
  switchContent: {
    display: 'flex',
    flex: 1,
  },
}))

const TopContainer = () => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const isMobileScreen = useMediaQuery(theme.breakpoints.down('xs'))
  const classes = useStyles()

  const { pathname } = useLocation()

  const { settings, globalStatus, navigation } = useStoreSelector((state) => state)
  const { theme: settingsTheme, lang } = settings
  const { alert, alertOpen } = globalStatus
  const { breadcrumbs } = navigation

  const globalTheme = useMemo(() => (settingsTheme === 'light' ? customTheme : customDarkTheme), [settingsTheme])
  const intlMessages = useMemo<Record<string, string>>(() => flat(messages[lang]), [lang])

  const dispatch = useStoreDispatch()
  const handleSnackClose = () => dispatch(setAlertOpen(false))

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
    <ThemeProvider theme={globalTheme}>
      <IntlProvider messages={intlMessages} locale={lang} defaultLocale="en">
        <CssBaseline />

        <Box className={openDrawer ? classes.rootShift : classes.root}>
          {!isMobileScreen && <Sidebar open={openDrawer} />}
          <Paper className={classes.main} component="main" elevation={0}>
            {/* <ControlBar /> */}
            <Navbar handleDrawerToggle={handleDrawerToggle} breadcrumbs={breadcrumbs} />

            <Box className={classes.switchContent}>
              <ContentContainer>
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

            {isMobileScreen && (
              <>
                <div className={classes.toolbar} />
                <MobileNavigation />
              </>
            )}

            <Auth open={authOpen} setOpen={setAuthOpen} />

            <Portal>
              <Snackbar
                anchorOrigin={{
                  vertical: 'bottom',
                  horizontal: 'center',
                }}
                autoHideDuration={9000}
                open={alertOpen}
                onClose={handleSnackClose}
              >
                <Alert severity={alert.type} onClose={handleSnackClose}>
                  {alert.message}
                </Alert>
              </Snackbar>
            </Portal>
          </Paper>
        </Box>
      </IntlProvider>
    </ThemeProvider>
  )
}

export default TopContainer
