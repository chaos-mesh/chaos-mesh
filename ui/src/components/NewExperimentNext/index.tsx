import { forwardRef, useImperativeHandle, useState } from 'react'

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
import _snakecase from 'lodash.snakecase'
import { data as scheduleSpecificData } from 'components/Schedule/types'
import { setExternalExperiment } from 'slices/experiments'
import { toCamelCase } from 'lib/utils'
import { useStoreDispatch } from 'store'
import { yamlToExperiment } from 'lib/formikhelpers'

type PanelType = 'initial' | 'existing' | 'yaml'

export interface NewExperimentHandles {
  setPanel: React.Dispatch<React.SetStateAction<PanelType>>
}

interface NewExperimentProps {
  onSubmit?: (experiment: { target: any; basic: any }) => void
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
    if (original.kind === 'Schedule') {
      const kind = original.spec.type
      const data = yamlToExperiment({
        kind,
        metadata: original.metadata,
        spec: original.spec[toCamelCase(kind)],
      })
      delete original.spec[toCamelCase(kind)]

      dispatch(
        setExternalExperiment({
          kindAction: [kind, data.target[_snakecase(kind)].action ?? ''],
          target: data.target,
          basic: { ...data.basic, ...scheduleSpecificData, ...original.spec },
        })
      )

      setPanel('initial')

      return
    }

    const data = yamlToExperiment(original)

    const kind = data.target.kind

    dispatch(
      setExternalExperiment({
        kindAction: [kind, data.target[_snakecase(kind)].action ?? ''],
        target: data.target,
        basic: data.basic,
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
          <Step3 onSubmit={onSubmit ? onSubmit : undefined} />
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
