/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { getIn, useFormikContext } from 'formik'
import { useMemo } from 'react'

import { useStoreDispatch } from 'store'

import { Env } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'

import PhysicalMachinesTable from './PhysicalMachinesTable'
import PodsTable from './PodsTable'

interface TargetsTableProps {
  env: Env
  scope?: string
  data: any[]
}

export interface TargetsTableActions {
  handleSelect: (name: string) => () => void
  isSelected: (name: string) => boolean
}

const TargetsTable = ({ env, scope = 'scope', data }: TargetsTableProps) => {
  const originalTargets = useMemo(() => data.map((d) => `${d.namespace}:${d.name}`), [data])
  const targetsCount = originalTargets.length

  const { values, setFieldValue } = useFormikContext()
  const formikTargets: string[] = getIn(values, `${scope}.pods`)

  const selected = formikTargets.length > 0 ? formikTargets : originalTargets
  const isSelected = (name: string) => selected.indexOf(name) !== -1
  const setSelected = (newVal: string[]) => setFieldValue(`${scope}.pods`, newVal)

  const dispatch = useStoreDispatch()

  const handleSelect = (name: string) => () => {
    const selectedIndex = selected.indexOf(name)
    let newSelected: string[] = []

    if (selectedIndex === -1) {
      newSelected = [...selected, name]
    } else if (selectedIndex === 0) {
      newSelected = selected.slice(1)
    } else if (selectedIndex === selected.length - 1) {
      newSelected = selected.slice(0, -1)
    } else if (selectedIndex > 0) {
      newSelected = [...selected.slice(0, selectedIndex), ...selected.slice(selectedIndex + 1)]
    }

    if (newSelected.length === 0) {
      dispatch(
        setAlert({
          type: 'warning',
          message: 'Please select at least one pod.',
        })
      )

      return
    }

    setSelected(newSelected.length === targetsCount ? [] : newSelected)
  }

  return (
    <>
      {env === 'k8s' && <PodsTable data={data} handleSelect={handleSelect} isSelected={isSelected} />}
      {env === 'physic' && <PhysicalMachinesTable data={data} handleSelect={handleSelect} isSelected={isSelected} />}
    </>
  )
}

export default TargetsTable
