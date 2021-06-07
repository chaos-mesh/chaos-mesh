import { Avatar, Typography } from '@material-ui/core'
import React, { useImperativeHandle, useState } from 'react'

import Space from 'components-mui/Space'
import T from 'components/T'

export interface MultiNodeHandles {
  current: number
  setCurrent: React.Dispatch<React.SetStateAction<number>>
}

interface MultiNodeProps {
  count: number
  setCurrentCallback?: (index: number) => boolean
}

const MultiNode: React.ForwardRefRenderFunction<MultiNodeHandles, MultiNodeProps> = (
  { count, setCurrentCallback },
  ref
) => {
  const [current, setCurrent] = useState(0)

  // Methods exposed to the parent
  useImperativeHandle(ref, () => ({
    current,
    setCurrent,
  }))

  const handleSetCurrent = (index: number) => () => {
    if (setCurrentCallback) {
      const result = setCurrentCallback(index)

      if (!result) {
        return
      }
    }

    setCurrent(index)
  }

  return (
    <Space direction="row" alignItems="center">
      <Typography>{T(`newW.node.chooseChildren`)}</Typography>
      {Array(count)
        .fill(0)
        .map((_, index) => (
          <Avatar
            key={index}
            sx={{
              width: 20,
              height: 20,
              fontSize: 16,
              bgcolor: current > index ? 'success.main' : current === index ? 'primary.main' : 'action.disabled',
              cursor: 'pointer',
            }}
            onClick={handleSetCurrent(index)}
          >
            {index + 1}
          </Avatar>
        ))}
    </Space>
  )
}

export default React.forwardRef(MultiNode)
