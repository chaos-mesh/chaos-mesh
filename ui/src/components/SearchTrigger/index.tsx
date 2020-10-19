import { Button, Paper } from '@material-ui/core'
import React, { useEffect } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import Modal from '@material-ui/core/Modal'
import { RootState } from 'store'
import Search from 'components/Search'
import SearchIcon from '@material-ui/icons/Search'
import T from 'components/T'
import { setSearchModalOpen } from 'slices/globalStatus'
import store from 'store'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) => {
  return createStyles({
    searchTrigger: {
      borderRadius: '4px',
      cursor: 'pointer',
      whiteSpace: 'nowrap',
    },
    searchModal: {
      [theme.breakpoints.down('md')]: {
        maxWidth: '80%',
      },
      position: 'relative',
      maxWidth: '40%',
      margin: `${theme.spacing(15)} auto auto`,
      padding: theme.spacing(3),
      overflowY: 'hidden',
      outline: 0,
    },
  })
})

const SearchTrigger: React.FC = () => {
  const classes = useStyles()

  const handleOpen = () => {
    store.dispatch(setSearchModalOpen(true))
  }

  const handleClose = () => {
    store.dispatch(setSearchModalOpen(false))
  }

  const searchModalOpen = useSelector((state: RootState) => state.globalStatus.searchModalOpen)

  useEffect(() => {
    if (!searchModalOpen) handleClose()
  }, [searchModalOpen])

  useEffect(() => {
    const keyMap: { [index: string]: boolean } = {}
    const keyDownHandler = (e: KeyboardEvent) => {
      keyMap[e.code] = true
      if (keyMap['ControlLeft'] && keyMap['KeyP']) {
        handleOpen()
      }
    }
    const keyUpHandler = (e: KeyboardEvent) => {
      keyMap[e.code] = false
    }
    document.addEventListener('keydown', keyDownHandler)
    document.addEventListener('keyup', keyUpHandler)
    return () => {
      document.removeEventListener('keydown', keyDownHandler)
      document.removeEventListener('keyup', keyUpHandler)
    }
  }, [])

  return (
    <>
      <Button
        variant="outlined"
        className={classes.searchTrigger}
        startIcon={<SearchIcon color="primary" />}
        onClick={handleOpen}
      >
        {T('search.placeholder')}
      </Button>
      <Modal open={searchModalOpen} onClose={handleClose}>
        <Paper elevation={3} className={classes.searchModal}>
          <Search></Search>
        </Paper>
      </Modal>
    </>
  )
}

export default SearchTrigger
