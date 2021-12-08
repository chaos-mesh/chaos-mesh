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
