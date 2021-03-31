import React, { useImperativeHandle, useState } from 'react'
import { Step, StepLabel, Stepper } from '@material-ui/core'

import AdjustIcon from '@material-ui/icons/Adjust'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  stepper: {
    justifyContent: 'end',
    flex: 0.75,
    padding: 0,
  },
  step: {
    paddingLeft: theme.spacing(6),
    paddingRight: theme.spacing(6),
  },
  asButton: {
    cursor: 'pointer',
  },
  finish: {
    color: theme.palette.success.main,
  },
  primary: {
    color: theme.palette.primary.main,
  },
  disabled: {
    color: theme.palette.action.disabled,
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
    <Stepper className={classes.stepper}>
      {Array(count)
        .fill(0)
        .map((_, index) => (
          <Step key={index}>
            <StepLabel
              icon={
                <AdjustIcon
                  className={clsx(
                    classes.asButton,
                    current > index ? classes.finish : current === index ? classes.primary : classes.disabled
                  )}
                />
              }
              onClick={handleSetCurrent(index)}
            />
          </Step>
        ))}
    </Stepper>
  )
}

export default React.forwardRef(MultiNode)
