import { Box, Button } from '@material-ui/core'
import React, { useState } from 'react'

import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowDropUpIcon from '@material-ui/icons/ArrowDropUp'

interface AdvancedOptionsProps {
  isOpen?: boolean
  beforeOpen?: () => void
  afterClose?: () => void
  title?: string
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
    <Box display="flex" flexDirection="column">
      <Box my={3} textAlign="right">
        <Button color="primary" startIcon={open ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />} onClick={setOpen}>
          {title ? title : 'Advanced Options'}
        </Button>
      </Box>
      <Box display={open ? 'unset' : 'none'}>{children}</Box>
    </Box>
  )
}

export default AdvancedOptions
