import { Step, StepLabel, Stepper } from '@material-ui/core'

import AdjustIcon from '@material-ui/icons/Adjust'
import React from 'react'
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

interface MultiNodeProps {
  count: number
  current: number
  setCurrent: React.Dispatch<React.SetStateAction<number>>
  setCurrentCallback?: (index: number) => void
}

const MultiNode: React.FC<MultiNodeProps> = ({ count, current, setCurrent, setCurrentCallback }) => {
  const classes = useStyles()

  const handleSetCurrent = (index: number) => () => {
    setCurrent(index)

    setCurrentCallback && setCurrentCallback(index)
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

export default MultiNode
