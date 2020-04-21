import React from 'react'
import { NavLink } from 'react-router-dom'

import {
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
} from '@material-ui/core'
import WebIcon from '@material-ui/icons/Web'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import TuneIcon from '@material-ui/icons/Tune'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'

import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    listItem: {
      '&.active': {
        color: theme.palette.primary.main,
        '& svg': {
          fill: theme.palette.primary.main,
        },
        '& .MuiListItemText-primary': {
          fontWeight: 500,
        },
      },
    },
    itemIcon: {
      paddingLeft: theme.spacing(3),
    },
    // necessary for content to be below app bar
    toolbar: {
      ...theme.mixins.toolbar,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    logo: {
      width: '12rem',
    },
  })
)

const icons = [
  <WebIcon />,
  <TuneIcon />,
  <BlurLinearIcon />,
  <ArchiveOutlinedIcon />,
]

export default function SideMenu() {
  const classes = useStyles()

  return (
    <>
      <NavLink to="/" className={classes.toolbar}>
        <img className={classes.logo} src="/logo.svg" alt="Chaos Mesh" />
      </NavLink>
      <Divider />
      <List>
        {['Overview', 'Experiments', 'Events', 'Archives'].map(
          (text, index) => (
            <ListItem
              button
              component={NavLink}
              to={`/${text.toLowerCase()}`}
              key={text}
              className={classes.listItem}
            >
              <ListItemIcon className={classes.itemIcon}>
                {icons[index]}
              </ListItemIcon>
              <ListItemText primary={text} />
            </ListItem>
          )
        )}
      </List>
    </>
  )
}
