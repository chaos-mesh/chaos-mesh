import { Box, Button, Card, CardActions, CardContent, IconButton, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentEventsPreview from 'components/ExperimentEventsPreview'
import { Link } from 'react-router-dom'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import React from 'react'
import day from 'lib/dayjs'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    card: {
      padding: theme.spacing(3),
    },
    actions: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
    },
    detailButton: {
      marginRight: theme.spacing(5),
    },
  })
)

interface ExperimentCardProps {
  experiment: Experiment
  handleSelect: (info: {
    namespace: string
    name: string
    kind: string
    title: string
    description: string
    action: string
  }) => void
  handleDialogOpen: (open: boolean) => void
}

const ExperimentCard: React.FC<ExperimentCardProps> = ({ experiment: e, handleSelect, handleDialogOpen }) => {
  const classes = useStyles()

  const handleDelete = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Delete ${e.Name}?`,
      description: "Once you delete this experiment, it can't be recovered.",
      action: 'delete',
    })
  }

  const handlePause = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Pause ${e.Name}?`,
      description: 'You can restart the experiment in the same position.',
      action: 'pause',
    })
  }

  const handleStart = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Start ${e.Name}?`,
      description: 'The operation will take effect immediately.',
      action: 'start',
    })
  }

  return (
    <Card className={classes.card} variant="outlined">
      <CardContent>
        <Box className={classes.actions}>
          <Typography variant="subtitle2">Created at {day(e.created).fromNow()}</Typography>
          <Box>
            {e.status.toLowerCase() === 'paused' ? (
              <IconButton color="primary" aria-label="Pause experiment" component="span" onClick={handleStart(e)}>
                <PlayCircleOutlineIcon />
              </IconButton>
            ) : (
              <IconButton color="primary" aria-label="Pause experiment" component="span" onClick={handlePause(e)}>
                <PauseCircleOutlineIcon />
              </IconButton>
            )}
            <IconButton color="primary" aria-label="Delete experiment" component="span" onClick={handleDelete(e)}>
              <DeleteOutlineIcon />
            </IconButton>
          </Box>
        </Box>
        <Typography variant="h6">{e.Name}</Typography>
        <Typography variant="subtitle1" color="textSecondary">
          {e.Kind}
        </Typography>
      </CardContent>
      <CardActions className={classes.actions}>
        <Box>
          <ExperimentEventsPreview events={e.events} />
        </Box>
        <Button
          className={classes.detailButton}
          component={Link}
          to={`/experiments/${e.Name}?namespace=${e.Namespace}&kind=${e.Kind}`}
          variant="outlined"
          size="small"
        >
          Detail
        </Button>
      </CardActions>
    </Card>
  )
}

export default ExperimentCard
