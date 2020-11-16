import { Box, CssBaseline, Snackbar, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect, useMemo, useState } from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'
import { RootState, useStoreDispatch } from 'store'
import { ThemeProvider, createStyles, makeStyles } from '@material-ui/core/styles'
import customTheme, { darkTheme as customDarkTheme } from 'theme'
import { drawerCloseWidth, drawerWidth } from './Sidebar'

import Alert from '@material-ui/lab/Alert'
import ContentContainer from 'components-mui/ContentContainer'
import Header from './Header'
import { IntlProvider } from 'react-intl'
import MobileNavigation from './MobileNavigation'
import Sidebar from './Sidebar'
import chaosMeshRoutes from 'routes'
import flat from 'flat'
import insertCommonStyle from 'lib/d3/insertCommonStyle'
import messages from 'i18n/messages'
import { setAlertOpen } from 'slices/globalStatus'
import { setNavigationBreadcrumbs } from 'slices/navigation'
import { useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme) =>
  createStyles({
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
      display: 'flex',
      flexDirection: 'column',
      minHeight: '100vh',
    },
    toolbar: theme.mixins.toolbar,
    switchContent: {
      display: 'flex',
      flex: 1,
    },
  })
)

const TopContainer = () => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const isMobileScreen = useMediaQuery(theme.breakpoints.down('xs'))
  const classes = useStyles()

  const { pathname } = useLocation()

  const { settings, globalStatus, navigation } = useSelector((state: RootState) => state)
  const { theme: settingTheme, lang } = settings
  const { alert, alertOpen } = globalStatus
  const { breadcrumbs } = navigation

  const globalTheme = useMemo(() => (settingTheme === 'light' ? customTheme : customDarkTheme), [settingTheme])
  const intlMessages = useMemo<Record<string, string>>(() => flat(messages[lang]), [lang])

  const dispatch = useStoreDispatch()
  const handleSnackClose = () => dispatch(setAlertOpen(false))

  const miniSidebar = window.localStorage.getItem('chaos-mesh-mini-sidebar') === 'y'
  const [openDrawer, setOpenDrawer] = useState(!miniSidebar)
  const handleDrawerToggle = () => {
    setOpenDrawer(!openDrawer)
    window.localStorage.setItem('chaos-mesh-mini-sidebar', openDrawer ? 'y' : 'n')
  }

  useEffect(insertCommonStyle, [])

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
        <Box className={openDrawer ? classes.rootShift : classes.root}>
          {/* CssBaseline: kickstart an elegant, consistent, and simple baseline to build upon. */}
          <CssBaseline />
          <Header openDrawer={openDrawer} handleDrawerToggle={handleDrawerToggle} breadcrumbs={breadcrumbs} />
          {!isMobileScreen && <Sidebar open={openDrawer} />}
          <main className={classes.main}>
            <div className={classes.toolbar} />

            <Box className={classes.switchContent}>
              <ContentContainer>
                <Switch>
                  <Redirect exact path="/" to="/overview" />
                  {chaosMeshRoutes.map((route) => (
                    <Route key={route.path! as string} {...route} />
                  ))}
                  <Redirect to="/overview" />
                </Switch>
              </ContentContainer>
            </Box>

            {isMobileScreen && (
              <>
                <div className={classes.toolbar} />
                <MobileNavigation />
              </>
            )}

            <Snackbar
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'center',
              }}
              autoHideDuration={10000}
              open={alertOpen}
              onClose={handleSnackClose}
            >
              <Alert severity={alert.type} onClose={handleSnackClose}>
                {alert.message}
              </Alert>
            </Snackbar>
          </main>
        </Box>
      </IntlProvider>
    </ThemeProvider>
  )
}

export default TopContainer
