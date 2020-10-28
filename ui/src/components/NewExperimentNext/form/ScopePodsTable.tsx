import { Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import React, { useMemo } from 'react'
import { getIn, useFormikContext } from 'formik'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import Paper from 'components-mui/Paper'
import T from 'components/T'
import { useStoreDispatch } from 'store'

const PaperContainer: React.FC = ({ children }) => (
  <Paper style={{ maxHeight: 768, overflow: 'scroll' }}>{children}</Paper>
)

interface ScopePodsTableProps {
  scope?: string
  pods: any[]
}

const ScopePodsTable: React.FC<ScopePodsTableProps> = ({ scope = 'scope', pods }) => {
  const originalPods = useMemo(() => pods.map((d) => d.name).reduce((acc, d) => acc.concat(d), []), [pods])
  const podsCount = originalPods.length

  const { values, setFieldValue } = useFormikContext()
  const formikPods: string[] = getIn(values, `${scope}.pods`)

  const selected = formikPods.length > 0 ? formikPods : originalPods
  const isSelected = (name: string) => selected.indexOf(name) !== -1
  const setSelected = (newVal: string[]) => setFieldValue(`${scope}.pods`, newVal)

  const dispatch = useStoreDispatch()

  const handleSelect = (name: string) => (_: React.MouseEvent<unknown>) => {
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
      dispatch(setAlertOpen(true))

      return
    }

    setSelected(newSelected.length === podsCount ? [] : newSelected)
  }

  return (
    <TableContainer component={PaperContainer}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell />
            <TableCell>{T('newE.scope.podsTable.name')}</TableCell>
            <TableCell>{T('newE.scope.podsTable.namespace')}</TableCell>
            <TableCell>{T('newE.scope.podsTable.ip')}</TableCell>
            <TableCell>{T('newE.scope.podsTable.state')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {pods.map((pod) => (
            <TableRow key={pod.name + pod.namespace} onClick={handleSelect(pod.name)}>
              <TableCell padding="checkbox">
                <Checkbox checked={isSelected(pod.name)} />
              </TableCell>
              <TableCell>{pod.name}</TableCell>
              <TableCell>{pod.namespace}</TableCell>
              <TableCell>{pod.ip}</TableCell>
              <TableCell>{pod.state}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default ScopePodsTable
