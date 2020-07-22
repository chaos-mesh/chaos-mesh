import { Box, Button, Grow, IconButton, Modal, Paper, Typography } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { setAlert, setAlertOpen } from 'slices/globalStatus'
import { useHistory, useParams } from 'react-router-dom'

import CloseIcon from '@material-ui/icons/Close'
import ConfirmDialog from 'components/ConfirmDialog'
import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Event } from 'api/events.type'
import EventDetail from 'components/EventDetail'
import EventsTable from 'components/EventsTable'
import { Experiment } from 'components/NewExperiment/types'
import JSONEditor from 'components/JSONEditor'
import Loading from 'components/Loading'
import NoteOutlinedIcon from '@material-ui/icons/NoteOutlined'
import PaperTop from 'components/PaperTop'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import { StateOfExperimentsEnum } from 'api/experiments.type'
import api from 'api'
import genEventsChart from 'lib/d3/eventsChart'
import { getStateofExperiments } from 'slices/experiments'
import { toTitleCase } from 'lib/utils'
import useErrorButtonStyles from 'lib/styles/errorButton'
import { usePrevious } from 'lib/hooks'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    height100: {
      [theme.breakpoints.up('md')]: {
        height: '100%',
      },
    },
    timelinePaper: {
      marginBottom: theme.spacing(3),
    },
    eventsChart: {
      height: 300,
      margin: theme.spacing(3),
    },
    eventDetailPaper: {
      position: 'absolute',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      overflow: 'scroll',
    },
    configPaper: {
      position: 'absolute',
      top: '50%',
      left: '50%',
      width: '50vw',
      height: '70vh',
      transform: 'translate(-50%, -50%)',
      [theme.breakpoints.down('xs')]: {
        width: '90vw',
      },
    },
    updateExperimentButton: {
      position: 'absolute',
      top: theme.spacing(1.5),
      right: theme.spacing(1.5),
      padding: '1px 9px',
      color: '#fff',
      borderColor: '#fff',
      fontSize: '0.75rem',
    },
  })
)

export default function ExperimentDetail() {
  const classes = useStyles()
  const errorButton = useErrorButtonStyles()

  const history = useHistory()
  const { search } = history.location
  const searchParams = new URLSearchParams(search)
  const name = searchParams.get('name')
  const eventID = searchParams.get('event')
  const status = searchParams.get('status')
  const { uuid } = useParams()

  const dispatch = useStoreDispatch()

  const chartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detail, setDetail] = useState<Experiment | null>(null)
  const [events, setEvents] = useState<Event[] | null>(null)
  const prevEvents = usePrevious(events)
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [eventDetailOpen, setEventDetailOpen] = useState(false)
  const [configOpen, setConfigOpen] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogInfo, setDialogInfo] = useState({
    title: '',
    description: '',
    action: 'delete',
  })
  const [paused, setPaused] = useState(status === 'Paused' ? true : false)

  const fetchExperimentDetail = () => {
    setLoading(true)

    api.experiments
      .detail(uuid)
      .then(({ data }) => setDetail(data))
      .catch(console.log)
  }

  const fetchEvents = () =>
    api.events
      .events()
      .then(({ data }) => setEvents(data.filter((d) => d.ExperimentID === uuid)))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })

  const onSelectEvent = (e: Event) => {
    setDetailLoading(true)
    setSelectedEvent(e)
    setEventDetailOpen(true)
    setTimeout(() => setDetailLoading(false), 500)
  }

  const closeEventDetail = () => {
    setEventDetailOpen(false)
    searchParams.set('event', 'null')
    history.replace(window.location.pathname + '?' + searchParams.toString())
  }

  useEffect(() => {
    fetchExperimentDetail()
    fetchEvents()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (prevEvents !== events && events) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
        selectEvent: onSelectEvent,
      })
    }

    if (eventID !== null && eventID !== 'null' && events) {
      onSelectEvent(events.filter((e) => e.ID === parseInt(eventID))[0])
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events, eventID])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleAction = (action: string) => () => {
    switch (action) {
      case 'delete':
        setDialogInfo({
          title: `Delete ${name}?`,
          description: "Once you delete this experiment, it can't be recovered.",
          action: 'delete',
        })

        break
      case 'pause':
        setDialogInfo({
          title: `Pause ${name}?`,
          description: 'You can restart the experiment in the same position.',
          action: 'pause',
        })

        break
      case 'start':
        setDialogInfo({
          title: `Start ${name}?`,
          description: 'The operation will take effect immediately.',
          action: 'start',
        })

        break
      default:
        break
    }

    setDialogOpen(true)
  }

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.experiments.deleteExperiment

        break
      case 'pause':
        actionFunc = api.experiments.pauseExperiment

        break
      case 'start':
        actionFunc = api.experiments.startExperiment

        break
      default:
        actionFunc = null
    }

    if (actionFunc === null) {
      return
    }

    setDialogOpen(false)

    actionFunc(uuid)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: `${toTitleCase(action)}${action === 'start' ? 'ed' : 'd'} successfully!`,
          })
        )
        dispatch(setAlertOpen(true))
        dispatch(getStateofExperiments())

        if (action === 'delete') {
          history.push('/experiments')
        }

        if (action === 'pause') {
          setPaused(true)
          searchParams.set('status', StateOfExperimentsEnum.Paused)
        }

        if (action === 'start') {
          setPaused(false)
          searchParams.set('status', StateOfExperimentsEnum.Running)
        }

        if (action === 'pause' || action === 'start') {
          history.replace(window.location.pathname + '?' + searchParams.toString())
        }
      })
      .catch(console.log)
  }

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Box display="flex" flexDirection="column" height="100%">
          <Box display="flex" justifyContent="space-between" mb={3}>
            <Box display="flex">
              <Box mr={3}>
                <Button
                  className={errorButton.root}
                  variant="outlined"
                  size="small"
                  startIcon={<DeleteOutlineIcon />}
                  onClick={handleAction('delete')}
                >
                  Delete
                </Button>
              </Box>
              {paused ? (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PlayCircleOutlineIcon />}
                  onClick={handleAction('start')}
                >
                  Start
                </Button>
              ) : (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PauseCircleOutlineIcon />}
                  onClick={handleAction('pause')}
                >
                  Pause
                </Button>
              )}
            </Box>
            <Button
              variant="outlined"
              size="small"
              color="primary"
              startIcon={<NoteOutlinedIcon />}
              onClick={onModalOpen}
            >
              Configuration
            </Button>
          </Box>
          <Paper className={classes.timelinePaper} variant="outlined">
            <PaperTop title="Timeline" />
            <div ref={chartRef} className={classes.eventsChart} />
          </Paper>
          <Box className={classes.height100} position="relative">
            <Paper className={classes.height100} variant="outlined">
              {events && <EventsTable events={events} detailed noExperiment />}
            </Paper>
            {eventDetailOpen && (
              <Paper
                variant="outlined"
                className={classes.eventDetailPaper}
                style={{
                  zIndex: 3, // .MuiTableCell-stickyHeader z-index: 2
                }}
              >
                <PaperTop title="Event Detail">
                  <IconButton color="primary" onClick={closeEventDetail}>
                    <CloseIcon />
                  </IconButton>
                </PaperTop>
                {selectedEvent && !detailLoading ? <EventDetail event={selectedEvent} /> : <Loading />}
              </Paper>
            )}
          </Box>
        </Box>
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <Paper className={classes.configPaper}>
          <JSONEditor json={detail} />
          <Button className={classes.updateExperimentButton} variant="outlined" size="small">
            Update
          </Button>
        </Paper>
      </Modal>

      {(!name || !uuid) && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <Box mb={3}>
            <ErrorOutlineIcon fontSize="large" />
          </Box>
          <Typography variant="h6" align="center">
            Please check the URL params and queries to provide the correct params.
          </Typography>
        </Box>
      )}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={dialogInfo.title}
        description={dialogInfo.description}
        handleConfirm={handleExperiment(dialogInfo.action)}
      />

      {loading && <Loading />}
    </>
  )
}
