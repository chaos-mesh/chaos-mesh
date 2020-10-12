import { AppBar, Box, Breadcrumbs, Button, IconButton, Paper, Toolbar, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { drawerCloseWidth, drawerWidth } from './Sidebar'

import { Link } from 'react-router-dom'
import MenuIcon from '@material-ui/icons/Menu'
import Modal from '@material-ui/core/Modal'
import { NavigationBreadCrumbProps } from 'slices/navigation.type'
import { RootState } from 'store'
import Search from 'components/Search'
import SearchIcon from '@material-ui/icons/Search'
import T from 'components/T'
import { setSearchModalOpen } from 'slices/globalStatus'
import store from 'store'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBarCommon: {
      borderBottom: `1px solid ${theme.palette.divider}`,
    },
    appBar: {
      marginLeft: drawerCloseWidth,
      width: `calc(100% - ${drawerCloseWidth})`,
      transition: theme.transitions.create(['width', 'margin'], {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.leavingScreen,
      }),
      [theme.breakpoints.down('xs')]: {
        width: '100%',
      },
    },
    appBarShift: {
      marginLeft: drawerWidth,
      width: `calc(100% - ${drawerWidth})`,
      transition: theme.transitions.create(['width', 'margin'], {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.enteringScreen,
      }),
    },
    menuButton: {
      marginLeft: theme.spacing(0),
      [theme.breakpoints.down('sm')]: {
        display: 'none',
      },
    },
    nav: {
      marginLeft: theme.spacing(4),
      '& .MuiBreadcrumbs-separator': {
        color: theme.palette.primary.main,
      },
    },
    hoverLink: {
      '&:hover': {
        color: theme.palette.primary.main,
        textDecoration: 'underline',
        cursor: 'pointer',
      },
    },
    searchTrigger: {
      borderRadius: '4px',
      color: '#969faf',
      cursor: 'pointer',
      '&:hover': {
        color: '#1c1e21',
      },
    },
    searchModal: {
      position: 'relative',
      maxWidth: '35rem',
      minHeight: '12.8125rem',
      margin: '3.75rem auto auto',
      padding: 12,
      background: '#f5f6f7',
      outline: 0,
    },
  })
)

function hasLocalBreadcrumb(b: string) {
  return ['overview', 'experiments', 'newExperiment', 'events', 'archives', 'settings'].includes(b)
}

const SearchTrigger: React.FC = () => {
  const classes = useStyles()
  const [open, setOpen] = useState(false)

  const handleOpen = () => {
    setOpen(true)
    store.dispatch(setSearchModalOpen(true))
  }

  const handleClose = () => {
    setOpen(false)
    store.dispatch(setSearchModalOpen(false))
  }

  const searchModalOpen = useSelector((state: RootState) => state.globalStatus.searchModalOpen)

  useEffect(() => {
    if (!searchModalOpen) handleClose()
  }, [searchModalOpen])

  return (
    <>
      <Button
        variant="outlined"
        className={classes.searchTrigger}
        startIcon={<SearchIcon color="primary" />}
        onClick={handleOpen}
      >
        Search
      </Button>
      <Modal open={open} onClose={handleClose}>
        <Paper elevation={3} className={classes.searchModal}>
          <Search></Search>
        </Paper>
      </Modal>
    </>
  )
}

interface HeaderProps {
  openDrawer: boolean
  handleDrawerToggle: () => void
  breadcrumbs: NavigationBreadCrumbProps[]
}

const Header: React.FC<HeaderProps> = ({ openDrawer, handleDrawerToggle, breadcrumbs }) => {
  const classes = useStyles()

  return (
    <AppBar
      className={`${openDrawer ? classes.appBarShift : classes.appBar} ${classes.appBarCommon}`}
      position="fixed"
      color="inherit"
      elevation={0}
    >
      <Toolbar>
        <IconButton
          className={classes.menuButton}
          color="primary"
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
                  <Link key={b.name} to={b.path} style={{ textDecoration: 'none' }}>
                    <Typography className={classes.hoverLink} variant="h6" component="h2" color="textSecondary">
                      {hasLocalBreadcrumb(b.name) ? T(`${b.name}.title`) : b.name}
                    </Typography>
                  </Link>
                ) : (
                  <Typography key={b.name} variant="h6" component="h2" color="primary">
                    {hasLocalBreadcrumb(b.name) ? T(`${b.name === 'newExperiment' ? 'newE' : b.name}.title`) : b.name}
                  </Typography>
                )
              })}
          </Breadcrumbs>
        </Box>
        <SearchTrigger />
      </Toolbar>
    </AppBar>
  )
}

export default Header
