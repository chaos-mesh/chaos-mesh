import Network from './Network'
import Pod from './Pod'
import React from 'react'
import { StepperFormProps } from 'components/NewExperiment/types'
import Tabs from 'components/Tabs'
import { resetOtherChaos } from 'lib/formikhelpers'

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
  tabIndex: number
  setTabIndex: (index: number) => void
}

const Target: React.FC<TargetProps> = ({ formProps, tabIndex, setTabIndex }) => {
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
