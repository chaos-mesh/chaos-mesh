import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import GitHubIcon from '@material-ui/icons/GitHub'
import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import { NavigationBreadCrumbProps } from 'slices/navigation.type'
import React from 'react'
import { drawerWidth } from './Sidebar'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBar: {
      [theme.breakpoints.up('md')]: {
        width: `calc(100% - ${drawerWidth})`,
        marginLeft: drawerWidth,
        boxShadow: 'none',
      },
    },
    menuButton: {
      [theme.breakpoints.up('md')]: {
        display: 'none',
      },
      marginLeft: theme.spacing(0),
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
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Header: React.FC<HeaderProps> = ({ handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar} position="fixed">
      <Toolbar>
        <IconButton
          className={classes.menuButton}
          color="inherit"
          edge="start"
          onClick={handleDrawerToggle}
          aria-label="Open drawer"
        >
          <MenuIcon />
        </IconButton>
        <Box display="flex" justifyContent="space-between" alignItems="center" width="100%">
          <Breadcrumbs className={classes.nav} aria-label="breadcrumb">
            {breadcrumbs.length > 0 &&
              breadcrumbs.map((b) => {
                return b.path ? (
                  <Link key={b.name} to={b.path}>
                    <Typography className={`${classes.whiteText} ${classes.hoverLink}`} variant="h6" component="h2">
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
