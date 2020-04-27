import { AppBar, Box, CssBaseline, Drawer, Hidden, IconButton, Toolbar, Typography } from '@material-ui/core'
import React, { useState } from 'react'
import { Redirect, Route, BrowserRouter as Router, Switch } from 'react-router-dom'
import { Theme, createStyles, makeStyles, useTheme } from '@material-ui/core/styles'

import MenuIcon from '@material-ui/icons/Menu'
import NavMenu from './components/NavMenu'
import { ThemeProvider } from '@material-ui/core/styles'
import chaosMeshRoutes from './routes'
import chaosMeshTheme from './theme'

const drawerWidth = '14rem'
const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
    },
    drawer: {
      [theme.breakpoints.up('sm')]: {
        flexShrink: 0,
        width: drawerWidth,
      },
    },
    appBar: {
      [theme.breakpoints.up('sm')]: {
        width: `calc(100% - ${drawerWidth})`,
        marginLeft: drawerWidth,
        boxShadow: 'none',
      },
    },
    menuButton: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      marginLeft: theme.spacing(0),
    },
    // necessary for content to be below app bar
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
      width: drawerWidth,
    },
    content: {
      flexGrow: 1,
    },
  })
)

export default function App() {
  const classes = useStyles()
  const theme = useTheme()
  const [mobileOpen, setMobileOpen] = useState(false)

  const handleDrawerToggle = () => setMobileOpen(!mobileOpen)

  const Header = () => (
    <AppBar position="fixed" className={classes.appBar}>
      <Toolbar>
        <IconButton
          color="inherit"
          aria-label="open drawer"
          edge="start"
          onClick={handleDrawerToggle}
          className={classes.menuButton}
        >
          <MenuIcon />
        </IconButton>
        <Box p={2}>
          <Typography variant="h6">Dashboard</Typography>
        </Box>
      </Toolbar>
    </AppBar>
  )

  const Nav = () => (
    <nav className={classes.drawer} aria-label="mailbox folders">
      {/* The implementation can be swapped with js to avoid SEO duplication of links. */}
      <Hidden smUp implementation="css">
        <Drawer
          variant="temporary"
          anchor={theme.direction === 'rtl' ? 'right' : 'left'}
          open={mobileOpen}
          onClose={handleDrawerToggle}
          classes={{
            paper: classes.drawerPaper,
          }}
          ModalProps={{
            keepMounted: true, // Better open performance on mobile.
          }}
        >
          <NavMenu />
        </Drawer>
      </Hidden>
      <Hidden xsDown implementation="css">
        <Drawer
          classes={{
            paper: classes.drawerPaper,
          }}
          variant="permanent"
          open
        >
          <NavMenu />
        </Drawer>
      </Hidden>
    </nav>
  )

  return (
    <ThemeProvider theme={chaosMeshTheme}>
      <div className={classes.root}>
        {/* CssBaseline: kickstart an elegant, consistent, and simple baseline to build upon. */}
        <CssBaseline />
        <Router>
          <Header />
          <Nav />
          <main className={classes.content}>
            <div className={classes.toolbar} />
            <Switch>
              <Redirect exact={true} path="/" to="/overview" />
              {chaosMeshRoutes.map((route) => (
                <Route key={route.path!.toString()} {...route} />
              ))}
            </Switch>
          </main>
        </Router>
      </div>
    </ThemeProvider>
  )
}
