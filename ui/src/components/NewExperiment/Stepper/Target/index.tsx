import { ChaosKindKeyMap, resetOtherChaos } from 'lib/formikhelpers'
import React, { useEffect, useState } from 'react'

import { Experiment } from 'components/NewExperiment/types'
import IO from './IO'
import Kernel from './Kernel'
import Network from './Network'
import Pod from './Pod'
import Stress from './Stress'
import Tabs from 'components/Tabs'
import Time from './Time'
import { useFormikContext } from 'formik'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'File System I/O' },
  { label: 'Linux Kernel' },
  { label: 'Clock' },
  { label: 'Stress CPU/Memory' },
]

const Target: React.FC = () => {
  const formikCtx = useFormikContext<Experiment>()
  const kind = formikCtx.values.target.kind

  const [tabIndex, setTabIndex] = useState(0)

  useEffect(() => {
    if (kind) {
      setTabIndex(Object.keys(ChaosKindKeyMap).indexOf(kind))
    }
  }, [kind])

  const handleActionChange = (kind: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    resetOtherChaos(formikCtx, kind, e.target.value)

  const tabPanels = [
    <Pod {...formikCtx} handleActionChange={handleActionChange('PodChaos')} />,
    <Network {...formikCtx} handleActionChange={handleActionChange('NetworkChaos')} />,
    <IO {...formikCtx} handleActionChange={handleActionChange('IoChaos')} />,
    <Kernel {...formikCtx} />,
    <Time {...formikCtx} />,
    <Stress {...formikCtx} />,
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
