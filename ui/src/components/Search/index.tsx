import { GlobalSearchData, PropForKeyword, SearchPath, searchGlobal } from 'lib/search'
import { Grid, InputAdornment, Paper, TextField, Typography } from '@material-ui/core'
import { Link, LinkProps } from 'react-router-dom'
import ListItem, { ListItemProps } from '@material-ui/core/ListItem'
import React, { ReactNode, useCallback, useEffect, useState } from 'react'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import { dayComparator, format } from 'lib/dayjs'

import { Archive } from 'api/archives.type'
import { Event } from 'api/events.type'
import { Experiment } from 'api/experiments.type'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import List from '@material-ui/core/List'
import ListItemText from '@material-ui/core/ListItemText'
import ListSubheader from '@material-ui/core/ListSubheader'
import Loading from 'components/Loading'
import SearchIcon from '@material-ui/icons/Search'
import Separate from 'components/Separate'
import T from 'components/T'
import Tooltip from 'components/Tooltip'
import _debounce from 'lodash.debounce'
import api from 'api'
import { assumeType } from 'lib/utils'
import { setSearchModalOpen } from 'slices/globalStatus'
import store from 'store'
import { useIntl } from 'react-intl'

const useStyles = makeStyles(() =>
  createStyles({
    searchContainer: {
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'spaceBetween',
    },
    searchResultContainer: {
      marginTop: 10,
      maxHeight: '480px',
      overflow: 'scroll',
    },
  })
)

const ListItemLink: React.FC<ListItemProps & { to: string }> = (props) => {
  const { to, children, style } = props

  const renderLink = React.useMemo(
    () => React.forwardRef<any, Omit<LinkProps, 'to'>>((itemProps, ref) => <Link to={to} ref={ref} {...itemProps} />),
    [to]
  )

  return (
    <ListItem button component={renderLink} style={style} onClick={() => store.dispatch(setSearchModalOpen(false))}>
      <>{children}</>
    </ListItem>
  )
}

interface HighLightTextProps {
  text: string
  children: string
}

const HighLightText: React.FC<HighLightTextProps> = ({ children, text }) => {
  const matchRes = children.match(new RegExp(text, 'i'))
  const startIndex = matchRes!.index!
  const frontPart = children.slice(0, startIndex)
  const highLightPart = children.slice(startIndex, startIndex + text.length)
  const endPart = children.slice(startIndex + text.length)
  return (
    <>
      <span>{frontPart}</span>
      <span style={{ background: '#E7F0FB', color: '#143A78' }}>{highLightPart}</span>
      <span>{endPart}</span>
    </>
  )
}

interface SearchResultForOneCateProps<T extends 'events' | 'experiments' | 'archives'> {
  category: T
  searchPath: SearchPath[keyof SearchPath]
  result: T extends 'events' ? Event[] : T extends 'experiments' ? Experiment[] : Archive[]
}

const SearchResultForOneCate = function <T extends 'events' | 'experiments' | 'archives'>(
  props: SearchResultForOneCateProps<T> & { children?: ReactNode }
) {
  const { category, searchPath, result } = props

  if (category === 'events') {
    ;((result as unknown) as Event[]).sort((a, b) => {
      return dayComparator(a.start_time, b.start_time)
    })
  }

  type searchResultNames = PropForKeyword | 'start_time' | 'name'

  const nameMap: {
    [k in searchResultNames]: string
  } = {
    experiment: 'Experiment',
    uid: 'UUID',
    experiment_id: 'UUID',
    ip: 'IP',
    kind: 'Kind',
    pod: 'Pod',
    namespace: 'Namespace',
    start_time: 'Start Time',
    name: 'Experiment',
  }

  type RequiredHLItem = { name: searchResultNames; alternativeName?: string; isHighLighted: true; value: string }
  type RequiredNoHLItem = { name: searchResultNames; alternativeName?: string; isHighLighted: false }
  type RequiredItems = (RequiredHLItem | RequiredNoHLItem)[]

  const requiredItems: RequiredItems =
    category === 'events'
      ? [
          { name: 'experiment', isHighLighted: false },
          { name: 'start_time', isHighLighted: false },
        ]
      : category === 'experiments' || category === 'archives'
      ? [
          { name: 'name', alternativeName: 'experiment', isHighLighted: false },
          { name: 'uid', isHighLighted: false },
        ]
      : []

  requiredItems.forEach((item) => {
    const posInSearchPath = searchPath.findIndex((path) => {
      return path.name === item.name || (item.alternativeName && item.alternativeName === path.name)
    })
    if (posInSearchPath !== -1) {
      item.isHighLighted = true
      if (item.isHighLighted === true) {
        item.value = searchPath[posInSearchPath].value
      }
    }
  })

  return (
    <Grid container direction="column" justify="space-between" spacing={3}>
      {((result as unknown) as (T extends 'events' ? Event : T extends 'experiments' ? Experiment : Archive)[])
        .map((res) => {
          return (
            <Grid
              item
              key={
                category === 'events' ? ((res as unknown) as Event).id : ((res as unknown) as Experiment & Archive).uid
              }
            >
              {
                <Paper variant="outlined">
                  <ListItemLink
                    style={{
                      flexDirection: 'column',
                      alignItems: 'flex-start',
                    }}
                    to={
                      category === 'events'
                        ? `/events?event_id=${((res as unknown) as Event).id}`
                        : category === 'experiments'
                        ? `/experiments/${((res as unknown) as Experiment).uid}`
                        : `/archives/${((res as unknown) as Archive).uid}`
                    }
                  >
                    <ListItemText
                      style={{
                        wordBreak: 'break-all',
                      }}
                      primary={
                        <Separate separator={<span>&nbsp;|&nbsp;</span>}>
                          {searchPath.map((path) => {
                            return (
                              <React.Fragment key={path.name + path.value}>
                                <span>{T(`search.result.keywords.${nameMap[path.name]}`)}: </span>
                                <HighLightText text={path.value}>{path.matchedValue}</HighLightText>
                              </React.Fragment>
                            )
                          })}
                        </Separate>
                      }
                      secondary={
                        <Separate separator={<span>&nbsp;|&nbsp;</span>}>
                          {requiredItems.map((item) => {
                            return (
                              <React.Fragment key={item.name}>
                                <span>{T(`search.result.keywords.${nameMap[item.name]}`)}: </span>
                                {item.isHighLighted ? (
                                  <HighLightText text={item.value}>{(res as any)[item.name]}</HighLightText>
                                ) : (
                                  <span>
                                    {item.name === 'start_time'
                                      ? format((res as any)[item.name])
                                      : (res as any)[item.name]}
                                  </span>
                                )}
                              </React.Fragment>
                            )
                          })}
                        </Separate>
                      }
                    />
                  </ListItemLink>
                </Paper>
              }
            </Grid>
          )
        })
        .slice(0, 3)}
    </Grid>
  )
}

const Search: React.FC = () => {
  const classes = useStyles()
  const intl = useIntl()
  const [search, setSearch] = useState('')
  const [showSearchResult, setShowSearchResult] = useState(false)
  const [searchResult, setSearchResult] = useState<GlobalSearchData | {}>()
  const [searchPath, setSearchPath] = useState<SearchPath>()
  const [loading, setLoading] = useState(false)
  const [isEmptySearch, setIsEmptySearch] = useState(true)

  const debounceSetSearch = useCallback(_debounce(setSearch, 500), [])
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const search = e.target.value
    search.length === 0 ? setIsEmptySearch(true) : setIsEmptySearch(false)
    debounceSetSearch(e.target.value)
  }

  const [globalSearchData, setGlobalSearchData] = useState<GlobalSearchData>()

  const fetchExperiments = () => {
    return api.experiments.experiments()
  }

  const fetchEvents = () => {
    return api.events.events()
  }

  const fetchArchives = () => {
    return api.archives.archives()
  }

  const fetchAll = () => {
    Promise.all([fetchExperiments(), fetchEvents(), fetchArchives()])
      .then((data) => {
        setGlobalSearchData({
          experiments: data[0].data,
          events: data[1].data,
          archives: data[2].data,
        })
      })
      .catch(console.log)
  }

  useEffect(fetchAll, [search])

  useEffect(() => {
    if (globalSearchData) {
      const { result, searchPath } = searchGlobal(globalSearchData, search)
      setSearchResult(result)
      setSearchPath(searchPath)
      isEmptySearch ? setShowSearchResult(false) : setShowSearchResult(true)
    }
    setLoading(false)
    // eslint-disable-next-line
  }, [search])

  useEffect(() => {
    globalSearchData ? setLoading(false) : setLoading(true)
  }, [globalSearchData])

  return (
    <div className={classes.searchContainer}>
      <TextField
        margin="dense"
        placeholder={intl.formatMessage({ id: 'search.placeholder' })}
        disabled={!globalSearchData}
        variant="outlined"
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon color="primary" />
            </InputAdornment>
          ),
          endAdornment: (
            <InputAdornment position="end">
              <Tooltip
                title={
                  <Typography variant="body2">
                    {T('search.tip.title')}
                    <ul style={{ marginBottom: 0, paddingLeft: '1rem' }}>
                      <li>{T('search.tip.namespace')}</li>
                      <li>{T('search.tip.kind')}</li>
                      <li>{T('search.tip.pod')}</li>
                      <li>{T('search.tip.ip')}</li>
                      <li>{T('search.tip.uuid')}</li>
                    </ul>
                  </Typography>
                }
                style={{ verticalAlign: 'sub' }}
                arrow
                interactive
              >
                <HelpOutlineIcon fontSize="small" />
              </Tooltip>
            </InputAdornment>
          ),
        }}
        inputProps={{
          style: { paddingTop: 8, paddingBottom: 8 },
        }}
        inputRef={(input) => input && input.focus()}
        onChange={handleSearchChange}
      />
      {showSearchResult && (
        <Paper elevation={0} className={classes.searchResultContainer}>
          {loading ? (
            <Loading></Loading>
          ) : (
            <List component="nav" aria-label="search-result">
              {Object.keys(searchResult || {}).map((key) => {
                assumeType<GlobalSearchData>(searchResult)
                assumeType<keyof GlobalSearchData>(key)
                return (
                  <React.Fragment key={key}>
                    <ListSubheader disableSticky={true} style={{ fontSize: '22px', padding: 0 }}>
                      {T(`search.result.category.${key}`)}
                    </ListSubheader>
                    {searchResult[key].length !== 0 ? (
                      <SearchResultForOneCate category={key} result={searchResult[key]} searchPath={searchPath![key]} />
                    ) : (
                      <Paper variant="outlined">
                        <ListItemLink to="/">
                          <ListItemText primary={T('search.result.noResult')} />
                        </ListItemLink>
                      </Paper>
                    )}
                  </React.Fragment>
                )
              })}
            </List>
          )}
        </Paper>
      )}
    </div>
  )
}

export default Search
