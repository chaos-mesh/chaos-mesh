import { Drawer, Hidden } from '@material-ui/core'
import { Theme, createStyles, makeStyles, useTheme } from '@material-ui/core/styles'

import NavMenu from './Nav'
import React from 'react'

export const drawerWidth = '14rem'
const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    drawer: {
      [theme.breakpoints.up('sm')]: {
        width: drawerWidth,
      },
    },
    drawerPaper: {
      width: drawerWidth,
    },
  })
)

interface SidebarProps {
  openMobileDrawer: boolean
  handleDrawerToggle: () => void
}

const Sidebar: React.FC<SidebarProps> = ({ openMobileDrawer, handleDrawerToggle }) => {
  const theme = useTheme()
  const classes = useStyles()

  return (
    <nav className={classes.drawer}>
      {/* The implementation can be swapped with js to avoid SEO duplication of links. */}
      <Hidden implementation="css" smUp>
        <Drawer
          classes={{
            paper: classes.drawerPaper,
          }}
          ModalProps={{
            keepMounted: true, // Better open performance on mobile.
          }}
          anchor={theme.direction === 'rtl' ? 'right' : 'left'}
          variant="temporary"
          open={openMobileDrawer}
          onClose={handleDrawerToggle}
        >
          <NavMenu />
        </Drawer>
      </Hidden>
      <Hidden implementation="css" xsDown>
        <Drawer
          classes={{
            paper: classes.drawerPaper,
          }}
          variant="permanent"
          open
        >
          <NavMenu />
        </Drawer>
      </Hidden>
    </nav>
  )
}

export default Sidebar
