import React, { useEffect } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { IconButton } from '@material-ui/core'
import Modal from '@material-ui/core/Modal'
import Paper from 'components-mui/Paper'
import { RootState } from 'store'
import Search from 'components/Search'
import SearchIcon from '@material-ui/icons/Search'
import { setSearchModalOpen } from 'slices/globalStatus'
import store from 'store'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) => {
  return createStyles({
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
    const keyMap: Record<string, boolean> = {}
    const keyDownHandler = (e: KeyboardEvent) => {
      keyMap[e.key] = true

      // In some cases, such as pressing multiple keys almost at the same time, the browser won't fire the keyup event repeatedly.
      if ((keyMap['Meta'] && keyMap['p']) || (keyMap['Control'] && keyMap['p'])) {
        e.preventDefault()

        handleOpen()

        delete keyMap['Meta']
        delete keyMap['Control']
        delete keyMap['p']
      }
    }

    document.addEventListener('keydown', keyDownHandler)
    return () => {
      document.removeEventListener('keydown', keyDownHandler)
    }
  }, [])

  return (
    <>
      <IconButton color="inherit" aria-label="Search" onClick={handleOpen}>
        <SearchIcon />
      </IconButton>
      <Modal open={searchModalOpen} onClose={handleClose}>
        <Paper elevation={3} className={classes.searchModal}>
          <Search />
        </Paper>
      </Modal>
    </>
  )
}

export default SearchTrigger
