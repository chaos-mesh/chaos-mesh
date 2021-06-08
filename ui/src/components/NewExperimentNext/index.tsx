import React, { useImperativeHandle, useState } from 'react'

import { Box } from '@material-ui/core'
import LoadFrom from './LoadFrom'
import Space from 'components-mui/Space'
import Step1 from './Step1'
import Step2 from './Step2'
import Step3 from './Step3'
import T from 'components/T'
import Tab from '@material-ui/core/Tab'
import TabContext from '@material-ui/lab/TabContext'
import TabList from '@material-ui/lab/TabList'
import TabPanel from '@material-ui/lab/TabPanel'

type PanelType = 'initial' | 'existing'

export interface NewExperimentHandles {
  setPanel: React.Dispatch<React.SetStateAction<PanelType>>
}

interface NewExperimentProps {
  initPanel?: PanelType
  onSubmit?: (experiment: { target: any; basic: any }) => void
  loadFrom?: boolean
  inWorkflow?: boolean
  inSchedule?: boolean
}

const NewExperiment: React.ForwardRefRenderFunction<NewExperimentHandles, NewExperimentProps> = (
  { initPanel = 'initial', onSubmit, loadFrom = true, inWorkflow = false, inSchedule = false },
  ref
) => {
  const [panel, setPanel] = useState<PanelType>(initPanel)

  useImperativeHandle(ref, () => ({
    setPanel,
  }))

  const onChange = (_: any, newValue: PanelType) => {
    setPanel(newValue)
  }

  const loadCallback = () => setPanel('initial')

  return (
    <TabContext value={panel}>
      {loadFrom && (
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <TabList onChange={onChange}>
            <Tab label={T(`${inSchedule ? 'newS' : 'newE'}.title`)} value="initial" />
            <Tab label={T('newE.loadFrom')} value="existing" />
          </TabList>
        </Box>
      )}
      <TabPanel value="initial" sx={{ p: 0, pt: 6 }}>
        <Space spacing={6}>
          <Step1 />
          <Step2 inWorkflow={inWorkflow} inSchedule={inSchedule} />
          <Step3 onSubmit={onSubmit ? onSubmit : undefined} />
        </Space>
      </TabPanel>
      <TabPanel value="existing" sx={{ p: 0, pt: 6 }}>
        {loadFrom && <LoadFrom loadCallback={loadCallback} inSchedule={inSchedule} />}
      </TabPanel>
    </TabContext>
  )
}

export default React.forwardRef(NewExperiment)
