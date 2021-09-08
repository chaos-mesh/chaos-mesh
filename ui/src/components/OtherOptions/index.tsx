import { Box, Button } from '@material-ui/core'

import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowDropUpIcon from '@material-ui/icons/ArrowDropUp'
import Space from 'components-mui/Space'
import T from 'components/T'
import { useState } from 'react'

interface OtherOptionsProps {
  isOpen?: boolean
  beforeOpen?: () => void
  afterClose?: () => void
  title?: string | JSX.Element
  disabled?: boolean
}

const OtherOptions: React.FC<OtherOptionsProps> = ({
  isOpen = false,
  beforeOpen,
  afterClose,
  title,
  disabled,
  children,
}) => {
  const [open, _setOpen] = useState(isOpen)

  const setOpen = () => {
    if (open) {
      _setOpen(false)
      typeof afterClose === 'function' && afterClose()
    } else {
      typeof beforeOpen === 'function' && beforeOpen()
      _setOpen(true)
    }
  }

  return (
    <Space>
      <Box textAlign="right">
        <Button
          color="primary"
          startIcon={open ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />}
          onClick={setOpen}
          disabled={disabled}
        >
          {title ? title : T('common.otherOptions')}
        </Button>
      </Box>
      <Space sx={{ display: open ? 'unset' : 'none' }}>{children}</Space>
    </Space>
  )
}

export default OtherOptions
