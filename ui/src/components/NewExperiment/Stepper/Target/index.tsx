import Network from './Network'
import Pod from './Pod'
import React from 'react'
import { StepperFormProps } from 'components/NewExperiment/types'
import VerticalTabs from 'components/VerticalTabs'
import { resetOtherChaos } from 'lib/formikhelpers'
import { targetVerticalTabsKinds as tabKinds } from 'lib/formikhelpers'

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
}

const Target: React.FC<TargetProps> = ({ formProps }) => {
  const { setFieldValue } = formProps

  const handleVerticalTabsChangeCallback = (index: number) => {
    setFieldValue('target.kind', tabKinds.map((k) => k.kind)[index])
  }

  const handleActionChange = (e: React.ChangeEvent<HTMLInputElement>) => resetOtherChaos(formProps, e.target.value)

  const tabPanels = [
    <Pod {...formProps} handleActionChange={handleActionChange} />,
    <Network {...formProps} handleActionChange={handleActionChange} />,
  ]

  const props = {
    tabs,
    tabPanels,
    onChangeCallback: handleVerticalTabsChangeCallback,
  }

  return <VerticalTabs {...props} />
}

export default Target
