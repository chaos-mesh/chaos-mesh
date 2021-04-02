import { Box, Card, Divider, GridList, GridListTile, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import { iconByKind, transByKind } from 'lib/byKind'
import { setKindAction as setKindActionToStore, setStep1, setTarget as setTargetToStore } from 'slices/experiments'
import targetData, { Kind, Target, schema } from './data/target'
import { useEffect, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import CheckIcon from '@material-ui/icons/Check'
import Kernel from './form/Kernel'
import Paper from 'components-mui/Paper'
import RadioButtonCheckedOutlinedIcon from '@material-ui/icons/RadioButtonCheckedOutlined'
import RadioButtonUncheckedOutlinedIcon from '@material-ui/icons/RadioButtonUncheckedOutlined'
import Stress from './form/Stress'
import T from 'components/T'
import TargetGenerated from './form/TargetGenerated'
import UndoIcon from '@material-ui/icons/Undo'
import _snakecase from 'lodash.snakecase'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => {
  const cardActive = {
    color: theme.palette.primary.main,
    borderColor: theme.palette.primary.main,
  }

  return {
    gridList: {
      flexWrap: 'nowrap',
      // Promote the list into his own layer on Chrome. This cost memory but helps keeping high FPS.
      transform: 'translateZ(0)',
    },
    card: {
      cursor: 'pointer',
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

const Step1 = () => {
  const theme = useTheme()
  const isDesktopScreen = useMediaQuery(theme.breakpoints.down('md'))
  const classes = useStyles()

  const state = useStoreSelector((state) => state)
  const { dnsServerCreate } = state.globalStatus
  let targetDataEntries = Object.entries(targetData) as [Kind, Target][]
  if (!dnsServerCreate) {
    targetDataEntries = targetDataEntries.filter((d) => d[0] !== 'DNSChaos')
  }
  const {
    kindAction: [_kind, _action],
    step1,
  } = state.experiments
  const dispatch = useStoreDispatch()

  const [kindAction, setKindAction] = useState<[Kind | '', string]>([_kind, _action])
  const [kind, action] = kindAction

  useEffect(() => {
    setKindAction([_kind, _action])
  }, [_kind, _action])

  const handleSelectTarget = (key: Kind) => () => setKindAction([key, ''])

  const handleSelectAction = (newAction: string) => () => {
    dispatch(setKindActionToStore([kind, newAction]))

    if (submitDirectly.includes(newAction)) {
      handleSubmitStep1({ action: newAction })
    }
  }

  const handleSubmitStep1 = (values: Record<string, any>) => {
    const result = {
      kind,
      [_snakecase(kind)]:
        submitDirectly.includes(values.action) && Object.keys(values).length === 1
          ? values
          : action
          ? {
              ...values,
              action,
            }
          : values,
    }

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep1', result)
    }

    dispatch(setTargetToStore(result))
    dispatch(setStep1(true))
  }

  const handleUndo = () => dispatch(setStep1(false))

  return (
    <Paper className={step1 ? classes.submit : ''}>
      <Box display="flex" justifyContent="space-between" mb={step1 ? 0 : 6}>
        <Box display="flex" alignItems="center">
          {step1 && (
            <Box display="flex" mr={3}>
              <CheckIcon className={classes.submitIcon} />
            </Box>
          )}
          <Typography>{T('newE.titleStep1')}</Typography>
        </Box>
        {step1 && <UndoIcon className={classes.asButton} onClick={handleUndo} />}
      </Box>
      <Box hidden={step1}>
        <Box overflow="hidden">
          <GridList className={classes.gridList} cols={isDesktopScreen ? 1.5 : 3.5} spacing={9} cellHeight="auto">
            {targetDataEntries.map(([key, t]) => (
              <GridListTile key={key}>
                <Card
                  className={clsx(classes.card, kind === key ? classes.cardActive : '')}
                  variant="outlined"
                  onClick={handleSelectTarget(key)}
                >
                  <Box display="flex" justifyContent="center" alignItems="center" height="100px">
                    <Box display="flex" justifyContent="center" flex={1}>
                      {iconByKind(key)}
                    </Box>
                    <Box display="flex" justifyContent="center" flex={1.5} textAlign="center">
                      <Typography variant="button">{transByKind(key)}</Typography>
                    </Box>
                  </Box>
                </Card>
              </GridListTile>
            ))}
          </GridList>
        </Box>
        {kind && (
          <Box overflow="hidden">
            <Box my={6}>
              <Divider />
            </Box>
            {targetData[kind].categories ? (
              <GridList className={classes.gridList} cols={isDesktopScreen ? 2.5 : 4.5} spacing={9} cellHeight="auto">
                {targetData[kind].categories!.map((d: any) => (
                  <GridListTile key={d.key}>
                    <Card
                      className={clsx(classes.card, action === d.key ? classes.cardActive : '')}
                      variant="outlined"
                      onClick={handleSelectAction(d.key)}
                    >
                      <Box display="flex" justifyContent="center" alignItems="center" height="50px">
                        <Box display="flex" justifyContent="center" alignItems="center" flex={1}>
                          {action === d.key ? <RadioButtonCheckedOutlinedIcon /> : <RadioButtonUncheckedOutlinedIcon />}
                        </Box>
                        <Box display="flex" justifyContent="center" alignItems="center" flex={2} px={1.5}>
                          <Typography variant="button">{d.name}</Typography>
                        </Box>
                      </Box>
                    </Card>
                  </GridListTile>
                ))}
              </GridList>
            ) : kind === 'KernelChaos' ? (
              <Kernel onSubmit={handleSubmitStep1} />
            ) : kind === 'TimeChaos' ? (
              <TargetGenerated
                data={targetData[kind].spec!}
                validationSchema={schema.TimeChaos!.default}
                onSubmit={handleSubmitStep1}
              />
            ) : kind === 'StressChaos' ? (
              <Stress onSubmit={handleSubmitStep1} />
            ) : null}
          </Box>
        )}
        {action && !submitDirectly.includes(action) && (
          <Box p={3}>
            <Box mb={6}>
              <Divider />
            </Box>
            <TargetGenerated
              // Force re-rendered after action changed
              key={kind + action}
              kind={kind}
              data={targetData[kind as Kind].categories!.filter(({ key }) => key === action)[0].spec}
              validationSchema={schema[kind as Kind]![action]}
              onSubmit={handleSubmitStep1}
            />
          </Box>
        )}
      </Box>
    </Paper>
  )
}

export default Step1
