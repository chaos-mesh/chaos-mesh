import { Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import { getIn, useFormikContext } from 'formik'

import PaperContainer from 'components-mui/PaperContainer'
import T from 'components/T'
import { setAlert } from 'slices/globalStatus'
import { useMemo } from 'react'
import { useStoreDispatch } from 'store'

interface ScopePodsTableProps {
  scope?: string
  pods: any[]
}

const ScopePodsTable: React.FC<ScopePodsTableProps> = ({ scope = 'scope', pods }) => {
  const originalPods = useMemo(
    () => pods.map((d) => `${d.namespace}:${d.name}`).reduce<string[]>((acc, d) => acc.concat(d), []),
    [pods]
  )
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
            <TableCell>{T('common.name')}</TableCell>
            <TableCell>{T('k8s.namespace')}</TableCell>
            <TableCell>{T('common.ip')}</TableCell>
            <TableCell>{T('common.status')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {pods.map((pod) => {
            const key = `${pod.namespace}:${pod.name}`

            return (
              <TableRow key={key} onClick={handleSelect(key)}>
                <TableCell padding="checkbox">
                  <Checkbox checked={isSelected(key)} />
                </TableCell>
                <TableCell>{pod.name}</TableCell>
                <TableCell>{pod.namespace}</TableCell>
                <TableCell>{pod.ip}</TableCell>
                <TableCell>{pod.state}</TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default ScopePodsTable
