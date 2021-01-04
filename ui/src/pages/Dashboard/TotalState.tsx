import React, { useEffect, useRef, useState } from 'react'

import NotFound from 'components-mui/NotFound'
import { StateOfExperiments } from 'api/experiments.type'
import T from 'components/T'
import api from 'api'
import genChaosStatePieChart from 'lib/d3/chaosStatePieChart'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

interface TotalStateProps {
  className?: string
}

const TotalState: React.FC<TotalStateProps> = (props) => {
  const intl = useIntl()

  const { theme } = useStoreSelector((state) => state.settings)
  const [s, setS] = useState<StateOfExperiments>({
    Running: 0,
    Waiting: 0,
    Paused: 0,
    Failed: 0,
    Finished: 0,
  })

  const chaosStatePieChartRef = useRef<any>(null)

  const fetchState = () => {
    api.experiments
      .state()
      .then((resp) => setS(resp.data))
      .catch(console.error)
  }

  useEffect(() => {
    fetchState()

    const id = setInterval(fetchState, 15000)

    return () => clearInterval(id)
  }, [])

  useEffect(() => {
    if (typeof chaosStatePieChartRef.current === 'function') {
      chaosStatePieChartRef.current(s)

      return
    }

    const update = genChaosStatePieChart({
      root: chaosStatePieChartRef.current,
      chaosStatus: s,
      intl,
      theme,
    })
    chaosStatePieChartRef.current = update
  }, [s, intl, theme])

  return (
    <>
      <div {...props} ref={chaosStatePieChartRef} />
      {s && Object.values(s).filter((d) => d !== 0).length === 0 && (
        <NotFound>{T('experiments.noExperimentsFound')}</NotFound>
      )}
    </>
  )
}

export default TotalState
