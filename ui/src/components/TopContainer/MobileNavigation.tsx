import { AppBar, IconButton, Toolbar, Fab } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import { Link } from 'react-router-dom'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { NavLink } from 'react-router-dom'
import React from 'react'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBar: {
      top: 'auto',
      bottom: 0,
    },
    fab: {
      position: 'fixed',
      bottom: theme.spacing(7.5),
      right: theme.spacing(3),
      zIndex: 1101, // .MuiAppBar-root z-index: 1100
    },
  })
)

const items = [
  { icon: <WebIcon />, href: 'overview' },
  {
    icon: <TuneIcon />,
    href: 'experiments',
  },
  {
    icon: <BlurLinearIcon />,
    href: 'events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    href: 'archives',
  },
]

const MobileNavigation = () => {
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar} position="fixed" color="inherit">
      <Toolbar>
        {items.map((i) => (
          <NavLink key={i.href} to={`/${i.href}`}>
            <IconButton color="primary">{i.icon}</IconButton>
          </NavLink>
        ))}
        <Fab
          component={Link}
          to="/newExperiment"
          className={classes.fab}
          color="primary"
          size="medium"
          aria-label="New experiment"
        >
          <AddIcon />
        </Fab>
      </Toolbar>
    </AppBar>
  )
}

export default MobileNavigation
