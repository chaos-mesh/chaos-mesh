import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'

import MenuIcon from '@material-ui/icons/Menu'
import MenuOpenIcon from '@material-ui/icons/MenuOpen'
import Namespace from './Namespace'
import { NavigationBreadCrumbProps } from 'slices/navigation'
import Search from 'components/Search'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles((theme) => ({
  toolbar: {
    marginBottom: theme.spacing(6),
  },
  appBar: {
    position: 'absolute',
    width: `calc(100% - ${theme.spacing(12)})`,
    margin: theme.spacing(6),
  },
  menuButton: {
    [theme.breakpoints.down('md')]: {
      display: 'none',
    },
  },
  nav: {
    color: 'inherit',
  },
}))

function hasLocalBreadcrumb(b: string) {
  return ['dashboard', 'workflows', 'schedules', 'experiments', 'events', 'archives', 'settings'].includes(b)
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
              <Breadcrumbs className={classes.nav} aria-label="breadcrumb">
                <Typography variant="h6" component="h2">
                  {hasLocalBreadcrumb(b.name)
                    ? T(
                        `${
                          breadcrumbs[1] && breadcrumbs[1].name === 'new'
                            ? 'new' + b.name.charAt(0).toUpperCase()
                            : b.name
                        }.title`
                      )
                    : b.name}
                </Typography>
              </Breadcrumbs>
            )}
            <Space direction="row">
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
