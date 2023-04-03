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
import { Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material'
import { TypesPod } from 'openapi/index.schemas'

import PaperContainer from '@ui/mui-extends/esm/PaperContainer'

import { T } from 'components/T'

import { TargetsTableActions } from '.'

interface PodsTableProps extends TargetsTableActions {
  data: TypesPod[]
}

export default function PosTable({ data, handleSelect, isSelected }: PodsTableProps) {
  return (
    <TableContainer component={PaperContainer}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell />
            <TableCell>
              <T id="common.name" />
            </TableCell>
            <TableCell>
              <T id="k8s.namespace" />
            </TableCell>
            <TableCell>
              <T id="common.ip" />
            </TableCell>
            <TableCell>
              <T id="common.state" />
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((d) => {
            const key = `${d.namespace}:${d.name}`

            return (
              <TableRow key={key} onClick={handleSelect(key)}>
                <TableCell padding="checkbox">
                  <Checkbox checked={isSelected(key)} />
                </TableCell>
                <TableCell>{d.name}</TableCell>
                <TableCell>{d.namespace}</TableCell>
                <TableCell>{d.ip}</TableCell>
                <TableCell>{d.state}</TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}
