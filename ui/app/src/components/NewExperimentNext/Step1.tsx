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
import CheckIcon from '@mui/icons-material/Check'
import RadioButtonCheckedOutlinedIcon from '@mui/icons-material/RadioButtonCheckedOutlined'
import RadioButtonUncheckedOutlinedIcon from '@mui/icons-material/RadioButtonUncheckedOutlined'
import UndoIcon from '@mui/icons-material/Undo'
import { Box, Card, Divider, Typography } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { Stale } from 'api/queryUtils'
import clsx from 'clsx'
import { useGetCommonConfig } from 'openapi'
import React from 'react'

import Paper from '@ui/mui-extends/esm/Paper'

import { useStoreDispatch, useStoreSelector } from 'store'

import { Env, setEnv, setKindAction, setSpec, setStep1 } from 'slices/experiments'

import i18n from 'components/T'

import { iconByKind, transByKind } from 'lib/byKind'

import _typesData, { Definition, Kind, dataPhysic, schema } from './data/types'
import Kernel from './form/Kernel'
import Stress from './form/Stress'
import TargetGenerated from './form/TargetGenerated'

const useStyles = makeStyles((theme) => {
  const cardActive = {
    color: theme.palette.primary.main,
    borderColor: theme.palette.primary.main,
  }

  return {
    card: {
      cursor: 'pointer',
      marginTop: theme.spacing(3),
      marginRight: theme.spacing(3),
      '&:hover': cardActive,
    },
    cardActive,
    submit: {
      borderColor: theme.palette.success.main,
    },
    submitIcon: {
      color: theme.palette.success.main,
    },
    asButton: {
      cursor: 'pointer',
    },
  }
})

const submitDirectly = ['pod-failure']

interface TypeCardProp {
  name: Env
  handleSwitchEnv: (env: Env) => () => void
  env: Env
}

const TypeCard: React.FC<TypeCardProp> = ({ name, handleSwitchEnv, env }) => {
  const classes = useStyles()
  const title = name === 'k8s' ? 'k8s.title' : 'physics.single'
  return (
    <Card
      className={clsx(classes.card, env === name ? classes.cardActive : '')}
      variant="outlined"
      onClick={handleSwitchEnv(name)}
    >
      <Box display="flex" justifyContent="center" alignItems="center" width={225} height={75}>
        <Box display="flex" justifyContent="center" flex={1}>
          {iconByKind(name)}
        </Box>
        <Box flex={1.5} textAlign="center">
          <Typography variant="button">{i18n(title)}</Typography>
        </Box>
      </Box>
    </Card>
  )
}

const Step1 = () => {
  const classes = useStyles()

  const state = useStoreSelector((state) => state)
  const {
    env,
    kindAction: [kind, action],
    step1,
  } = state.experiments
  const dispatch = useStoreDispatch()

  const { data: config } = useGetCommonConfig({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  const typesData = env === 'k8s' ? _typesData : dataPhysic
  let typesDataEntries = Object.entries(typesData) as [Kind, Definition][]
  if (!config?.dns_server_create) {
    typesDataEntries = typesDataEntries.filter((d) => d[0] !== 'DNSChaos')
  }

  const handleSelectTarget = (key: Kind) => () => {
    dispatch(setKindAction([key, '']))
  }

  const handleSelectAction = (newAction: string) => () => {
    dispatch(setKindAction([kind, newAction]))

    if (submitDirectly.includes(newAction)) {
      handleSubmitStep1({ action: newAction })
    }
  }

  const handleSubmitStep1 = (values: Record<string, any>) => {
    const result = action
      ? {
          ...values,
          action,
        }
      : values

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep1:', result)
    }

    dispatch(setSpec(result))
    dispatch(setStep1(true))
  }

  const handleUndo = () => dispatch(setStep1(false))

  const handleSwitchEnv = (env: Env) => () => {
    dispatch(setKindAction(['', '']))
    dispatch(setEnv(env))
  }

  return (
    <Paper className={step1 ? classes.submit : ''}>
      <Box display="flex" justifyContent="space-between" mb={step1 ? 0 : 3}>
        <Box display="flex" alignItems="center">
          {step1 && (
            <Box display="flex" mr={3}>
              <CheckIcon className={classes.submitIcon} />
            </Box>
          )}
          <Typography>{i18n('newE.titleStep1')}</Typography>
        </Box>
        {step1 && <UndoIcon className={classes.asButton} onClick={handleUndo} />}
      </Box>
      <Box hidden={step1}>
        <Box display="flex">
          <TypeCard name="k8s" handleSwitchEnv={handleSwitchEnv} env={env} />
          <TypeCard name="physic" handleSwitchEnv={handleSwitchEnv} env={env} />
        </Box>
        <Divider sx={{ my: 6 }} />
      </Box>
      <Box hidden={step1}>
        <Box display="flex" flexWrap="wrap">
          {typesDataEntries.map(([key]) => (
            <Card
              key={key}
              className={clsx(classes.card, kind === key ? classes.cardActive : '')}
              variant="outlined"
              onClick={handleSelectTarget(key)}
            >
              <Box display="flex" justifyContent="center" alignItems="center" width={280} height={75}>
                <Box display="flex" justifyContent="center" flex={1}>
                  {iconByKind(key)}
                </Box>
                <Box flex={1.5} textAlign="center">
                  <Typography variant="button">{transByKind(key)}</Typography>
                </Box>
              </Box>
            </Card>
          ))}
        </Box>
        {kind && (
          <Box overflow="hidden">
            <Box mt={6} mb={3}>
              <Divider />
            </Box>
            {(typesData as any)[kind].categories ? (
              <Box display="flex" flexWrap="wrap">
                {(typesData as any)[kind].categories!.map((d: any) => (
                  <Card
                    key={d.key}
                    className={clsx(classes.card, action === d.key ? classes.cardActive : '')}
                    variant="outlined"
                    onClick={handleSelectAction(d.key)}
                  >
                    <Box display="flex" justifyContent="center" alignItems="center" width={210} height={50}>
                      <Box display="flex" justifyContent="center" alignItems="center" flex={0.5}>
                        {action === d.key ? <RadioButtonCheckedOutlinedIcon /> : <RadioButtonUncheckedOutlinedIcon />}
                      </Box>
                      <Box flex={1.5} textAlign="center">
                        <Typography variant="button">{d.name}</Typography>
                      </Box>
                    </Box>
                  </Card>
                ))}
              </Box>
            ) : kind === 'KernelChaos' ? (
              <Box mt={6}>
                <Kernel onSubmit={handleSubmitStep1} />
              </Box>
            ) : kind === 'TimeChaos' ? (
              <Box mt={6}>
                <TargetGenerated
                  env={env}
                  kind={kind}
                  data={(typesData as any)[kind].spec!}
                  validationSchema={env === 'k8s' ? schema.TimeChaos!.default : undefined}
                  onSubmit={handleSubmitStep1}
                />
              </Box>
            ) : kind === 'StressChaos' ? (
              <Box mt={6}>
                <Stress onSubmit={handleSubmitStep1} />
              </Box>
            ) : (kind as any) === 'ProcessChaos' ? (
              <Box mt={6}>
                <TargetGenerated
                  env={env}
                  kind={kind}
                  data={(typesData as any)[kind].spec!}
                  onSubmit={handleSubmitStep1}
                />
              </Box>
            ) : null}
          </Box>
        )}
        {action && !submitDirectly.includes(action) && (
          <>
            <Divider sx={{ my: 6 }} />
            <TargetGenerated
              // Force re-rendered after action changed
              key={kind + action}
              env={env}
              kind={kind}
              data={(typesData as any)[kind as Kind].categories!.filter(({ key }: any) => key === action)[0].spec}
              validationSchema={
                env === 'k8s' ? (schema[kind as Kind] ? schema[kind as Kind]![action] : undefined) : undefined
              }
              onSubmit={handleSubmitStep1}
            />
          </>
        )}
      </Box>
    </Paper>
  )
}

export default Step1
