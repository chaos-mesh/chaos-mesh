import { AppBar, Box, Breadcrumbs, IconButton, MenuItem, TextField, Toolbar, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { drawerCloseWidth, drawerWidth } from './Sidebar'
import { useHistory, useLocation } from 'react-router-dom'

import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import { NavigationBreadCrumbProps } from 'slices/navigation'
import T from 'components/T'
import api from 'api'
import { makeStyles } from '@material-ui/core/styles'
import { setNameSpace } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme) => ({
  appBarCommon: {
    maxHeight: 64,
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  appBar: {
    marginLeft: drawerCloseWidth,
    width: `calc(100% - ${drawerCloseWidth})`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    [theme.breakpoints.down('xs')]: {
      width: '100%',
    },
  },
  appBarShift: {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth})`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  menuButton: {
    marginLeft: theme.spacing(0),
    marginRight: theme.spacing(3),
    [theme.breakpoints.down('sm')]: {
      display: 'none',
    },
  },
  nav: {
    '& .MuiBreadcrumbs-separator': {
      color: theme.palette.primary.main,
    },
    [theme.breakpoints.down('sm')]: {
      marginLeft: theme.spacing(3),
    },
    [theme.breakpoints.down('xs')]: {
      marginLeft: theme.spacing(4),
    },
  },
  hoverLink: {
    '&:hover': {
      color: theme.palette.primary.main,
      textDecoration: 'underline',
      cursor: 'pointer',
    },
  },
  namespaces: {
    minWidth: 180,
    margin: 0,
  },
}))

function hasLocalBreadcrumb(b: string) {
  return ['overview', 'experiments', 'newExperiment', 'events', 'archives', 'settings'].includes(b)
}

interface HeaderProps {
  openDrawer: boolean
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Header: React.FC<HeaderProps> = ({ openDrawer, handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()
  const history = useHistory()
  const { pathname } = useLocation()

  const { namespace } = useSelector((state: RootState) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const [namespaces, setNamespaces] = useState(['All'])

  const fetchNamespaces = () => {
    api.common
      .chaosAvailableNamespaces()
      .then(({ data }) => setNamespaces(['All', ...data]))
      .catch(console.error)
  }

  const handleSelectGlobalNamespace = (e: React.ChangeEvent<HTMLInputElement>) => {
    const ns = e.target.value

    api.auth.namespace(ns)
    dispatch(setNameSpace(ns))

    history.replace('/namespaceSetted')
    setTimeout(() => history.replace(pathname))
  }

  useEffect(fetchNamespaces, [])

  return (
    <AppBar
      className={`${openDrawer ? classes.appBarShift : classes.appBar} ${classes.appBarCommon}`}
      position="fixed"
      color="inherit"
      elevation={0}
    >
      <Toolbar>
        <IconButton
          className={classes.menuButton}
          color="primary"
          edge="start"
          aria-label="Toggle drawer"
          onClick={handleDrawerToggle}
        >
          <MenuIcon />
        </IconButton>
        <Box display="flex" alignItems="center" width="100%">
          <Breadcrumbs className={classes.nav}>
            {breadcrumbs.length > 0 &&
              breadcrumbs.map((b) => {
                return b.path ? (
                  <Link key={b.name} to={b.path} style={{ textDecoration: 'none' }}>
                    <Typography className={classes.hoverLink} variant="h6" component="h2" color="textSecondary">
                      {hasLocalBreadcrumb(b.name) ? T(`${b.name}.title`) : b.name}
                    </Typography>
                  </Link>
                ) : (
                  <Typography key={b.name} variant="h6" component="h2" color="primary">
                    {hasLocalBreadcrumb(b.name) ? T(`${b.name === 'newExperiment' ? 'newE' : b.name}.title`) : b.name}
                  </Typography>
                )
              })}
          </Breadcrumbs>
          <Box ml={6}>
            <TextField
              className={classes.namespaces}
              variant="outlined"
              color="primary"
              margin="dense"
              select
              label={T('common.chooseNamespace')}
              value={namespace}
              onChange={handleSelectGlobalNamespace}
            >
              {namespaces.map((option) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </TextField>
          </Box>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default Header
