import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { drawerCloseWidth, drawerWidth } from './Sidebar'

import GitHubIcon from '@material-ui/icons/GitHub'
import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import { NavigationBreadCrumbProps } from 'slices/navigation.type'
import React from 'react'
import clsx from 'clsx'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
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
        color: '#fff',
      },
    },
    whiteText: {
      color: '#fff',
    },
    hoverLink: {
      '&:hover': {
        textDecoration: 'underline',
        cursor: 'pointer',
      },
    },
  })
)

interface HeaderProps {
  openDrawer: boolean
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Header: React.FC<HeaderProps> = ({ openDrawer, handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()

  return (
    <AppBar className={openDrawer ? classes.appBarShift : classes.appBar} position="fixed">
      <Toolbar>
        <IconButton
          className={classes.menuButton}
          color="inherit"
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
                  <Link key={b.name} to={b.path}>
                    <Typography className={clsx(classes.whiteText, classes.hoverLink)} variant="h6" component="h2">
                      {b.name}
                    </Typography>
                  </Link>
                ) : (
                  <Typography key={b.name} className={classes.whiteText} variant="h6" component="h2">
                    {b.name}
                  </Typography>
                )
              })}
          </Breadcrumbs>
          <IconButton
            component="a"
            href="https://github.com/pingcap/chaos-mesh"
            target="_blank"
            color="inherit"
            aria-label="Chaos Mesh GitHub link"
          >
            <GitHubIcon />
          </IconButton>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default Header
