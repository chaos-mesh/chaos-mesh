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
import ScheduleIcon from '@material-ui/icons/Schedule'
import SearchIcon from '@material-ui/icons/Search'
import T from 'components/T'
import Tooltip from 'components-mui/Tooltip'
import _debounce from 'lodash.debounce'
import api from 'api'
import clsx from 'clsx'
import { format } from 'lib/luxon'
import { makeStyles } from '@material-ui/styles'
import search from 'lib/search'
import { truncate } from 'lib/utils'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

const Chip = (props: ChipProps) => <MUIChip {...props} variant="outlined" size="small" />

const useStyles = makeStyles((theme) => ({
  search: {
    minWidth: 360,
    '& .MuiInputBase-root': {
      paddingLeft: '9px !important',
      paddingRight: '15px !important',
    },
  },
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

type Option = Experiment | Archive

const Search: React.FC = () => {
  const classes = useStyles()
  const intl = useIntl()
  const history = useHistory()

  const [open, setOpen] = useState(false)
  const [options, setOptions] = useState<Option[]>([])
  const [noResult, setNoResult] = useState(true)
  const loading = open && options.length === 0

  const debounceExecSearch = useMemo(
    () =>
      _debounce(async (s: string) => {
        setNoResult(false)
        setOpen(true)

        const [experiments, archives] = [
          (await api.experiments.experiments()).data.map((d) => ({ ...d, is: 'experiment' as 'experiment' })),
          (await api.archives.archives()).data.map((d) => ({ ...d, is: 'archive' as 'archive' })),
        ]

        const result = search({ experiments, archives }, s)
        const newOptions = [...result.experiments, ...result.archives]

        setOptions(newOptions)
        if (newOptions.length === 0) {
          setNoResult(true)
        }
      }, 500),
    []
  )

  const groupBy = (option: Option) =>
    (option.is === 'experiment' ? T('experiments.title', intl) : T('archives.title', intl)) as string

  const getOptionLabel = (option: Option) => option.name

  const renderOption = (_: any, option: Option) => {
    const type = option.is

    let uid = (option as Experiment).uid
    let name = option.name
    const kind = (option as Experiment).kind
    const time = option.created_at
    let link = ''

    switch (type) {
      case 'experiment':
        link = `/${type}s/${(option as Experiment).uid}`
        break
      case 'archive':
      default:
        break
    }

    const onClick = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()

      history.push(link)
      setOpen(false)
    }

    return (
      <Box pl={6} pb={3} onClick={onClick}>
        <Typography variant="subtitle1" gutterBottom>
          {name}
        </Typography>
        <div className={classes.chipContainer}>
          <Chip color="primary" icon={<FingerprintIcon />} label={truncate(uid)} title={uid} />
          <Chip label={kind} />
          {type !== 'archive' && <Chip icon={<ScheduleIcon />} label={format(time)} />}
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
      className={clsx(classes.search, 'nav-search')}
      freeSolo
      open={open}
      onClose={() => setOpen(false)}
      loading={loading}
      loadingText={noResult ? T('search.result.noResult') : T('search.result.acquiring')}
      options={options}
      groupBy={groupBy}
      getOptionLabel={getOptionLabel}
      renderOption={renderOption}
      onInputChange={onInputChange}
      renderInput={(params) => (
        <TextField
          {...params}
          size="small"
          label={T('search.placeholder')}
          aria-label="Search"
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
