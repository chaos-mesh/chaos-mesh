import { AppBar, IconButton, Toolbar } from '@material-ui/core'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { NavLink } from 'react-router-dom'
import React from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
  appBar: {
    top: 'auto',
    bottom: 0,
  },
})

const items = [
  { icon: <WebIcon />, href: '/dashboard' },
  {
    icon: <TuneIcon />,
    href: '/experiments',
  },
  {
    icon: <BlurLinearIcon />,
    href: '/events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    href: '/archives',
  },
  {
    icon: <SettingsIcon />,
    href: '/settings',
  },
]

const MobileNavigation = () => {
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar} position="fixed" color="inherit">
      <Toolbar>
        {items.map((i) => (
          <NavLink key={i.href} to={i.href}>
            <IconButton color="primary">{i.icon}</IconButton>
          </NavLink>
        ))}
      </Toolbar>
    </AppBar>
  )
}

export default MobileNavigation
