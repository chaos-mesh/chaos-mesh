/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import loadable from '@loadable/component'
import { Box, Button, Card, Modal, Typography } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { Ace } from 'ace-builds'
import clsx from 'clsx'
import yaml from 'js-yaml'
import { postExperiments, postSchedules } from 'openapi'
import { useEffect, useRef, useState } from 'react'
import { useIntl } from 'react-intl'

import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch } from 'store'

import { setAlert, setConfirm } from 'slices/globalStatus'

import i18n from 'components/T'
import YAML from 'components/YAML'

import { iconByKind } from 'lib/byKind'
import { PreDefinedValue, getDB } from 'lib/idb'

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
    const exp: any = yaml.load(yamlEditor!.getValue())

    const isSchedule = exp['kind'] === 'Schedule'
    const action = isSchedule ? (schedule: any) => postSchedules(schedule) : (chaos: any) => postExperiments(chaos)

    action(exp)
      .then(() => {
        seteditorOpen(false)
        dispatch(
          setAlert({
            type: 'success',
            message: i18n('confirm.success.create', intl),
          })
        )
      })
      .catch(console.error)
  }

  const handleDeleteConfirm = () => {
    dispatch(
      setConfirm({
        title: `${i18n('common.delete', intl)} ${experiment!.name}`,
        description: i18n('common.deleteDesc', intl),
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
        message: i18n('confirm.success.delete', intl),
      })
    )
  }

  return (
    <>
      <Space direction="row" sx={{ height: 88, overflowX: 'scroll' }}>
        <YAML
          callback={saveExperiment}
          ButtonProps={{ className: clsx(classes.card, classes.addCard, 'tutorial-predefined') }}
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
                        {i18n('common.delete')}
                      </Button>
                      <Button variant="contained" color="primary" size="small" onClick={handleApplyExperiment}>
                        {i18n('common.submit')}
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
