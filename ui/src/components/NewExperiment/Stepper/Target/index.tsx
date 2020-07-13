import { ChaosKindKeyMap, resetOtherChaos } from 'lib/formikhelpers'
import React, { useEffect, useState } from 'react'

import { Experiment } from 'components/NewExperiment/types'
import Network from './Network'
import Pod from './Pod'
import Tabs from 'components/Tabs'
import { useFormikContext } from 'formik'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'File system I/O' },
  { label: 'Linux Kernel' },
  { label: 'Clock' },
  { label: 'Stress CPU/Memory' },
]

const Target: React.FC = () => {
  const formikCtx = useFormikContext<Experiment>()

  const [tabIndex, setTabIndex] = useState(0)

  useEffect(() => {
    const kind = formikCtx.values.target.kind

    if (kind) {
      setTabIndex(Object.keys(ChaosKindKeyMap).indexOf(kind))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleActionChange = (kind: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    resetOtherChaos(formikCtx, kind, e.target.value)

  const tabPanels = [
    <Pod {...formikCtx} handleActionChange={handleActionChange('PodChaos')} />,
    <Network {...formikCtx} handleActionChange={handleActionChange('NetworkChaos')} />,
  ]

  const props = {
    tabs,
    tabPanels,
    tabIndex,
    setTabIndex,
  }

  return <Tabs {...props} />
}

export default Target
