import { Divider, List, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { NavLink } from 'react-router-dom'
import React from 'react'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    toolbar: {
      ...theme.mixins.toolbar,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    logo: {
      width: '75%',
    },
    listItem: {
      '&.active': {
        color: theme.palette.primary.main,
        '& svg': {
          fill: theme.palette.primary.main,
        },
        '& .MuiListItemText-primary': {
          fontWeight: 'bold',
        },
      },
    },
    itemIcon: {
      paddingLeft: theme.spacing(3),
    },
  })
)

const listItems = [
  { icon: <WebIcon />, text: 'Overview' },
  {
    icon: <TuneIcon />,
    text: 'Experiments',
  },
  {
    icon: <BlurLinearIcon />,
    text: 'Events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    text: 'Archives',
  },
]

export default function Nav() {
  const classes = useStyles()

  return (
    <>
      <NavLink to="/" className={classes.toolbar}>
        <img className={classes.logo} src="/logo.svg" alt="Chaos Mesh Logo" />
      </NavLink>
      <Divider />
      <List>
        {listItems.map(({ icon, text }) => (
          <ListItem key={text} className={classes.listItem} component={NavLink} to={`/${text.toLowerCase()}`} button>
            <ListItemIcon className={classes.itemIcon}>{icon}</ListItemIcon>
            <ListItemText primary={text} />
          </ListItem>
        ))}
      </List>
    </>
  )
}
