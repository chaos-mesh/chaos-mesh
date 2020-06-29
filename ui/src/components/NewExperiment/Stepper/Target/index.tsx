import { ChaosKindKeyMap, resetOtherChaos } from 'lib/formikhelpers'
import React, { useEffect, useState } from 'react'

import Network from './Network'
import Pod from './Pod'
import { StepperFormProps } from 'components/NewExperiment/types'
import Tabs from 'components/Tabs'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'File system I/O' },
  { label: 'Linux Kernel' },
  { label: 'Clock' },
  { label: 'Stress CPU/Memory' },
]

interface TargetProps {
  formProps: StepperFormProps
}

const Target: React.FC<TargetProps> = ({ formProps }) => {
  const [tabIndex, setTabIndex] = useState(0)

  useEffect(() => {
    const kind = formProps.values.target.kind

    if (kind) {
      setTabIndex(Object.keys(ChaosKindKeyMap).indexOf(kind))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleActionChange = (kind: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    resetOtherChaos(formProps, kind, e.target.value)

  const tabPanels = [
    <Pod {...formProps} handleActionChange={handleActionChange('PodChaos')} />,
    <Network {...formProps} handleActionChange={handleActionChange('NetworkChaos')} />,
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
