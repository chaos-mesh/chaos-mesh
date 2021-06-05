import { Avatar, Typography } from '@material-ui/core'
import React, { useImperativeHandle, useState } from 'react'

import Space from 'components-mui/Space'
import T from 'components/T'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles((theme) => ({
  avatar: {
    width: 20,
    height: 20,
    fontSize: '1rem',
    cursor: 'pointer',
  },
  finish: {
    background: theme.palette.success.main,
  },
  primary: {
    background: theme.palette.primary.main,
  },
  disabled: {
    background: theme.palette.action.disabled,
  },
}))

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
  const classes = useStyles()
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
    <Space alignItems="center">
      <Typography>{T(`newW.node.chooseChildren`)}</Typography>
      {Array(count)
        .fill(0)
        .map((_, index) => (
          <Avatar
            key={index}
            className={clsx(
              classes.avatar,
              current > index ? classes.finish : current === index ? classes.primary : classes.disabled
            )}
            onClick={handleSetCurrent(index)}
          >
            {index + 1}
          </Avatar>
        ))}
    </Space>
  )
}

export default React.forwardRef(MultiNode)
