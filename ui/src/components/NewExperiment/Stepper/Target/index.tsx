import Network from './Network'
import Pod from './Pod'
import React from 'react'
import { StepperFormProps } from 'components/NewExperiment/types'
import VerticalTabs from 'components/VerticalTabs'
import { resetOtherChaos } from 'lib/formikhelpers'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'File system I/O', disabled: true },
  { label: 'Linux Kernel', disabled: true },
  { label: 'Clock', disabled: true },
  { label: 'Stress CPU/Memory', disabled: true },
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

  return <VerticalTabs {...props} />
}

export default Target
