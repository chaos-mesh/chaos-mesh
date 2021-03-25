import { Box, Breadcrumbs, Link } from '@material-ui/core'
import React, { useImperativeHandle, useState } from 'react'

import LoadFrom from './LoadFrom'
import Space from 'components-mui/Space'
import Step1 from './Step1'
import Step2 from './Step2'
import Step3 from './Step3'
import T from 'components/T'

type PanelType = 'initial' | 'existing'

export interface NewExperimentHandles {
  setShowNewPanel: React.Dispatch<React.SetStateAction<PanelType>>
}

interface NewExperimentProps {
  initPanel?: PanelType
  onSubmit?: (experiment: { target: any; basic: any }) => void
  loadFrom?: boolean
}

const NewExperiment: React.ForwardRefRenderFunction<NewExperimentHandles, NewExperimentProps> = (
  { initPanel = 'initial', onSubmit, loadFrom = true },
  ref
) => {
  const [showNewPanel, setShowNewPanel] = useState<PanelType>(initPanel)

  useImperativeHandle(ref, () => ({
    setShowNewPanel,
  }))

  const loadCallback = () => setShowNewPanel('initial')

  const internalOnSubmit = (experiment: any) => {
    onSubmit!(experiment)

    setShowNewPanel('existing')
  }

  return (
    <>
      {loadFrom && (
        <Box mb={6}>
          <Breadcrumbs aria-label="breadcrumb">
            <Link
              href="#"
              color={showNewPanel === 'initial' ? 'primary' : 'inherit'}
              onClick={() => setShowNewPanel('initial')}
            >
              {T('newE.title')}
            </Link>
            <Link
              href="#"
              color={showNewPanel === 'existing' ? 'primary' : 'inherit'}
              onClick={() => setShowNewPanel('existing')}
            >
              {T('newE.loadFrom')}
            </Link>
          </Breadcrumbs>
        </Box>
      )}
      <Space spacing={6} vertical style={{ display: showNewPanel === 'initial' ? 'initial' : 'none' }}>
        <Step1 />
        <Step2 />
        <Step3 onSubmit={onSubmit ? internalOnSubmit : undefined} />
      </Space>
      {loadFrom && (
        <Box style={{ display: showNewPanel === 'existing' ? 'initial' : 'none' }}>
          <LoadFrom loadCallback={loadCallback} />
        </Box>
      )}
    </>
  )
}

export default React.forwardRef(NewExperiment)
