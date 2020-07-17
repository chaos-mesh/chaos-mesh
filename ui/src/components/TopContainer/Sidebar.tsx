import { Box, Button, Divider, Drawer, List, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { Link } from 'react-router-dom'
import { NavLink } from 'react-router-dom'
import React from 'react'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'
import clsx from 'clsx'

export const drawerWidth = '14rem'
export const drawerCloseWidth = '5.25rem'
const useStyles = makeStyles((theme: Theme) =>
  createStyles({
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
    listItemIcon: {
      paddingLeft: theme.spacing(3),
      paddingRight: theme.spacing(9),
    },
    hidden: {
      display: 'none',
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
      <NavLink to="/" className={classes.toolbar}>
        <img
          className={open ? classes.logo : classes.logoMini}
          src={open ? '/logo.svg' : '/logo-mini.svg'}
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
          {open ? 'New Experiment' : <AddIcon />}
        </Button>
      </Box>

      <Divider />

      <List>
        {listItems.map(({ icon, text }) => (
          <ListItem key={text} className={classes.listItem} component={NavLink} to={`/${text.toLowerCase()}`} button>
            <ListItemIcon className={classes.listItemIcon}>{icon}</ListItemIcon>
            <ListItemText primary={text} />
          </ListItem>
        ))}
      </List>
    </Drawer>
  )
}

export default Sidebar
