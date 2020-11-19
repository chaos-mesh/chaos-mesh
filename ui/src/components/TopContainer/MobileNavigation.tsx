import { AppBar, IconButton, Toolbar } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import AddBoxIcon from '@material-ui/icons/AddBox'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { NavLink } from 'react-router-dom'
import React from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
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
      bottom: theme.spacing(6),
      right: theme.spacing(6),
      zIndex: 1101, // .MuiAppBar-root z-index: 1100
    },
  })
)

const items = [
  { icon: <AddBoxIcon />, href: 'newExperiment' },
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
  {
    icon: <SettingsIcon />,
    href: 'settings',
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
      </Toolbar>
    </AppBar>
  )
}

export default MobileNavigation
