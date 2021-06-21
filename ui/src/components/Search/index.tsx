import {
  Autocomplete,
  Box,
  ChipProps,
  CircularProgress,
  InputAdornment,
  Chip as MUIChip,
  TextField,
  Typography,
} from '@material-ui/core'
import { useMemo, useState } from 'react'

import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'
import FingerprintIcon from '@material-ui/icons/Fingerprint'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import Paper from 'components-mui/Paper'
import { Schedule } from 'api/schedules.type'
import ScheduleIcon from '@material-ui/icons/Schedule'
import SearchIcon from '@material-ui/icons/Search'
import T from 'components/T'
import Tooltip from 'components-mui/Tooltip'
import { Workflow } from 'api/workflows.type'
import _debounce from 'lodash.debounce'
import api from 'api'
import { format } from 'lib/luxon'
import { makeStyles } from '@material-ui/styles'
import search from 'lib/search'
import { truncate } from 'lib/utils'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

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

type Option = Workflow | Schedule | Experiment | Archive

const Search: React.FC = () => {
  const classes = useStyles()
  const history = useHistory()
  const intl = useIntl()

  const [open, setOpen] = useState(false)
  const [options, setOptions] = useState<Option[]>([])
  const [noResult, setNoResult] = useState(true)
  const loading = open && options.length === 0

  const debounceExecSearch = useMemo(
    () =>
      _debounce(async (s: string) => {
        setNoResult(false)
        setOpen(true)

        const [workflows, schedules, experiments, archives, archivedWorkflows, archivedSchedules] = [
          (await api.workflows.workflows()).data.map((d) => ({
            ...d,
            is: 'workflow' as 'workflow',
            kind: 'Workflow',
          })),
          (await api.schedules.schedules()).data.map((d) => ({ ...d, is: 'schedule' as 'schedule' })),
          (await api.experiments.experiments()).data.map((d) => ({ ...d, is: 'experiment' as 'experiment' })),
          (await api.archives.archives()).data.map((d) => ({ ...d, is: 'archive' as 'archive' })),
          (await api.workflows.archives()).data.map((d) => ({ ...d, is: 'archive' as 'archive' })),
          (await api.schedules.archives()).data.map((d) => ({ ...d, is: 'archive' as 'archive' })),
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

  const groupBy = (option: Option) => T(`${option.is}s.title`, intl)
  const getOptionLabel = (option: Option) => option.name
  const filterOptions = (options: Option[]) => options

  const renderOption = (_: any, option: Option) => {
    const type = option.is

    let uid = option.uid
    let name = option.name
    const kind: string = type === 'workflow' ? 'Workflow' : (option as any).kind
    const time = option.created_at
    let link = `/${type}s/${option.uid}`

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

    const onClick = () => {
      history.replace(link)
      setOpen(false)
    }

    return (
      <Box key={uid} onClick={onClick} sx={{ pl: 6, py: 3, cursor: 'pointer', ':hover': { bgcolor: 'action.hover' } }}>
        <Typography variant="subtitle1" gutterBottom>
          {name}
        </Typography>
        <div className={classes.chipContainer}>
          <Chip color="primary" icon={<FingerprintIcon />} label={truncate(uid)} title={uid} />
          <Chip label={kind} />
          <Chip icon={<ScheduleIcon />} label={format(time)} />
        </div>
      </Box>
    )
  }

  const onInputChange = (_: any, newVal: string) => {
    if (newVal) {
      debounceExecSearch(newVal)
    }
  }

  return (
    <Autocomplete
      sx={{ minWidth: 360 }}
      className="tutorial-search"
      freeSolo
      open={open}
      onClose={() => setOpen(false)}
      loading={loading}
      loadingText={noResult ? T('search.result.noResult') : T('search.result.acquiring')}
      options={options}
      groupBy={groupBy}
      getOptionLabel={getOptionLabel}
      filterOptions={filterOptions}
      renderOption={renderOption}
      onInputChange={onInputChange}
      renderInput={(params) => (
        <TextField
          {...params}
          size="small"
          label={T('search.placeholder')}
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
                        {T('search.tip.title')}
                        <ul className={classes.tooltip}>
                          <li>{T('search.tip.namespace')}</li>
                          <li>{T('search.tip.kind')}</li>
                        </ul>
                      </Typography>
                    }
                  >
                    <HelpOutlineIcon fontSize="small" />
                  </Tooltip>
                </InputAdornment>
              </>
            ),
          }}
        />
      )}
      PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
    />
  )
}

export default Search
