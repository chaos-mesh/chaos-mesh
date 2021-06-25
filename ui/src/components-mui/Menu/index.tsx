import { IconButton, Menu as MUIMenu, MenuProps } from '@material-ui/core'

import MoreVertIcon from '@material-ui/icons/MoreVert'
import T from 'components/T'
import { useIntl } from 'react-intl'
import { useState } from 'react'

const Menu = ({ children, ...rest }: Partial<MenuProps>) => {
  const intl = useIntl()

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)

  const onClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const onClose = () => {
    setAnchorEl(null)
  }

  return (
    <div>
      <IconButton size="small" title={T('common.options', intl)} onClick={onClick}>
        <MoreVertIcon />
      </IconButton>
      <MUIMenu {...rest} anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={onClose}>
        {children}
      </MUIMenu>
    </div>
  )
}

export default Menu
