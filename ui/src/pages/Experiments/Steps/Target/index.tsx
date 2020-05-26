import React from 'react'

import VerticalTabs from 'components/VerticalTabs'
import Pod from './Pod'
import Network from './Network'
import { StepProps } from '../../types'

const tabs = [
  { label: 'Pod Lifecycle' },
  { label: 'Network' },
  { label: 'Fife system I/O', disabled: true },
  { label: 'Linux Kernel', disabled: true },
  { label: 'Clock', disabled: true },
  { label: 'Stress CPU/Memory', disabled: true },
]

export default function TargetStep({ formProps }: StepProps) {
  const tabPanels = [<Pod {...formProps} />, <Network {...formProps} />]

  const props = {
    tabs,
    tabPanels,
  }

  return <VerticalTabs {...props} />
}
