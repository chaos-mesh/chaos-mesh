import Network from './Network'
import Pod from './Pod'
import React from 'react'
import { StepperFormProps } from 'components/NewExperiment/types'
import VerticalTabs from 'components/VerticalTabs'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'Fife system I/O', disabled: true },
  { label: 'Linux Kernel', disabled: true },
  { label: 'Clock', disabled: true },
  { label: 'Stress CPU/Memory', disabled: true },
]

interface TargetStepProps {
  formProps: StepperFormProps
}

const TargetStep: React.FC<TargetStepProps> = ({ formProps }) => {
  const tabPanels = [<Pod {...formProps} />, <Network {...formProps} />]

  const props = {
    tabs,
    tabPanels,
  }

  return <VerticalTabs {...props} />
}

export default TargetStep
