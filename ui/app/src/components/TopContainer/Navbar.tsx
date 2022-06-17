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
import MenuIcon from '@mui/icons-material/Menu'
import MenuOpenIcon from '@mui/icons-material/MenuOpen'
import { AppBar, Box, IconButton, Toolbar } from '@mui/material'

import Space from '@ui/mui-extends/esm/Space'

import Search from 'components/Search'

import Namespace from './Namespace'

interface HeaderProps {
  openDrawer: boolean
  handleDrawerToggle: () => void
}

const Navbar: React.FC<HeaderProps> = ({ openDrawer, handleDrawerToggle }) => (
  <AppBar position="static" color="transparent" elevation={0} sx={{ pl: 5, pr: 8 }}>
    <Toolbar disableGutters>
      <Box display="flex" justifyContent="space-between" alignItems="center" width="100%">
        <IconButton size="large" onClick={handleDrawerToggle} sx={{ color: 'onSurfaceVariant.main' }}>
          {openDrawer ? <MenuOpenIcon fontSize="medium" /> : <MenuIcon fontSize="medium" />}
        </IconButton>
        <Space direction="row">
          <Search />
          <Namespace />
        </Space>
      </Box>
    </Toolbar>
  </AppBar>
)

export default Navbar
