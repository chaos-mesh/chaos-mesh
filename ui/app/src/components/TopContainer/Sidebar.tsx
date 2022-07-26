/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined'
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined'
import MenuBookOutlinedIcon from '@mui/icons-material/MenuBookOutlined'
import ScheduleIcon from '@mui/icons-material/Schedule'
import ScienceOutlinedIcon from '@mui/icons-material/ScienceOutlined'
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined'
import TimelineOutlinedIcon from '@mui/icons-material/TimelineOutlined'
import {
  Box,
  CSSObject,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemTextProps,
  Drawer as MuiDrawer,
  ListItemText as MuiListItemText,
} from '@mui/material'
import { Theme, styled } from '@mui/material/styles'
import { NavLink } from 'react-router-dom'

import { useStoreSelector } from 'store'

import i18n from 'components/T'

import logoMiniWhite from 'images/logo-mini-white.svg'
import logoMini from 'images/logo-mini.svg'
import logoWhite from 'images/logo-white.svg'
import logo from 'images/logo.svg'

export const openedWidth = 256
export const closedWidth = 64

const openedMixin = (theme: Theme): CSSObject => ({
  width: openedWidth,
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  overflowX: 'hidden',
})

const closedMixin = (theme: Theme): CSSObject => ({
  width: closedWidth,
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  overflowX: 'hidden',
})

const Drawer = styled(MuiDrawer, { shouldForwardProp: (prop) => prop !== 'open' })(({ theme, open }) => ({
  width: openedWidth,
  '& .MuiDrawer-paper': {
    ...(open ? openedMixin(theme) : closedMixin(theme)),
  },
  '& .MuiListItemButton-root': {
    paddingLeft: !open ? theme.spacing(3) : 16, // original paddingLeft is 16
  },
}))

const SidebarNavHoverProperties = (theme: Theme) => ({
  background: theme.palette.secondaryContainer.main,
  color: theme.palette.onSecondaryContainer.main,
  borderRadius: 4,
  '& .MuiListItemIcon-root': {
    color: theme.palette.onSecondaryContainer.main,
  },
})

const SidebarNav = styled(List)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  color: theme.palette.onSurfaceVariant.main,
  '& .MuiListItemButton-root': {
    width: '80%',
    marginBottom: theme.spacing(4),
    paddingTop: theme.spacing(1),
    paddingBottom: theme.spacing(1),
    '&:hover, &.active': SidebarNavHoverProperties(theme),
  },
  '& .MuiListItemIcon-root': {
    minWidth: 0,
    marginRight: theme.spacing(8),
    color: theme.palette.onSurfaceVariant.main,
  },
}))

const ListItemText = (props: ListItemTextProps) => (
  <MuiListItemText
    {...props}
    primaryTypographyProps={{
      ...props.primaryTypographyProps,
      fontWeight: 'medium',
    }}
  />
)

const listItems = [
  { icon: <DashboardOutlinedIcon />, text: 'dashboard' },
  {
    icon: <AccountTreeOutlinedIcon />,
    text: 'workflows',
  },
  {
    icon: <ScheduleIcon />,
    text: 'schedules',
  },
  {
    icon: <ScienceOutlinedIcon />,
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
  const { theme } = useStoreSelector((state) => state.settings)

  return (
    <Drawer variant="permanent" open={open}>
      <Box sx={{ width: open ? 160 : 32, m: '0 auto', my: 8 }}>
        <NavLink to="/">
          <img
            src={open ? (theme === 'light' ? logo : logoWhite) : theme === 'light' ? logoMini : logoMiniWhite}
            alt="Chaos Mesh"
          />
        </NavLink>
      </Box>
      <Box display="flex" flexDirection="column" justifyContent="space-between" height="100%">
        <SidebarNav>
          {listItems.map(({ icon, text }) => (
            <ListItemButton key={text} className={`tutorial-${text}`} component={NavLink} to={'/' + text}>
              <ListItemIcon>{icon}</ListItemIcon>
              <ListItemText primary={i18n(`${text}.title`)} />
            </ListItemButton>
          ))}
        </SidebarNav>

        <SidebarNav>
          <ListItemButton component="a" href="https://chaos-mesh.org/docs" target="_blank">
            <ListItemIcon>
              <MenuBookOutlinedIcon />
            </ListItemIcon>
            <ListItemText primary={i18n('common.doc')} />
          </ListItemButton>
        </SidebarNav>
      </Box>
    </Drawer>
  )
}

export default Sidebar
