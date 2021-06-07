import { Box, Button, Grid, Grow } from '@material-ui/core'
import { setAlert, setConfirm } from 'slices/globalStatus'
import { useEffect, useState } from 'react'
import { useHistory, useParams } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import { Ace } from 'ace-builds'
import Alert from '@material-ui/lab/Alert'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import CloudDownloadOutlinedIcon from '@material-ui/icons/CloudDownloadOutlined'
import { Event } from 'api/events.type'
import EventsTimeline from 'components/EventsTimeline'
import { ExperimentSingle } from 'api/experiments.type'
import Loading from 'components-mui/Loading'
import ObjectConfiguration from 'components/ObjectConfiguration'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import PublishIcon from '@material-ui/icons/Publish'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import fileDownload from 'js-file-download'
import loadable from '@loadable/component'
import { makeStyles } from '@material-ui/styles'
import { useIntl } from 'react-intl'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  yamlEditorWrapper: {
    flex: 1,
    width: `calc(100% + ${theme.spacing(9)})`,
    marginLeft: theme.spacing(-4.5),
  },
}))

export default function Single() {
  const classes = useStyles()

  const history = useHistory()
  const { uuid } = useParams<{ uuid: uuid }>()

  const intl = useIntl()

  const { theme } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [single, setSingle] = useState<ExperimentSingle>()
  const [events, setEvents] = useState<Event[]>([])
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()

  const fetchExperiment = () => {
    api.experiments
      .single(uuid)
      .then(({ data }) => setSingle(data))
      .catch(console.error)
  }

  useEffect(() => {
    fetchExperiment()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const fetchEvents = () => {
      api.events
        .events({ object_id: uuid, limit: 999 })
        .then(({ data }) => setEvents(data))
        .catch(console.error)
        .finally(() => {
          setLoading(false)
        })
    }

    if (single) {
      fetchEvents()
    }
  }, [uuid, single])

  const handleSelect = (action: string) => () => {
    switch (action) {
      case 'archive':
        dispatch(
          setConfirm({
            title: `${intl.formatMessage({ id: 'archives.single' })} ${single!.name}`,
            description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
            handle: handleAction('archive'),
          })
        )

        break
      case 'pause':
        dispatch(
          setConfirm({
            title: `${intl.formatMessage({ id: 'common.pause' })} ${single!.name}`,
            description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
            handle: handleAction('pause'),
          })
        )

        break
      case 'start':
        dispatch(
          setConfirm({
            title: `${intl.formatMessage({ id: 'common.start' })} ${single!.name}`,
            description: intl.formatMessage({ id: 'experiments.startDesc' }),
            handle: handleAction('start'),
          })
        )

        break
    }
  }

  const handleAction = (action: string) => () => {
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

    if (actionFunc) {
      actionFunc(uuid)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: intl.formatMessage({ id: `confirm.${action}Successfully` }),
            })
          )

          if (action === 'archive') {
            history.push('/experiments')
          }

          if (action === 'pause' || action === 'start') {
            setTimeout(fetchExperiment, 300)
          }
        })
        .catch(console.error)
    }
  }

  const handleDownloadExperiment = () => fileDownload(yaml.dump(single!.kube_object), `${single!.name}.yaml`)

  const handleUpdateExperiment = () => {
    const data = yaml.load(yamlEditor!.getValue())

    api.experiments
      .update(data)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'confirm.updateSuccessfully' }),
          })
        )

        fetchExperiment()
      })
      .catch(console.error)
  }

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div>
          <Space spacing={6}>
            <Space direction="row">
              <Button
                variant="outlined"
                size="small"
                startIcon={<ArchiveOutlinedIcon />}
                onClick={handleSelect('archive')}
              >
                {T('archives.single')}
              </Button>
              {single?.status === 'paused' ? (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PlayCircleOutlineIcon />}
                  onClick={handleSelect('start')}
                >
                  {T('common.start')}
                </Button>
              ) : (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PauseCircleOutlineIcon />}
                  onClick={handleSelect('pause')}
                >
                  {T('common.pause')}
                </Button>
              )}
            </Space>

            {single?.failed_message && (
              <Alert severity="error">
                An error occurred: <b>{single.failed_message}</b>
              </Alert>
            )}

            <Paper>
              <PaperTop title={T('common.configuration')}></PaperTop>
              {single && <ObjectConfiguration config={single} />}
            </Paper>

            <Grid container>
              <Grid item xs={12} lg={6} sx={{ pr: 3 }}>
                <Paper sx={{ height: 600, overflow: 'scroll' }}>
                  <PaperTop title={T('events.title')} />
                  <Box flex={1}>
                    <EventsTimeline events={events} />
                  </Box>
                </Paper>
              </Grid>
              <Grid item xs={12} lg={6} sx={{ pl: 3 }}>
                <Paper sx={{ height: 600, pb: 0 }}>
                  {single && (
                    <Box display="flex" flexDirection="column" height="100%">
                      <PaperTop title={T('common.definition')}>
                        <Space direction="row">
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
                            color="primary"
                            size="small"
                            startIcon={<PublishIcon />}
                            onClick={handleUpdateExperiment}
                          >
                            {T('common.update')}
                          </Button>
                        </Space>
                      </PaperTop>
                      <Box className={classes.yamlEditorWrapper}>
                        <YAMLEditor theme={theme} data={yaml.dump(single.kube_object)} mountEditor={setYAMLEditor} />
                      </Box>
                    </Box>
                  )}
                </Paper>
              </Grid>
            </Grid>
          </Space>
        </div>
      </Grow>

      {loading && <Loading />}
    </>
  )
}
