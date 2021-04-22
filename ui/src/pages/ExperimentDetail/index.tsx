import { Box, Button, Grid, Grow, Modal } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import EventsTable, { EventsTableHandles } from 'components/EventsTable'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import { useEffect, useRef, useState } from 'react'
import { useHistory, useParams } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import { Ace } from 'ace-builds'
import Alert from '@material-ui/lab/Alert'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import CloudDownloadOutlinedIcon from '@material-ui/icons/CloudDownloadOutlined'
import { Event } from 'api/events.type'
import ExperimentConfiguration from 'components/ExperimentConfiguration'
import { ExperimentDetail as ExperimentDetailType } from 'api/experiments.type'
import Loading from 'components-mui/Loading'
import NoteOutlinedIcon from '@material-ui/icons/NoteOutlined'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import fileDownload from 'js-file-download'
import genEventsChart from 'lib/d3/eventsChart'
import loadable from '@loadable/component'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { usePrevious } from 'lib/hooks'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) =>
  createStyles({
    eventsChart: {
      height: 150,
      margin: theme.spacing(3),
    },
    eventDetailPaper: {
      position: 'absolute',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      overflowY: 'scroll',
    },
    configPaper: {
      position: 'absolute',
      top: '50%',
      left: '50%',
      width: '50vw',
      height: '90vh',
      transform: 'translate(-50%, -50%)',
      [theme.breakpoints.down('sm')]: {
        width: '90vw',
      },
    },
  })
)

const initialSelected = {
  title: '',
  description: '',
  action: '',
}

export default function ExperimentDetail() {
  const classes = useStyles()

  const intl = useIntl()

  const history = useHistory()
  const { uuid } = useParams<{ uuid: string }>()

  const { theme } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const chartRef = useRef<HTMLDivElement>(null)
  const eventsTableRef = useRef<EventsTableHandles>(null)
  const confirmRef = useRef<ConfirmDialogHandles>(null)

  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ExperimentDetailType>()
  const [events, setEvents] = useState<Event[]>()
  const prevEvents = usePrevious(events)
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()
  const [configOpen, setConfigOpen] = useState(false)
  const [selected, setSelected] = useState(initialSelected)

  const fetchExperimentDetail = () => {
    api.experiments
      .detail(uuid)
      .then(({ data }) => setDetail(data))
      .catch(console.error)
  }

  const fetchEvents = () =>
    api.events
      .events({ experimentName: detail!.yaml.metadata.name })
      .then(({ data }) => setEvents(data))
      .catch(console.error)
      .finally(() => {
        setLoading(false)
      })

  useEffect(() => {
    fetchExperimentDetail()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (detail) {
      fetchEvents()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail])

  useEffect(() => {
    if (prevEvents !== events && prevEvents?.length !== events?.length && events) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
        intl,
        theme,
        options: {
          enableLegends: false,
          onSelectEvent: eventsTableRef.current!.onSelectEvent,
        },
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleAction = (action: string) => () => {
    switch (action) {
      case 'archive':
        setSelected({
          title: `${intl.formatMessage({ id: 'archives.single' })} ${detail!.name}`,
          description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
          action: 'archive',
        })

        break
      case 'pause':
        setSelected({
          title: `${intl.formatMessage({ id: 'common.pause' })} ${detail!.name}`,
          description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
          action: 'pause',
        })

        break
      case 'start':
        setSelected({
          title: `${intl.formatMessage({ id: 'common.start' })} ${detail!.name}`,
          description: intl.formatMessage({ id: 'experiments.startDesc' }),
          action: 'start',
        })

        break
    }

    confirmRef.current!.setOpen(true)
  }

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'archive':
        actionFunc = api.experiments.del

        break
      case 'pause':
        actionFunc = api.experiments.pause

        break
      case 'start':
        actionFunc = api.experiments.start

        break
      default:
        actionFunc = null
    }

    confirmRef.current!.setOpen(false)

    if (actionFunc) {
      actionFunc(uuid)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: intl.formatMessage({ id: `common.${action}Successfully` }),
            })
          )

          if (action === 'archive') {
            history.push('/experiments')
          }

          if (action === 'pause' || action === 'start') {
            setTimeout(fetchExperimentDetail, 300)
          }
        })
        .catch(console.error)
    }
  }

  const handleDownloadExperiment = () => fileDownload(yaml.dump(detail!.yaml), `${detail!.name}.yaml`)

  const handleUpdateExperiment = () => {
    const data = yaml.load(yamlEditor!.getValue())

    api.experiments
      .update(data)
      .then(() => {
        setConfigOpen(false)
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
          })
        )

        fetchExperimentDetail()
      })
      .catch(console.error)
  }

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={6}>
          <Grid item xs={12}>
            <Space>
              <Button
                variant="outlined"
                size="small"
                startIcon={<ArchiveOutlinedIcon />}
                onClick={handleAction('archive')}
              >
                {T('archives.single')}
              </Button>
              {detail?.status === 'Paused' ? (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PlayCircleOutlineIcon />}
                  onClick={handleAction('start')}
                >
                  {T('common.start')}
                </Button>
              ) : (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PauseCircleOutlineIcon />}
                  onClick={handleAction('pause')}
                >
                  {T('common.pause')}
                </Button>
              )}
            </Space>
          </Grid>

          {detail?.failed_message && (
            <Grid item xs={12}>
              <Alert severity="error">
                An error occurred: <b>{detail.failed_message}</b>
              </Alert>
            </Grid>
          )}

          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('common.configuration')}>
                <Space>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<CloudDownloadOutlinedIcon />}
                    onClick={handleDownloadExperiment}
                  >
                    {T('common.download')}
                  </Button>
                  <Button
                    variant="outlined"
                    size="small"
                    color="primary"
                    startIcon={<NoteOutlinedIcon />}
                    onClick={onModalOpen}
                  >
                    {T('common.update')}
                  </Button>
                </Space>
              </PaperTop>
              {detail && <ExperimentConfiguration experimentDetail={detail} />}
            </Paper>
          </Grid>

          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('common.timeline')} />
              <div ref={chartRef} className={classes.eventsChart} />
            </Paper>
          </Grid>

          <Grid item xs={12}>
            {events && <EventsTable ref={eventsTableRef} events={events} />}
          </Grid>
        </Grid>
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <div>
          <Paper className={classes.configPaper} padding={0}>
            {detail && configOpen && (
              <Box display="flex" flexDirection="column" height="100%">
                <Box px={3} pt={3}>
                  <PaperTop title={detail.name}>
                    <Button variant="contained" color="primary" size="small" onClick={handleUpdateExperiment}>
                      {T('common.confirm')}
                    </Button>
                  </PaperTop>
                </Box>
                <Box flex={1}>
                  <YAMLEditor theme={theme} data={yaml.dump(detail.yaml)} mountEditor={setYAMLEditor} />
                </Box>
              </Box>
            )}
          </Paper>
        </div>
      </Modal>

      <ConfirmDialog
        ref={confirmRef}
        title={selected.title}
        description={selected.description}
        onConfirm={handleExperiment(selected.action)}
      />

      {loading && <Loading />}
    </>
  )
}
