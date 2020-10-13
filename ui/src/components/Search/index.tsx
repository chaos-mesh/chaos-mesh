import { GlobalSearchData, SearchPath, searchGlobal } from 'lib/search'
import { Grid, InputAdornment, Paper, TextField, Typography } from '@material-ui/core'
import { Link, LinkProps } from 'react-router-dom'
import ListItem, { ListItemProps } from '@material-ui/core/ListItem'
import React, { ReactNode, useCallback, useEffect, useState } from 'react'
import { createStyles, makeStyles } from '@material-ui/core/styles'

import { Archive } from 'api/archives.type'
import { Event } from 'api/events.type'
import { Experiment } from 'api/experiments.type'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import List from '@material-ui/core/List'
import ListItemText from '@material-ui/core/ListItemText'
import ListSubheader from '@material-ui/core/ListSubheader'
import Loading from 'components/Loading'
import SearchIcon from '@material-ui/icons/Search'
import Tooltip from 'components/Tooltip'
import _debounce from 'lodash.debounce'
import api from 'api'
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
      background: '#f5f6f7',
      maxHeight: '30rem',
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
  const hasExperiment = searchPath.some((path) => {
    return Object.keys(path)[0] === 'Experiment'
  })
  const hasUUID = searchPath.some((path) => {
    return Object.keys(path)[0] === 'UUID'
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
                    {hasExperiment || (
                      <ListItemText
                        primary={'Experiment'}
                        secondary={
                          category === 'events'
                            ? ((res as any) as Event).experiment
                            : ((res as any) as Experiment & Archive).name
                        }
                      />
                    )}

                    {hasUUID || category === 'events' || (
                      <ListItemText primary={'UUID'} secondary={((res as any) as Experiment & Archive).uid} />
                    )}
                    {category !== 'events' || (
                      <ListItemText primary={'Start Time'} secondary={((res as any) as Event).start_time} />
                    )}
                    {searchPath.map((path) => {
                      return (
                        <ListItemText
                          primary={Object.keys(path).filter((key) => key !== 'value')[0]}
                          secondary={
                            <HighLightText text={path.value}>
                              {path[Object.keys(path).filter((key) => key !== 'value')[0]]}
                            </HighLightText>
                          }
                          key={Object.keys(path).filter((key) => key !== 'value')[0]}
                        />
                      )
                    })}
                  </ListItemLink>
                </Paper>
              }
            </Grid>
          )
        })
        .slice(0, 5)}
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
  const [focus, setFocus] = useState(false)
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
    if (!focus) setShowSearchResult(false)
  }, [focus, globalSearchData])

  return (
    <div className={classes.searchContainer}>
      <TextField
        margin="dense"
        placeholder={intl.formatMessage({ id: 'common.search' })}
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
                    The following search syntax can help to locate the events quickly:
                    <ul style={{ marginBottom: 0, paddingLeft: '1rem' }}>
                      <li>namespace:default xxx will search for events with namespace default</li>
                      <li>
                        kind:NetworkChaos xxx will search for events with kind NetworkChaos, you can also type kind:net
                        because the search is fuzzy
                      </li>
                      <li>pod:echoserver-774cdcc8b6-nrm65 will search for events by affected pod</li>
                      <li>ip:172.17.0.6 is similar to pod:xxx, filter by pod IP</li>
                      <li>uuid:2f79a4d6-1952-45b5-b2d5-ce715823c7a7 will search for events by experimental uuid</li>
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
        onChange={handleSearchChange}
        onFocus={() => setFocus(true)}
      />
      {showSearchResult && (
        <Paper elevation={0} className={classes.searchResultContainer}>
          {loading ? (
            <Loading></Loading>
          ) : (
            <List component="nav" aria-label="search-result">
              {Object.keys(searchResult || {}).map((key, index) => {
                return (
                  <React.Fragment key={key}>
                    <ListSubheader disableSticky={true} style={{ fontSize: '22px', padding: 0 }}>
                      {key}
                    </ListSubheader>
                    {((searchResult as GlobalSearchData)[key as keyof GlobalSearchData] as Array<
                      GlobalSearchData[keyof GlobalSearchData][number]
                    >).length !== 0 ? (
                      <SearchResultForOneCate
                        category={key as keyof GlobalSearchData}
                        result={(searchResult as GlobalSearchData)[key as keyof GlobalSearchData]}
                        searchPath={searchPath![key as keyof GlobalSearchData]}
                      />
                    ) : (
                      <Paper variant="outlined">
                        <ListItemLink to="/">
                          <ListItemText primary="No Result" />
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
