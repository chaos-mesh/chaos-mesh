import { Autocomplete, Box, Chip, CircularProgress, InputAdornment, TextField, Typography } from '@material-ui/core'
import React, { useMemo, useState } from 'react'

import { Archive } from 'api/archives.type'
import { Event } from 'api/events.type'
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

type Option = Experiment | Event | Archive

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

        const [experiments, events, archives] = (
          await Promise.all([api.experiments.experiments(), api.events.events(), api.archives.archives()])
        ).map((d) => d.data)

        const result = search({ experiments, events, archives } as any, s)
        result.events = result.events.reverse().slice(0, 5)
        const newOptions = [...result.experiments, ...result.events, ...result.archives]

        setOptions(newOptions)
        if (newOptions.length === 0) {
          setNoResult(true)
        }
      }, 500),
    []
  )

  const filterOptions = (options: any) => options

  const groupBy = (option: Option) =>
    (option as Experiment).status
      ? intl.formatMessage({ id: 'experiments.title' })
      : (option as Event).name
      ? intl.formatMessage({ id: 'events.title' })
      : intl.formatMessage({ id: 'archives.title' })

  const getOptionLabel = (option: Option) => (option as Event).name || (option as any).name

  const renderOption = (_: any, option: Option) => {
    const type = (option as Experiment).status ? 'experiment' : (option as Event).name ? 'event' : 'archive'
    let link = ''
    let name = ''
    let uid = (option as Experiment).uid
    let kind = (option as Experiment).kind
    let time = ''

    switch (type) {
      case 'experiment':
      case 'archive':
        link = `/${type}s/${(option as Experiment).uid}`
        name = (option as Experiment).name
        time = (option as Experiment).created
        break
      case 'event':
        link = `/${type}s?event_id=${(option as Event).id}`
        name = (option as Event).name
        time = (option as Event).created_at
        break
      default:
        break
    }

    const onClick = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()

      history.push(link)
      setOpen(false)
    }

    return (
      <Box onClick={onClick}>
        <Typography gutterBottom>{name}</Typography>
        <Box className={classes.chipContainer}>
          {type !== 'event' ? (
            <Chip
              variant="outlined"
              color="primary"
              size="small"
              icon={<FingerprintIcon />}
              label={truncate(uid)}
              title={uid}
            />
          ) : (
            <Chip
              variant="outlined"
              color="primary"
              size="small"
              icon={<FingerprintIcon />}
              label={(option as Event).id}
            />
          )}
          <Chip variant="outlined" size="small" label={kind} />
          {type !== 'archive' && <Chip variant="outlined" size="small" icon={<ScheduleIcon />} label={format(time)} />}
        </Box>
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
      filterOptions={filterOptions}
      groupBy={groupBy}
      getOptionLabel={getOptionLabel}
      renderOption={renderOption}
      onInputChange={onInputChange}
      renderInput={(params) => (
        <TextField
          {...params}
          variant="outlined"
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
      PaperComponent={(props) => <Paper {...props} padding={0} />}
    />
  )
}

export default Search
