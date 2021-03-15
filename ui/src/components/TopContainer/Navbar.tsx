import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'

import MenuIcon from '@material-ui/icons/Menu'
import MenuOpenIcon from '@material-ui/icons/MenuOpen'
import Namespace from './Namespace'
import { NavigationBreadCrumbProps } from 'slices/navigation'
import React from 'react'
import Search from 'components/Search'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  toolbar: {
    margin: theme.spacing(6),
    marginBottom: 0,
  },
  appBar: {
    position: 'absolute',
    width: `calc(100% - ${theme.spacing(12)})`,
    margin: theme.spacing(6),
  },
  menuButton: {
    [theme.breakpoints.down('sm')]: {
      display: 'none',
    },
  },
  nav: {
    color: 'inherit',
    [theme.breakpoints.down('xs')]: {
      display: 'none',
    },
  },
  navRight: {
    display: 'flex',
    alignItems: 'center',
    [theme.breakpoints.down('xs')]: {
      width: '100%',
    },
  },
}))

function hasLocalBreadcrumb(b: string) {
  return [
    'dashboard',
    'newExperiment',
    'experiments',
    'workflows',
    'events',
    'archives',
    'settings',
    'swagger',
  ].includes(b)
}

interface HeaderProps {
  openDrawer: boolean
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Navbar: React.FC<HeaderProps> = ({ openDrawer, handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()

  const b = breadcrumbs[0] // first breadcrumb

  return (
    <>
      <Toolbar className={classes.toolbar} />
      <AppBar className={classes.appBar} color="inherit" elevation={0}>
        <Toolbar disableGutters>
          <IconButton
            className={classes.menuButton}
            color="inherit"
            edge="start"
            aria-label="Toggle drawer"
            onClick={handleDrawerToggle}
          >
            {openDrawer ? <MenuOpenIcon /> : <MenuIcon />}
          </IconButton>
          <Box display="flex" justifyContent="space-between" alignItems="center" width="100%">
            {b && (
              <Breadcrumbs className={classes.nav}>
                <Typography variant="h6" component="h2">
                  {hasLocalBreadcrumb(b.name) ? T(`${b.name === 'newExperiment' ? 'newE' : b.name}.title`) : b.name}
                </Typography>
              </Breadcrumbs>
            )}
            <Space className={classes.navRight}>
              <Search />
              <Namespace />
            </Space>
          </Box>
        </Toolbar>
      </AppBar>
    </>
  )
}

export default Navbar
