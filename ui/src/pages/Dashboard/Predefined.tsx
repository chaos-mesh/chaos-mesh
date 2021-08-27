import { Box, Button, Card, Modal, Typography } from '@material-ui/core'
import { PreDefinedValue, getDB } from 'lib/idb'
import { parseSubmit, yamlToExperiment } from 'lib/formikhelpers'
import { setAlert, setConfirm } from 'slices/globalStatus'
import { useEffect, useRef, useState } from 'react'

import { Ace } from 'ace-builds'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import YAML from 'components/YAML'
import api from 'api'
import clsx from 'clsx'
import { iconByKind } from 'lib/byKind'
import loadable from '@loadable/component'
import { makeStyles } from '@material-ui/styles'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  card: {
    flex: '0 0 240px',
    cursor: 'pointer',
    '&:hover': {
      background: theme.palette.action.hover,
    },
  },
  addCard: {
    width: 210,
  },
  editorPaperWrapper: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    width: '75vw',
    height: '90vh',
    padding: 0,
    transform: 'translate(-50%, -50%)',
    [theme.breakpoints.down('lg')]: {
      width: '90vw',
    },
  },
}))

const Predefined = () => {
  const classes = useStyles()

  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const idb = useRef(getDB())

  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()
  const [editorOpen, seteditorOpen] = useState(false)
  const [experiment, setExperiment] = useState<PreDefinedValue>()
  const [experiments, setExperiments] = useState<PreDefinedValue[]>([])

  async function getExperiments() {
    setExperiments(await (await idb.current).getAll('predefined'))
  }

  useEffect(() => {
    getExperiments()
  }, [])

  const saveExperiment = async (_y: any) => {
    const db = await idb.current

    const y: any = yaml.load(_y)

    await db.put('predefined', {
      name: y.metadata.name,
      kind: y.kind,
      yaml: y,
    })

    getExperiments()
  }

  const onModalOpen = (exp: PreDefinedValue) => () => {
    seteditorOpen(true)
    setExperiment(exp)
  }
  const onModalClose = () => seteditorOpen(false)

  const handleApplyExperiment = () => {
    const { basic, target } = yamlToExperiment(yaml.load(yamlEditor!.getValue()))
    const parsedValues = parseSubmit({
      ...basic,
      target,
    })

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug parsedValues:', parsedValues)
    }

    api.experiments
      .newExperiment(parsedValues)
      .then(() => {
        seteditorOpen(false)
        dispatch(
          setAlert({
            type: 'success',
            message: T('confirm.success.create', intl),
          })
        )
      })
      .catch(console.error)
  }

  const handleDeleteConfirm = () => {
    dispatch(
      setConfirm({
        title: `${T('common.delete', intl)} ${experiment!.name}`,
        description: T('common.deleteDesc', intl),
        handle: handleDeleteExperiment,
      })
    )
  }

  const handleDeleteExperiment = async () => {
    const db = await idb.current

    await db.delete('predefined', experiment!.name)

    getExperiments()
    seteditorOpen(false)
    dispatch(
      setAlert({
        type: 'success',
        message: T('confirm.success.delete', intl),
      })
    )
  }

  return (
    <>
      <Space direction="row" sx={{ height: 88, overflowX: 'scroll' }}>
        <YAML
          callback={saveExperiment}
          buttonProps={{ className: clsx(classes.card, classes.addCard, 'tutorial-predefined') }}
        />
        {experiments.map((d) => (
          <Card key={d.name} className={classes.card} variant="outlined" onClick={onModalOpen(d)}>
            <Box display="flex" justifyContent="center" alignItems="center" height="100%">
              <Box display="flex" justifyContent="center" flex={1}>
                {iconByKind(d.kind)}
              </Box>
              <Box display="flex" justifyContent="center" flex={2} px={1.5}>
                <Typography>{d.name}</Typography>
              </Box>
            </Box>
          </Card>
        ))}
      </Space>
      <Modal open={editorOpen} onClose={onModalClose}>
        <div>
          <Paper className={classes.editorPaperWrapper}>
            {experiment && (
              <Box display="flex" flexDirection="column" height="100%">
                <Box px={3} pt={3}>
                  <PaperTop title={experiment.name}>
                    <Space direction="row">
                      <Button color="secondary" size="small" onClick={handleDeleteConfirm}>
                        {T('common.delete')}
                      </Button>
                      <Button variant="contained" color="primary" size="small" onClick={handleApplyExperiment}>
                        {T('common.submit')}
                      </Button>
                    </Space>
                  </PaperTop>
                </Box>
                <Box flex={1}>
                  <YAMLEditor data={yaml.dump(experiment.yaml)} mountEditor={setYAMLEditor} />
                </Box>
              </Box>
            )}
          </Paper>
        </div>
      </Modal>
    </>
  )
}

export default Predefined
