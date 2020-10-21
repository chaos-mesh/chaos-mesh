import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { drawerCloseWidth, drawerWidth } from './Sidebar'

import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import { NavigationBreadCrumbProps } from 'slices/navigation.type'
import React from 'react'
import SearchTrigger from 'components/SearchTrigger'
import T from 'components/T'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBarCommon: {
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
      [theme.breakpoints.down('sm')]: {
        display: 'none',
      },
    },
    nav: {
      marginLeft: theme.spacing(3),
      '& .MuiBreadcrumbs-separator': {
        color: theme.palette.primary.main,
      },
    },
    hoverLink: {
      '&:hover': {
        color: theme.palette.primary.main,
        textDecoration: 'underline',
        cursor: 'pointer',
      },
    },
  })
)

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
        <Box display="flex" justifyContent="space-between" alignItems="center" width="100%">
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
        </Box>
        <SearchTrigger />
      </Toolbar>
    </AppBar>
  )
}

export default Header
