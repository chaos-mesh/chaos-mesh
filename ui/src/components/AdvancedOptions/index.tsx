import { Box, Button } from '@material-ui/core'
import React, { useState } from 'react'

import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowDropUpIcon from '@material-ui/icons/ArrowDropUp'

interface AdvancedOptionsProps {
  isOpen?: boolean
}

const AdvancedOptions: React.FC<AdvancedOptionsProps> = ({ isOpen = false, children }) => {
  const [open, setOpen] = useState(isOpen)

  return (
    <Box display="flex" flexDirection="column">
      <Box my={3} textAlign="right">
        <Button
          color="primary"
          startIcon={open ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />}
          onClick={() => setOpen(!open)}
        >
          Advanced Options
        </Button>
      </Box>
      <Box display={open ? 'unset' : 'none'}>{children}</Box>
    </Box>
  )
}

export default AdvancedOptions
