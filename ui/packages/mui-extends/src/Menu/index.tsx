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

import { IconButton, Menu as MUIMenu, MenuProps } from '@mui/material'

import MoreVertIcon from '@mui/icons-material/MoreVert'
import { useState } from 'react'

const Menu: React.FC<Omit<MenuProps, 'anchorEl' | 'open' | 'onClose'>> = ({ title, children, ...rest }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)

  const onClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const onClose = () => {
    setAnchorEl(null)
  }

  return (
    <div>
      <IconButton size="small" onClick={onClick}>
        <MoreVertIcon />
      </IconButton>
      <MUIMenu {...rest} anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={onClose}>
        {children}
      </MUIMenu>
    </div>
  )
}

export default Menu
