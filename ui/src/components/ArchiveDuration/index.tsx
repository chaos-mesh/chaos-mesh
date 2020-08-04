import {
  Box,
  Paper,
  Step,
  StepConnector,
  StepIconProps,
  StepLabel,
  Stepper,
  Theme,
  Typography,
  createStyles,
  makeStyles,
  withStyles,
} from '@material-ui/core'

import Check from '@material-ui/icons/Check'
import React from 'react'
import clsx from 'clsx'
import { format } from 'lib/dayjs'

const QontoConnector = withStyles({
  alternativeLabel: {
    top: 10,
    left: 'calc(-50% + 16px)',
    right: 'calc(50% + 16px)',
  },
  active: {
    '& $line': {
      borderColor: '#172d72',
    },
  },
  completed: {
    '& $line': {
      borderColor: '#172d72',
    },
  },
  line: {
    borderColor: '#eaeaf0',
    borderTopWidth: 3,
  },
})(StepConnector)

const useQontoStepIconStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      color: '#eaeaf0',
      display: 'flex',
      height: 22,
      alignItems: 'center',
    },
    active: {
      color: theme.palette.primary.main,
    },
    circle: {
      width: 8,
      height: 8,
      borderRadius: '50%',
      backgroundColor: 'currentColor',
    },
    completed: {
      color: theme.palette.primary.main,
      fontSize: theme.typography.h6.fontSize,
    },
  })
)

function QontoStepIcon(props: StepIconProps) {
  const classes = useQontoStepIconStyles()
  const { active, completed } = props

  return (
    <div
      className={clsx(classes.root, {
        [classes.active]: active,
      })}
    >
      {completed ? <Check className={classes.completed} /> : <div className={classes.circle} />}
    </div>
  )
}

const ArchiveDuration: React.FC<{ start: string; end: string }> = ({ start, end }) => {
  const steps = [format(start), format(end)]

  return (
    <Paper variant="outlined">
      <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100px" my={6}>
        <Typography variant="overline">Duration</Typography>
        <Box width="100%" mt={6}>
          <Stepper alternativeLabel activeStep={1} connector={<QontoConnector />} style={{ padding: 0 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel StepIconComponent={QontoStepIcon}>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>
        </Box>
      </Box>
    </Paper>
  )
}

export default ArchiveDuration
