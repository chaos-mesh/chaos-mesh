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
import FingerprintIcon from '@mui/icons-material/Fingerprint'
import HelpOutlineIcon from '@mui/icons-material/HelpOutline'
import ScheduleIcon from '@mui/icons-material/Schedule'
import SearchIcon from '@mui/icons-material/Search'
import {
  Autocomplete,
  Box,
  ChipProps,
  CircularProgress,
  InputAdornment,
  Chip as MUIChip,
  TextField,
  Typography,
} from '@mui/material'
import { makeStyles } from '@mui/styles'
import _ from 'lodash'
import {
  getArchives,
  getArchivesSchedules,
  getArchivesWorkflows,
  getExperiments,
  getSchedules,
  getWorkflows,
} from 'openapi'
import { CoreWorkflowMeta, TypesArchive, TypesExperiment, TypesSchedule } from 'openapi/index.schemas'
import { useMemo, useState } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import Paper from '@ui/mui-extends/esm/Paper'
import Tooltip from '@ui/mui-extends/esm/Tooltip'

import i18n from 'components/T'

import { format } from 'lib/luxon'
import search from 'lib/search'

const Chip = (props: ChipProps) => <MUIChip {...props} variant="outlined" size="small" />

const useStyles = makeStyles((theme) => ({
  tooltip: {
    marginBottom: 0,
    paddingLeft: theme.spacing(3),
    '& > li': {
      marginTop: theme.spacing(1.5),
    },
  },
  chipContainer: {
    display: 'flex',
    flexWrap: 'wrap',
    '& > *': {
      margin: theme.spacing(0.75),
      marginLeft: 0,
    },
  },
}))

type OptionCategory = CoreWorkflowMeta | TypesSchedule | TypesExperiment | TypesArchive
type Option = OptionCategory & { is?: string }

const Search: React.FC = () => {
  const classes = useStyles()
  const navigate = useNavigate()
  const intl = useIntl()

  const [open, setOpen] = useState(false)
  const [options, setOptions] = useState<Option[]>([])
  const [noResult, setNoResult] = useState(true)
  const loading = open && options.length === 0

  const debounceExecSearch = useMemo(
    () =>
      _.debounce(async (s: string) => {
        setNoResult(false)
        setOpen(true)

        const [workflows, schedules, experiments, archives, archivedWorkflows, archivedSchedules] = [
          (await getWorkflows()).map((d) => ({
            ...d,
            is: 'workflow' as 'workflow',
            kind: 'Workflow',
          })),
          (await getSchedules()).map((d) => ({ ...d, is: 'schedule' })),
          (await getExperiments()).map((d) => ({ ...d, is: 'experiment' })),
          (await getArchives()).map((d) => ({ ...d, is: 'archive' })),
          (await getArchivesWorkflows()).map((d) => ({ ...d, is: 'archive' })),
          (await getArchivesSchedules()).map((d) => ({ ...d, is: 'archive' })),
        ]

        const result = search(
          { workflows, schedules, experiments, archives: [...archives, ...archivedWorkflows, ...archivedSchedules] },
          s
        )
        const newOptions = [...result.workflows, ...result.schedules, ...result.experiments, ...result.archives]

        setOptions(newOptions)
        if (newOptions.length === 0) {
          setNoResult(true)
        }
      }, 500),
    []
  )

  const groupBy = (option: Option) => i18n(`${option.is}s.title`, intl)
  const getOptionLabel = (option: Option) => option.name!
  const isOptionEqualToValue = (option: Option, value: Option) => option.uid === value.uid
  const filterOptions = (options: Option[]) => options

  const determineKind = (option: Option) => (option.is === 'workflow' ? 'Workflow' : (option as any).kind)
  const determineLink = (uuid: uuid, type: Option['is'], kind: string) => {
    let link = `/${type}s/${uuid}`

    switch (type) {
      case 'archive':
        switch (kind) {
          case 'Workflow':
          case 'Schedule':
            link = `${link}?kind=${kind.toLowerCase()}`
            break
          default:
            link = `${link}?kind=experiment`
            break
        }
        break
    }

    return link
  }

  const renderOption = (props: any, option: Option) => {
    const type = option.is

    const uuid = option.uid!
    const name = option.name
    const kind = determineKind(option)
    const time = option.created_at!

    const onClick = () => {
      navigate(determineLink(uuid, type, kind))
      setOpen(false)
    }

    return (
      <li {...props} onClick={onClick}>
        <Box>
          <Typography variant="subtitle1" gutterBottom>
            {name}
          </Typography>
          <div className={classes.chipContainer}>
            <Chip color="primary" icon={<FingerprintIcon />} label={_.truncate(uuid)} title={uuid} />
            <Chip label={kind} />
            <Chip icon={<ScheduleIcon />} label={format(time)} />
          </div>
        </Box>
      </li>
    )
  }

  const onChange = (_: any, value: Option, reason: string) => {
    if (reason === 'selectOption') {
      navigate(determineLink(value.uid!, value.is, determineKind(value!)))
    }
  }

  const onInputChange = (_: any, newVal: string, reason: string) => {
    if (newVal && reason !== 'reset') {
      debounceExecSearch(newVal)
    }
  }

  return (
    <Autocomplete
      sx={{ minWidth: 360 }}
      className="tutorial-search"
      size="small"
      open={open}
      onClose={() => setOpen(false)}
      loading={loading}
      loadingText={noResult ? i18n('search.result.noResult') : i18n('search.result.acquiring')}
      options={options}
      groupBy={groupBy}
      getOptionLabel={getOptionLabel}
      isOptionEqualToValue={isOptionEqualToValue}
      filterOptions={filterOptions}
      renderOption={renderOption}
      onChange={onChange}
      onInputChange={onInputChange}
      renderInput={(params) => (
        <TextField
          {...params}
          label={i18n('search.placeholder')}
          InputProps={{
            ...params.InputProps,
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
            endAdornment: (
              <>
                {loading && !noResult ? <CircularProgress color="inherit" size={15} /> : null}
                <InputAdornment position="end">
                  <Tooltip
                    title={
                      <Typography variant="body2" component="div">
                        {i18n('search.tip.title')}
                        <ul className={classes.tooltip}>
                          <li>{i18n('search.tip.namespace')}</li>
                          <li>{i18n('search.tip.kind')}</li>
                        </ul>
                      </Typography>
                    }
                  >
                    <HelpOutlineIcon />
                  </Tooltip>
                </InputAdornment>
              </>
            ),
          }}
        />
      )}
      PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
      disableClearable
    />
  )
}

export default Search
