import { Box, Button, Card, Chip, Grid, SvgIcon, Typography } from '@material-ui/core'
import { Experiment, ExperimentKind } from 'components/NewExperiment/types'
import React, { useState } from 'react'
import { createStyles, makeStyles } from '@material-ui/core/styles'

import AccessTimeIcon from '@material-ui/icons/AccessTime'
import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft'
import IO from './IO'
import Kernel from './Kernel'
import { ReactComponent as LinuxKernelIcon } from './images/linux-kernel.svg'
import Network from './Network'
import PanToolOutlinedIcon from '@material-ui/icons/PanToolOutlined'
import Pod from './Pod'
import { ReactComponent as PodLifecycleIcon } from './images/pod-lifecycle.svg'
import PortableWifiOffOutlinedIcon from '@material-ui/icons/PortableWifiOffOutlined'
import SelectAllOutlinedIcon from '@material-ui/icons/SelectAllOutlined'
import Stress from './Stress'
import Time from './Time'
import { resetOtherChaos } from 'lib/formikhelpers'
import { useFormikContext } from 'formik'

const tabs = [
  {
    key: 'PodChaos',
    label: 'Pod Lifecycle',
    icon: (
      <SvgIcon fontSize="large">
        <PodLifecycleIcon />
      </SvgIcon>
    ),
  },
  {
    key: 'NetworkChaos',
    label: 'Network',
    icon: <PortableWifiOffOutlinedIcon fontSize="large" />,
  },
  { key: 'IoChaos', label: 'File System I/O', icon: <PanToolOutlinedIcon fontSize="large" /> },
  {
    key: 'KernelChaos',
    label: 'Linux Kernel',
    icon: (
      <SvgIcon fontSize="large">
        <LinuxKernelIcon />
      </SvgIcon>
    ),
  },
  { key: 'TimeChaos', label: 'Clock', icon: <AccessTimeIcon fontSize="large" /> },
  {
    key: 'StressChaos',
    label: 'Stress CPU/Memory',
    icon: <SelectAllOutlinedIcon fontSize="large" />,
  },
]

const useStyles = makeStyles((theme) =>
  createStyles({
    card: {
      position: 'relative',
      cursor: 'pointer',
      '&:hover': {
        borderColor: theme.palette.primary.main,
      },
    },
  })
)

const Target: React.FC = () => {
  const classes = useStyles()

  const formikCtx = useFormikContext<Experiment>()
  const kind = formikCtx.values.target.kind

  const [selected, setSelected] = useState<ExperimentKind | ''>('')

  const handleActionChange = (kind: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    resetOtherChaos(formikCtx, kind, e.target.value)

  const handleSelectTarget = (kind: ExperimentKind) => () => setSelected(kind)

  const renderBySelected = () => {
    switch (selected) {
      case 'PodChaos':
        return <Pod handleActionChange={handleActionChange('PodChaos')} />
      case 'NetworkChaos':
        return <Network handleActionChange={handleActionChange('NetworkChaos')} />
      case 'IoChaos':
        return <IO handleActionChange={handleActionChange('IoChaos')} />
      case 'KernelChaos':
        return <Kernel />
      case 'TimeChaos':
        return <Time />
      case 'StressChaos':
        return <Stress />
      default:
        return null
    }
  }

  return (
    <>
      {selected === '' && (
        <Grid container spacing={3}>
          {tabs.map((tab) => (
            <Grid key={tab.key} item xs={6}>
              <Card variant="outlined" className={classes.card} onClick={handleSelectTarget(tab.key as ExperimentKind)}>
                <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="150px">
                  <Box display="flex" justifyContent="center" alignItems="center" flex={2}>
                    {tab.icon}
                  </Box>
                  <Box display="flex" justifyContent="center" alignItems="center" flex={1} px={3} textAlign="center">
                    <Typography variant="overline">{tab.label}</Typography>
                  </Box>
                </Box>
                {kind === tab.key && (
                  <Box position="absolute" top="0.5rem" right="0.5rem">
                    <Chip label="Configured" icon={<CheckCircleOutlineIcon />} size="small" color="primary" />
                  </Box>
                )}
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
      {selected !== '' && (
        <Box mb={3}>
          <Button startIcon={<ChevronLeftIcon />} onClick={() => setSelected('')}>
            Back
          </Button>
        </Box>
      )}
      {renderBySelected()}
    </>
  )
}

export default Target
