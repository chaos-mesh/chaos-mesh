import React, { FC, useState } from 'react'
import { Link } from 'react-router-dom'

import {
  Button,
  Box,
  Card,
  CardContent,
  CardActions,
  Drawer,
  IconButton,
  LinearProgress,
  Typography,
} from '@material-ui/core'
import AddIcon from '@material-ui/icons/Add'
import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import { createStyles, Theme, makeStyles } from '@material-ui/core/styles'

import ContentContainer from '../../components/ContentContainer'
import InfoList from '../../components/InfoList'
import NewExperiment from '../../pages/Experiments/New'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    card: {
      marginBottom: theme.spacing(4),
      '&:last-child': {
        marginBottom: 0,
      },
    },
    cardContent: {
      display: 'flex',
      maxHeight: '20rem',
      padding: theme.spacing(4),
      overflow: 'auto',
    },
    cardActions: {
      justifyContent: 'space-between',
      padding: `0 ${theme.spacing(4)} ${theme.spacing(2)}`,
    },
    linearProgress: {
      flex: 1,
      minWidth: '6rem',
      height: '1rem',
      margin: `${theme.spacing(2)} ${theme.spacing(4)} ${theme.spacing(2)} 0`,
    },
  })
)

interface ExperimentProps {
  info: { [key: string]: string }
}
// TODO: ui polish
const ExperimentCard: FC<ExperimentProps> = ({ info, children }) => {
  const classes = useStyles()

  return (
    <Card className={classes.card}>
      <CardContent className={classes.cardContent}>
        <Box flexBasis="20rem">
          <InfoList info={info} />
        </Box>
        <Box flexGrow={1} px={4} py={2}>
          <Typography variant="h5">Events</Typography>
          <Box display="flex" flexWrap="wrap" mt={2}>
            {children}
          </Box>
        </Box>
      </CardContent>
      <CardActions className={classes.cardActions}>
        <Button size="large" color="primary" component={Link} to={`/experiments/${info.name}`}>
          Detail
        </Button>
        <Box>
          <IconButton aria-label="pause">
            <PauseCircleOutlineIcon fontSize="large" />
          </IconButton>
          <IconButton aria-label="delete">
            <DeleteOutlineIcon fontSize="large" />
          </IconButton>
        </Box>
      </CardActions>
    </Card>
  )
}

export default function Experiments() {
  const classes = useStyles()

  const [isOpen, setIsOpen] = useState(false)
  // TODO: interval to fetch experiment list
  const fakeList = [
    {
      name: 'tikv-failure',
      namespace: 'tidb-demo',
      kind: 'PodChaos',
      created: '1 day ago',
    },
    {
      name: 'tidb-failure',
      namespace: 'tidb-demo',
      kind: 'PodChaos',
      created: '2 days ago',
    },
  ]

  const toggleDrawer = (isOpen: boolean) => () => {
    setIsOpen(isOpen)
  }

  return (
    <>
      <Button variant="outlined" startIcon={<AddIcon />} onClick={toggleDrawer(true)}>
        New Experiment
      </Button>

      <ContentContainer>
        {fakeList.map((item: { [key: string]: string }, index) => {
          return (
            <ExperimentCard key={item.name + item.namespace} info={item}>
              {/* TODO: fake event progress, polish ui with tooltip as a component */}
              <LinearProgress variant="determinate" color="primary" value={20} className={classes.linearProgress} />
              <LinearProgress variant="determinate" color="secondary" value={50} className={classes.linearProgress} />
              {index === 0 && (
                <LinearProgress variant="determinate" color="primary" value={100} className={classes.linearProgress} />
              )}
            </ExperimentCard>
          )
        })}
      </ContentContainer>

      {/* New Experiment Stepper Drawer */}
      <Drawer anchor="right" open={isOpen} onClose={toggleDrawer(false)}>
        <NewExperiment />
      </Drawer>
    </>
  )
}
