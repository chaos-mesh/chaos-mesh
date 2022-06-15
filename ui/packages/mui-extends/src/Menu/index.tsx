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
import MoreVertIcon from '@mui/icons-material/MoreVert'
import { IconButton, IconButtonProps, MenuProps, Menu as MuiMenu, SvgIconProps, styled } from '@mui/material'
import { useState } from 'react'

const StyledMenu = styled((props: MenuProps) => <MuiMenu elevation={0} {...props} />)(({ theme }) => ({
  '& .MuiPaper-root': {
    border: `1px solid ${theme.palette.divider}`,
  },
  '& .MuiMenu-list': {
    padding: '4px 0',
  },
  '& .MuiMenuItem-root': {
    padding: '3px 8px',
  },
}))

const Menu: React.FC<
  Omit<MenuProps, 'anchorEl' | 'open' | 'onClose'> & { IconButtonProps?: IconButtonProps; IconProps?: SvgIconProps }
> = ({ IconButtonProps, IconProps, children, ...rest }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)

  const onClick = (e: React.SyntheticEvent<HTMLButtonElement>) => {
    e.stopPropagation()

    setAnchorEl(e.currentTarget)
  }

  const onClose = (e: React.SyntheticEvent) => {
    e && e.stopPropagation() // Allow no event.

    setAnchorEl(null)
  }

  return (
    <>
      <IconButton {...IconButtonProps} onClick={onClick}>
        <MoreVertIcon {...IconProps} />
      </IconButton>
      <StyledMenu {...rest} anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={onClose}>
        {/* If `children` is a function, the return type must be an array of React elements. */}
        {typeof children === 'function' ? children({ onClose }) : children}
      </StyledMenu>
    </>
  )
}

export default Menu
