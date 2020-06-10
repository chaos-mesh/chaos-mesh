import { Box, CssBaseline, Snackbar, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { drawerCloseWidth, drawerWidth } from './Sidebar'

import Alert from '@material-ui/lab/Alert'
import Header from './Header'
import MobileNavigation from './MobileNavigation'
import Sidebar from './Sidebar'
import StatusBar from 'components/StatusBar'
import chaosMeshRoutes from 'routes'
import { setAlertOpen } from 'slices/globalStatus'
import { setNavigationBreadcrumbs } from 'slices/navigation'
import { useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
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

  const { globalStatus, navigation } = useSelector((state: RootState) => state)
  const { alert, alertOpen } = globalStatus
  const { breadcrumbs } = navigation
  const dispatch = useStoreDispatch()
  const handleSnackClose = () => dispatch(setAlertOpen(false))

  const [openDrawer, setOpenDrawer] = useState(!isTabletScreen)
  const handleDrawerToggle = () => setOpenDrawer(!openDrawer)

  useEffect(() => {
    dispatch(setNavigationBreadcrumbs(pathname))
  }, [dispatch, pathname])

  useEffect(() => setOpenDrawer(!isTabletScreen), [isTabletScreen])

  return (
    <Box className={openDrawer ? classes.rootShift : classes.root}>
      {/* CssBaseline: kickstart an elegant, consistent, and simple baseline to build upon. */}
      <CssBaseline />
      <Header openDrawer={openDrawer} handleDrawerToggle={handleDrawerToggle} breadcrumbs={breadcrumbs} />
      {!isMobileScreen && <Sidebar open={openDrawer} />}
      <main className={classes.main}>
        <div className={classes.toolbar} />

        <StatusBar />

        <Box className={classes.switchContent}>
          <Switch>
            <Redirect path="/" to="/overview" exact />
            {chaosMeshRoutes.map((route) => (
              <Route key={route.path! as string} {...route} />
            ))}
          </Switch>
        </Box>

        {isMobileScreen && <MobileNavigation />}

        <Snackbar
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center',
          }}
          autoHideDuration={10000}
          open={alertOpen}
          onClose={handleSnackClose}
        >
          <Alert variant="outlined" severity={alert.type} onClose={handleSnackClose}>
            {alert.message}
          </Alert>
        </Snackbar>
      </main>
    </Box>
  )
}

export default TopContainer
