import { Box, Button } from '@material-ui/core'
import React, { useState } from 'react'

import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowDropUpIcon from '@material-ui/icons/ArrowDropUp'

const AdvancedOptions: React.FC = ({ children }) => {
  const [open, setOpen] = useState(false)

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
