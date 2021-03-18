import { Box, Breadcrumbs, Link } from '@material-ui/core'
import React, { useState } from 'react'

import { Experiment } from 'components/NewExperiment/types'
import LoadFrom from './LoadFrom'
import Step1 from './Step1'
import Step2 from './Step2'
import Step3 from './Step3'
import T from 'components/T'

type PanelType = 'initial' | 'existing'

interface NewExperimentProps {
  initPanel?: PanelType
  onSubmit?: (values: Experiment) => void
  loadFrom?: boolean
}

const NewExperiment: React.FC<NewExperimentProps> = ({ initPanel = 'initial', onSubmit, loadFrom = true }) => {
  const [showNewPanel, setShowNewPanel] = useState<PanelType>(initPanel)

  const loadCallback = () => setShowNewPanel('initial')

  const internalOnSubmit = (values: Experiment) => {
    onSubmit!(values)

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
      <Box style={{ display: showNewPanel === 'initial' ? 'initial' : 'none' }}>
        <Step1 />
        <Box mt={6}>
          <Step2 />
        </Box>
        <Box mt={loadFrom ? 6 : undefined}>
          <Step3 onSubmit={onSubmit ? internalOnSubmit : undefined} />
        </Box>
      </Box>
      {loadFrom && (
        <Box style={{ display: showNewPanel === 'existing' ? 'initial' : 'none' }}>
          <LoadFrom loadCallback={loadCallback} />
        </Box>
      )}
    </>
  )
}

export default NewExperiment
