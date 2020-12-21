import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { iconByKind, transByKind } from 'lib/byKind'

import { Experiment } from 'api/experiments.type'
import { ExperimentKind } from 'components/NewExperiment/types'
import api from 'api'

interface ChaosProps {
  kind: ExperimentKind
  data: { sum: number }
}

const Chaos: React.FC<ChaosProps> = ({ kind, data }) => (
  <Grid item xs={12} md={4}>
    <Box display="flex" alignItems="center" py={4.5}>
      <Box display="flex" justifyContent="center" alignItems="center" flex={1}>
        {iconByKind(kind)}
      </Box>
      <Box flex={1.5}>
        <Typography variant="button" color="textSecondary" gutterBottom>
          {transByKind(kind)}
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
      <Chaos kind="PodChaos" data={{ sum: experiments['PodChaos'] }} />
      <Chaos kind="NetworkChaos" data={{ sum: experiments['NetworkChaos'] }} />
      <Chaos kind="IoChaos" data={{ sum: experiments['IoChaos'] }} />
      <Chaos kind="KernelChaos" data={{ sum: experiments['KernelChaos'] }} />
      <Chaos kind="TimeChaos" data={{ sum: experiments['TimeChaos'] }} />
      <Chaos kind="StressChaos" data={{ sum: experiments['StressChaos'] }} />
      <Chaos kind="DNSChaos" data={{ sum: experiments['DNSChaos'] }} />
    </Grid>
  )
}

export default TotalExperiments
