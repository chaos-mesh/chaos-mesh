import { Box, Grid, SvgIcon, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { ReactComponent as ClockIcon } from 'images/chaos/time.svg'
import { ReactComponent as DNSIcon } from 'images/chaos/dns.svg'
import { Experiment } from 'api/experiments.type'
import { ExperimentKind } from 'components/NewExperiment/types'
import { ReactComponent as FileSystemIOIcon } from 'images/chaos/io.svg'
import { ReactComponent as LinuxKernelIcon } from 'images/chaos/kernel.svg'
import { ReactComponent as NetworkIcon } from 'images/chaos/network.svg'
import { ReactComponent as PodLifecycleIcon } from 'images/chaos/pod.svg'
import { ReactComponent as StressIcon } from 'images/chaos/stress.svg'
import api from 'api'

interface ChaosProps {
  icon: JSX.Element
  kind: ExperimentKind
  data: { sum: number }
}

const Chaos: React.FC<ChaosProps> = ({ icon, kind, data }) => (
  <Grid item xs={12} md={4}>
    <Box display="flex" alignItems="center" py={4.5}>
      <Box display="flex" justifyContent="center" alignItems="center" flex={1}>
        {icon}
      </Box>
      <Box flex={1.5}>
        <Typography variant="button" color="textSecondary" gutterBottom>
          {kind}
        </Typography>
        <Typography variant="h5">{data.sum}</Typography>
      </Box>
    </Box>
  </Grid>
)

const TotalExperiments = () => {
  const [experiments, setExperiments] = useState<Record<ExperimentKind, number>>({
    PodChaos: 0,
    NetworkChaos: 0,
    IoChaos: 0,
    KernelChaos: 0,
    TimeChaos: 0,
    StressChaos: 0,
    DNSChaos: 0,
  })

  const fetchExperiments = () => {
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments((prev) => ({ ...prev, ...processExperiments(data) })))
      .catch(console.error)
  }

  const processExperiments = (data: Experiment[]) =>
    data.reduce<Record<string, number>>((acc, e) => {
      if (acc[e.kind]) {
        acc[e.kind] += 1
      } else {
        acc[e.kind] = 1
      }

      return acc
    }, {})

  useEffect(fetchExperiments, [])

  return (
    <Grid container>
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <PodLifecycleIcon />
          </SvgIcon>
        }
        kind="PodChaos"
        data={{ sum: experiments['PodChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <NetworkIcon />
          </SvgIcon>
        }
        kind="NetworkChaos"
        data={{ sum: experiments['NetworkChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <FileSystemIOIcon />
          </SvgIcon>
        }
        kind="IoChaos"
        data={{ sum: experiments['IoChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <LinuxKernelIcon />
          </SvgIcon>
        }
        kind="KernelChaos"
        data={{ sum: experiments['KernelChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <ClockIcon />
          </SvgIcon>
        }
        kind="TimeChaos"
        data={{ sum: experiments['TimeChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <StressIcon />
          </SvgIcon>
        }
        kind="StressChaos"
        data={{ sum: experiments['StressChaos'] }}
      />
      <Chaos
        icon={
          <SvgIcon fontSize="large" color="primary">
            <DNSIcon />
          </SvgIcon>
        }
        kind="DNSChaos"
        data={{ sum: experiments['DNSChaos'] }}
      />
    </Grid>
  )
}

export default TotalExperiments
