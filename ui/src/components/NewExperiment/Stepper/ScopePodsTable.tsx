import { Checkbox, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import React, { useEffect, useMemo, useRef, useState } from 'react'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import { Experiment } from 'components/NewExperiment/types'
import { useFormikContext } from 'formik'
import { useStoreDispatch } from 'store'

const PaperOutlined: React.FC = ({ children }) => <Paper variant="outlined">{children}</Paper>

interface ScopePodsTableProps {
  scope?: string
  pods: any[]
}

const ScopePodsTable: React.FC<ScopePodsTableProps> = ({ scope = 'scope', pods }) => {
  const { setFieldValue } = useFormikContext<Experiment>()

  const dispatch = useStoreDispatch()

  const formPods = useMemo(() => pods.map((d) => d.name).reduce((acc, d) => acc.concat(d), []), [pods])
  const podsCount = formPods.length
  const podsCountRef = useRef(podsCount)

  const selectedRef = useRef(formPods)
  const [selected, _setSelected] = useState<string[]>(selectedRef.current)
  const setSelected = (newVal: string[]) => {
    selectedRef.current = newVal
    _setSelected(selectedRef.current)
  }

  const numSelected = selected.length
  const isSelected = (name: string) => selected.indexOf(name) !== -1

  useEffect(() => {
    setSelected(formPods)
    podsCountRef.current = formPods.length
  }, [formPods])

  useEffect(
    () => () => {
      if (selectedRef.current.length === podsCountRef.current) {
        return
      }

      setFieldValue(
        `${scope}.pods`,
        pods
          .filter((pod) => selectedRef.current.indexOf(pod.name) !== -1)
          .reduce((acc, d) => {
            if (acc.hasOwnProperty(d.namespace)) {
              acc[d.namespace].push(d.name)
            } else {
              acc[d.namespace] = [d.name]
            }

            return acc
          }, {})
      )
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked) {
      const newSelecteds = pods.map((p) => p.name)

      setSelected(newSelecteds)

      return
    }

    setSelected([])
  }

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

    setSelected(newSelected)
  }

  return (
    <TableContainer component={PaperOutlined}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                indeterminate={numSelected > 0 && numSelected < podsCount}
                checked={podsCount > 0 && numSelected === podsCount}
                onChange={handleSelectAll}
              />
            </TableCell>
            <TableCell>Name</TableCell>
            <TableCell>Namespace</TableCell>
            <TableCell>IP</TableCell>
            <TableCell>State</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {pods.map((pod) => (
            <TableRow key={pod.name} onClick={handleSelect(pod.name)}>
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
