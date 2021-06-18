import { Box, Button } from '@material-ui/core'

import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowDropUpIcon from '@material-ui/icons/ArrowDropUp'
import Space from 'components-mui/Space'
import T from 'components/T'
import { useState } from 'react'

interface AdvancedOptionsProps {
  isOpen?: boolean
  beforeOpen?: () => void
  afterClose?: () => void
  title?: string | JSX.Element
}

const AdvancedOptions: React.FC<AdvancedOptionsProps> = ({
  isOpen = false,
  beforeOpen,
  afterClose,
  title,
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
    <>
      <Box mb={3} textAlign="right">
        <Button color="primary" startIcon={open ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />} onClick={setOpen}>
          {title ? title : T('common.advancedOptions')}
        </Button>
      </Box>
      <Space sx={{ display: open ? 'unset' : 'none' }}>{children}</Space>
    </>
  )
}

export default AdvancedOptions
