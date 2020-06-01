import { Box, CssBaseline } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import Header from './Header'
import Sidebar from './Sidebar'
import StatusBar from 'components/StatusBar'
import chaosMeshRoutes from 'routes'
import { drawerWidth } from './Sidebar'
import { setNavigationBreadcrumbs } from 'slices/navigation'
import { useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
      flexDirection: 'column',
      [theme.breakpoints.up('sm')]: {
        width: `calc(100% - ${drawerWidth})`,
        marginLeft: drawerWidth,
      },
    },
    toolbar: theme.mixins.toolbar,
    main: {
      display: 'flex',
      flexDirection: 'column',
      minHeight: '100vh',
    },
    switchContent: {
      display: 'flex',
      flex: 1,
    },
  })
)

const TopContainer = () => {
  const classes = useStyles()

  const [openMobileDrawer, setOpenMobileDrawer] = useState(false)
  const handleDrawerToggle = () => setOpenMobileDrawer(!openMobileDrawer)

  const { pathname } = useLocation()
  const { breadcrumbs } = useSelector((state: RootState) => state.navigation)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(setNavigationBreadcrumbs(pathname))
  }, [dispatch, pathname])

  return (
    <Box className={classes.root}>
      {/* CssBaseline: kickstart an elegant, consistent, and simple baseline to build upon. */}
      <CssBaseline />
      <Header handleDrawerToggle={handleDrawerToggle} breadcrumbs={breadcrumbs} />
      <Sidebar openMobileDrawer={openMobileDrawer} handleDrawerToggle={handleDrawerToggle} />
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
      </main>
    </Box>
  )
}

export default TopContainer
