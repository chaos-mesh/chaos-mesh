import { useStoreDispatch, useStoreSelector } from 'store'

import { IconButton } from '@material-ui/core'
import Modal from '@material-ui/core/Modal'
import Paper from 'components-mui/Paper'
import React from 'react'
import Search from 'components/Search'
import SearchIcon from '@material-ui/icons/Search'
import { makeStyles } from '@material-ui/core/styles'
import { setSearchModalOpen } from 'slices/globalStatus'

const useStyles = makeStyles((theme) => ({
  modalPaperWrapper: {
    maxWidth: '40%',
    margin: '0 auto',
    marginTop: theme.spacing(9),
    [theme.breakpoints.down('md')]: {
      maxWidth: '80%',
    },
  },
  modalPaper: {
    padding: theme.spacing(3),
  },
}))

const SearchTrigger: React.FC = () => {
  const classes = useStyles()

  const { searchModalOpen } = useStoreSelector((state) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const handleOpen = () => dispatch(setSearchModalOpen(true))
  const handleClose = () => dispatch(setSearchModalOpen(false))

  return (
    <>
      <IconButton className="nav-search" color="inherit" aria-label="Search" onClick={handleOpen}>
        <SearchIcon />
      </IconButton>
      <Modal open={searchModalOpen} onClose={handleClose}>
        <div className={classes.modalPaperWrapper}>
          <Paper className={classes.modalPaper}>
            <Search />
          </Paper>
        </div>
      </Modal>
    </>
  )
}

export default SearchTrigger
