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
import { useComponentActions } from '@/zustand/component'
import type { Env } from '@/zustand/experiment'
import { Alert } from '@mui/material'
import { getIn, useFormikContext } from 'formik'
import { useMemo } from 'react'

import { T } from '@/components/T'

import { calculateBlastRadius } from '@/lib/utils'

import PhysicalMachinesTable from './PhysicalMachinesTable'
import PodsTable from './PodsTable'

interface TargetsTableProps {
  env: Env
  scope?: string
  modeScope?: string
  data: any[]
}

export interface TargetsTableActions {
  handleSelect: (name: string) => () => void
  isSelected: (name: string) => boolean
}

const TargetsTable = ({ env, scope = 'scope', modeScope = '', data }: TargetsTableProps) => {
  const originalTargets = useMemo(() => data.map((d) => `${d.namespace}:${d.name}`), [data])
  const targetsCount = originalTargets.length

  const { values, setFieldValue } = useFormikContext()
  const formikTargets: string[] = getIn(values, `${scope}.${env === 'k8s' ? 'pods' : 'physicalMachines'}`) || []

  const selected = formikTargets.length > 0 ? formikTargets : originalTargets
  const mode = getIn(values, modeScope ? `${modeScope}.mode` : 'mode')
  const modeValue = getIn(values, modeScope ? `${modeScope}.value` : 'value')
  const blastRadius = calculateBlastRadius(mode, modeValue, selected.length)
  const allCandidatesAffected = blastRadius?.exact && blastRadius.maximum === selected.length
  const isSelected = (name: string) => selected.indexOf(name) !== -1
  const setSelected = (newVal: string[]) =>
    setFieldValue(`${scope}.${env === 'k8s' ? 'pods' : 'physicalMachines'}`, newVal)

  const { setAlert } = useComponentActions()

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
      setAlert({
        type: 'warning',
        message: 'Please select at least one target.',
      })

      return
    }

    setSelected(newSelected.length === targetsCount ? [] : newSelected)
  }

  return (
    <>
      {env === 'k8s' && (
        <Alert severity={selected.length === 0 || !blastRadius ? 'warning' : 'info'} sx={{ mb: 2 }}>
          {selected.length === 0 ? (
            <T id="newE.scope.blastRadiusNoTargets" />
          ) : blastRadius ? (
            <>
              <T
                id={blastRadius.exact ? 'newE.scope.blastRadiusExact' : 'newE.scope.blastRadiusRange'}
                values={{
                  affected: blastRadius.maximum,
                  candidates: selected.length,
                  minimum: blastRadius.minimum,
                  maximum: blastRadius.maximum,
                }}
              />{' '}
              <T
                id={allCandidatesAffected ? 'newE.scope.blastRadiusAllTargets' : 'newE.scope.blastRadiusRandomTargets'}
              />
            </>
          ) : (
            <T id="newE.scope.blastRadiusInvalid" />
          )}
        </Alert>
      )}
      {env === 'k8s' && <PodsTable data={data} handleSelect={handleSelect} isSelected={isSelected} />}
      {env === 'physic' && <PhysicalMachinesTable data={data} handleSelect={handleSelect} isSelected={isSelected} />}
    </>
  )
}

export default TargetsTable
