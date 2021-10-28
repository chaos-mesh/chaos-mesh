/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { forwardRef, useImperativeHandle, useState } from 'react'
import { setEnv, setExternalExperiment } from 'slices/experiments'

import { Box } from '@material-ui/core'
import ByYAML from './ByYAML'
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
import { parseYAML } from 'lib/formikhelpers'
import { useStoreDispatch } from 'store'

type PanelType = 'initial' | 'existing' | 'yaml'

export interface NewExperimentHandles {
  setPanel: React.Dispatch<React.SetStateAction<PanelType>>
}

interface NewExperimentProps {
  onSubmit?: (parsedValues: any) => void
  loadFrom?: boolean
  inWorkflow?: boolean
  inSchedule?: boolean
}

const NewExperiment: React.ForwardRefRenderFunction<NewExperimentHandles, NewExperimentProps> = (
  { onSubmit, loadFrom = true, inWorkflow, inSchedule },
  ref
) => {
  const dispatch = useStoreDispatch()

  const [panel, setPanel] = useState<PanelType>('initial')

  useImperativeHandle(ref, () => ({
    setPanel,
  }))

  const onChange = (_: any, newValue: PanelType) => {
    setPanel(newValue)
  }

  const fillExperiment = (original: any) => {
    const { kind, basic, spec } = parseYAML(original)
    const env = basic.spec.address.length ? 'physic' : 'k8s'
    const action =
      env === 'k8s' ? spec.action ?? '' : (kind as any) !== 'ProcessChaos' && kind !== 'TimeChaos' ? spec.action : ''

    dispatch(setEnv(env))
    dispatch(
      setExternalExperiment({
        kindAction: [kind, action],
        spec,
        basic,
      })
    )

    setPanel('initial')
  }

  return (
    <TabContext value={panel}>
      {loadFrom && (
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <TabList onChange={onChange}>
            <Tab label={T(`${inSchedule ? 'newS' : 'newE'}.title`)} value="initial" />
            <Tab label={T('newE.loadFrom')} value="existing" />
            <Tab label={T('newE.byYAML')} value="yaml" />
          </TabList>
        </Box>
      )}
      <TabPanel value="initial" sx={{ p: 0, pt: 6 }}>
        <Space spacing={6}>
          <Step1 />
          <Step2 inWorkflow={inWorkflow} inSchedule={inSchedule} />
          <Step3 onSubmit={onSubmit ? onSubmit : undefined} inSchedule={inSchedule} />
        </Space>
      </TabPanel>
      <TabPanel value="existing" sx={{ p: 0, pt: 6 }}>
        {loadFrom && <LoadFrom callback={fillExperiment} inSchedule={inSchedule} inWorkflow={inWorkflow} />}
      </TabPanel>
      <TabPanel value="yaml" sx={{ p: 0, pt: 6 }}>
        <ByYAML callback={fillExperiment} />
      </TabPanel>
    </TabContext>
  )
}

export default forwardRef(NewExperiment)
