import { AppBar, Box, Breadcrumbs, IconButton, Toolbar, Typography } from '@material-ui/core'
import { Theme, makeStyles } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import Namespace from './Namespace'
import { NavigationBreadCrumbProps } from 'slices/navigation'
import React from 'react'
import SearchTrigger from 'components/SearchTrigger'
import Space from 'components-mui/Space'
import T from 'components/T'
import { useHistory } from 'react-router-dom'

const useStyles = makeStyles((theme: Theme) => ({
  fill: {
    margin: theme.spacing(6),
    marginBottom: 0,
  },
  appBar: {
    position: 'absolute',
    width: `calc(100% - ${theme.spacing(12)})`,
    margin: theme.spacing(6),
    borderRadius: theme.shape.borderRadius,
  },
  menuButton: {
    marginLeft: theme.spacing(0),
    [theme.breakpoints.down('sm')]: {
      display: 'none',
    },
  },
  nav: {
    marginLeft: theme.spacing(3),
    color: 'inherit',
    [theme.breakpoints.down('xs')]: {
      display: 'none',
    },
  },
  hoverLink: {
    textDecoration: 'none',
    '&:link': {
      color: 'inherit',
    },
    '&:visited': {
      color: 'inherit',
    },
    '&:hover': {
      textDecoration: 'underline',
      cursor: 'pointer',
    },
  },
  tail: {
    [theme.breakpoints.down('xs')]: {
      width: '100%',
    },
  },
}))

function hasLocalBreadcrumb(b: string) {
  return ['dashboard', 'newExperiment', 'experiments', 'events', 'archives', 'settings', 'swagger'].includes(b)
}

interface HeaderProps {
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Navbar: React.FC<HeaderProps> = ({ handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()
  const history = useHistory()

  return (
    <>
      <Toolbar className={classes.fill} />
      <AppBar className={classes.appBar}>
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
                    <Link key={b.name} to={b.path} className={classes.hoverLink}>
                      <Typography variant="h6" component="h2" color="inherit">
                        {hasLocalBreadcrumb(b.name) ? T(`${b.name}.title`) : b.name}
                      </Typography>
                    </Link>
                  ) : (
                    <Typography key={b.name} variant="h6" component="h2" color="inherit">
                      {hasLocalBreadcrumb(b.name) ? T(`${b.name === 'newExperiment' ? 'newE' : b.name}.title`) : b.name}
                    </Typography>
                  )
                })}
            </Breadcrumbs>
            <Space className={classes.tail} display="flex" justifyContent="space-between" alignItems="center">
              <Namespace />
              <Box>
                <SearchTrigger />
                <IconButton
                  className="nav-new-experiment"
                  color="inherit"
                  aria-label="New Experiment"
                  onClick={() => history.push('/newExperiment')}
                >
                  <AddIcon />
                </IconButton>
              </Box>
            </Space>
          </Box>
        </Toolbar>
      </AppBar>
    </>
  )
}

export default Navbar
