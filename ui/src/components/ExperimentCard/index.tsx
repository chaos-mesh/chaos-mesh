import { Box, Button, Card, CardActions, CardContent, IconButton, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import { Experiment } from 'api/experiments.type'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import React from 'react'
import day from 'lib/dayjs'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    card: {
      padding: theme.spacing(3),
    },
    detailButton: {
      marginLeft: 'auto',
      marginRight: theme.spacing(6),
    },
  })
)

interface ExperimentCardProps {
  experiment: Experiment
}

const ExperimentCard: React.FC<ExperimentCardProps> = ({ experiment: e }) => {
  const classes = useStyles()

  return (
    <Card className={classes.card} variant="outlined">
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="center">
          <Typography variant="subtitle2">Created at {day(e.created).fromNow()}</Typography>
          <Box>
            <IconButton color="primary" aria-label="Pause experiment" component="span">
              <PauseCircleOutlineIcon />
            </IconButton>
            <IconButton color="primary" aria-label="Pause experiment" component="span">
              <DeleteOutlineIcon />
            </IconButton>
          </Box>
        </Box>
        <Typography variant="h6">{e.Name}</Typography>
        <Typography variant="subtitle1" color="textSecondary">
          {e.Kind}
        </Typography>
      </CardContent>
      <CardActions>
        <Button className={classes.detailButton} size="small">
          Detail
        </Button>
      </CardActions>
    </Card>
  )
}

export default ExperimentCard
