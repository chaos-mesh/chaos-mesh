import { Box, Drawer, List, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import DashboardOutlinedIcon from '@material-ui/icons/DashboardOutlined'
import GitHubIcon from '@material-ui/icons/GitHub'
import HttpOutlinedIcon from '@material-ui/icons/HttpOutlined'
import MenuBookOutlinedIcon from '@material-ui/icons/MenuBookOutlined'
import { NavLink } from 'react-router-dom'
import React from 'react'
import SettingsOutlinedIcon from '@material-ui/icons/SettingsOutlined'
import StorageOutlinedIcon from '@material-ui/icons/StorageOutlined'
import T from 'components/T'
import TimelineOutlinedIcon from '@material-ui/icons/TimelineOutlined'
import clsx from 'clsx'
import logo from 'images/logo.svg'
import logoMini from 'images/logo-mini.svg'
import logoMiniWhite from 'images/logo-mini-white.svg'
import logoWhite from 'images/logo-white.svg'
import { makeStyles } from '@material-ui/core/styles'
import { useStoreSelector } from 'store'

export const drawerWidth = '14rem'
export const drawerCloseWidth = '5rem'
const useStyles = makeStyles((theme) => {
  const listItemHover = {
    background: theme.palette.primary.main,
    cursor: 'pointer',
    '& svg': {
      fill: '#fff',
    },
    '& .MuiListItemText-primary': {
      color: '#fff',
    },
  }

  return {
    drawer: {
      width: drawerWidth,
    },
    drawerPaperRoot: {
      background: theme.palette.background.default,
      border: 'none',
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
      minHeight: 56,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      marginTop: theme.spacing(6),
    },
    logo: {
      width: '75%',
    },
    logoMini: {
      width: 36,
    },
    list: {
      padding: `${theme.spacing(6)} 0`,
    },
    listItem: {
      width: `calc(100% - ${theme.spacing(6)})`,
      height: 48,
      marginLeft: theme.spacing(3),
      marginBottom: theme.spacing(3),
      borderRadius: theme.shape.borderRadius,
      '&:last-child': {
        marginBottom: 0,
      },
      '&:hover': listItemHover,
      '&.active': {
        ...listItemHover,
      },
    },
    listItemIcon: {
      paddingRight: theme.spacing(9),
    },
  }
})

const listItems = [
  { icon: <DashboardOutlinedIcon />, text: 'dashboard' },
  {
    icon: <StorageOutlinedIcon />,
    text: 'experiments',
  },
  {
    icon: <TimelineOutlinedIcon />,
    text: 'events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    text: 'archives',
  },
  {
    icon: <SettingsOutlinedIcon />,
    text: 'settings',
  },
]

interface SidebarProps {
  open: boolean
}

const Sidebar: React.FC<SidebarProps> = ({ open }) => {
  const classes = useStyles()

  const { theme } = useStoreSelector((state) => state.settings)

  return (
    <Drawer
      className={clsx(classes.drawer, {
        [classes.drawerOpen]: open,
        [classes.drawerClose]: !open,
      })}
      classes={{
        paper: clsx(classes.drawerPaperRoot, {
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
              src={open ? (theme === 'light' ? logo : logoWhite) : theme === 'light' ? logoMini : logoMiniWhite}
              alt="Chaos Mesh"
            />
          </NavLink>

          <List className={classes.list}>
            {listItems.map(({ icon, text }) => (
              <ListItem
                key={text}
                className={clsx(classes.listItem, `sidebar-${text}`)}
                component={NavLink}
                to={`/${text}`}
                button
              >
                <ListItemIcon className={classes.listItemIcon}>{icon}</ListItemIcon>
                <ListItemText primary={T(`${text}.title`)} />
              </ListItem>
            ))}
          </List>
        </Box>

        <List className={classes.list}>
          <ListItem className={classes.listItem} component={NavLink} to="/swagger" button>
            <ListItemIcon className={classes.listItemIcon}>
              <HttpOutlinedIcon />
            </ListItemIcon>
            <ListItemText primary="Swagger API" />
          </ListItem>

          <ListItem
            className={classes.listItem}
            component="a"
            href="https://chaos-mesh.org/docs"
            target="_blank"
            button
          >
            <ListItemIcon className={classes.listItemIcon}>
              <MenuBookOutlinedIcon />
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
