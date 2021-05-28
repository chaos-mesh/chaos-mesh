import { IconButton, Menu as MUIMenu, MenuProps } from '@material-ui/core'

import MoreVertIcon from '@material-ui/icons/MoreVert'
import { useState } from 'react'

const Menu = ({ children, ...rest }: Partial<MenuProps>) => {
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
