import { Box, Button, Divider, Drawer, List, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import DescriptionOutlinedIcon from '@material-ui/icons/DescriptionOutlined'
import GitHubIcon from '@material-ui/icons/GitHub'
import { Link } from 'react-router-dom'
import { NavLink } from 'react-router-dom'
import React from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
import T from 'components/T'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'
import clsx from 'clsx'
import logo from 'images/logo.svg'
import logoMini from 'images/logo-mini.svg'

export const drawerWidth = '14rem'
export const drawerCloseWidth = '5.25rem'
const useStyles = makeStyles((theme: Theme) => {
  const listItemHover = {
    color: theme.palette.primary.main,
    '& svg': {
      fill: theme.palette.primary.main,
    },
  }

  return createStyles({
    drawer: {
      width: drawerWidth,
    },
    drawerOpen: {
      width: drawerWidth,
      transition: theme.transitions.create('width', {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.enteringScreen,
      }),
    },
    drawerClose: {
      width: drawerCloseWidth,
      transition: theme.transitions.create('width', {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.leavingScreen,
      }),
      overflowX: 'hidden',
      [theme.breakpoints.down('xs')]: {
        display: 'none',
      },
    },
    toolbar: {
      ...theme.mixins.toolbar,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    logo: {
      width: '75%',
    },
    logoMini: {
      width: 36,
    },
    listItem: {
      '&:hover': listItemHover,
      '&.active': {
        ...listItemHover,
        '& .MuiListItemText-primary': {
          fontWeight: 'bold',
        },
      },
    },
    listItemIcon: {
      paddingLeft: theme.spacing(3),
      paddingRight: theme.spacing(9),
    },
    hidden: {
      display: 'none',
    },
  })
})

const listItems = [
  { icon: <WebIcon />, text: 'overview' },
  {
    icon: <TuneIcon />,
    text: 'experiments',
  },
  {
    icon: <BlurLinearIcon />,
    text: 'events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    text: 'archives',
  },
  {
    icon: <SettingsIcon />,
    text: 'settings',
  },
]

interface SidebarProps {
  open: boolean
}

const Sidebar: React.FC<SidebarProps> = ({ open }) => {
  const classes = useStyles()

  return (
    <Drawer
      className={clsx(classes.drawer, {
        [classes.drawerOpen]: open,
        [classes.drawerClose]: !open,
      })}
      classes={{
        paper: clsx({
          [classes.drawerOpen]: open,
          [classes.drawerClose]: !open,
        }),
      }}
      variant="permanent"
    >
      <Box display="flex" flexDirection="column" justifyContent="space-between" height="100%">
        <Box>
          <NavLink to="/" className={classes.toolbar}>
            <img
              className={open ? classes.logo : classes.logoMini}
              src={open ? logo : logoMini}
              alt="Chaos Mesh Logo"
            />
          </NavLink>
          <Divider />

          <Box display="flex" justifyContent="center" px={3} py="8px">
            <Button
              component={Link}
              to="/newExperiment"
              style={{ width: '100%' }}
              variant="outlined"
              color="primary"
              startIcon={open && <AddIcon />}
            >
              {open ? T('newE.title') : <AddIcon />}
            </Button>
          </Box>

          <Divider />

          <List>
            {listItems.map(({ icon, text }) => (
              <ListItem key={text} className={classes.listItem} component={NavLink} to={`/${text}`} button>
                <ListItemIcon className={classes.listItemIcon}>{icon}</ListItemIcon>
                <ListItemText primary={T(`${text}.title`)} />
              </ListItem>
            ))}
          </List>
        </Box>

        <List>
          <ListItem
            className={classes.listItem}
            component="a"
            href="https://chaos-mesh.org/docs"
            target="_blank"
            button
          >
            <ListItemIcon className={classes.listItemIcon}>
              <DescriptionOutlinedIcon />
            </ListItemIcon>
            <ListItemText primary={T('common.doc')} />
          </ListItem>

          <ListItem
            className={classes.listItem}
            component="a"
            href="https://github.com/chaos-mesh/chaos-mesh"
            target="_blank"
            button
          >
            <ListItemIcon className={classes.listItemIcon}>
              <GitHubIcon />
            </ListItemIcon>
            <ListItemText primary="GitHub" />
          </ListItem>
        </List>
      </Box>
    </Drawer>
  )
}

export default Sidebar
