import { Button, Paper } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import Modal from '@material-ui/core/Modal'
import { RootState } from 'store'
import Search from 'components/Search'
import SearchIcon from '@material-ui/icons/Search'
import T from 'components/T'
import { setSearchModalOpen } from 'slices/globalStatus'
import store from 'store'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    searchTrigger: {
      borderRadius: '4px',
      cursor: 'pointer',
      whiteSpace: 'nowrap',
    },
    searchModal: {
      position: 'relative',
      maxWidth: '40%',
      minHeight: '206px',
      margin: '3.75rem auto auto',
      padding: 12,
      outline: 0,
    },
  })
)

const SearchTrigger: React.FC = () => {
  const classes = useStyles()
  const [open, setOpen] = useState(false)

  const handleOpen = () => {
    setOpen(true)
    store.dispatch(setSearchModalOpen(true))
  }

  const handleClose = () => {
    setOpen(false)
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
      <Modal open={open} onClose={handleClose}>
        <Paper elevation={3} className={classes.searchModal}>
          <Search></Search>
        </Paper>
      </Modal>
    </>
  )
}

export default SearchTrigger
