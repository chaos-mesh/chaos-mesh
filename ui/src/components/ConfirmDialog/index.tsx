import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@material-ui/core'

import React from 'react'

interface ConfirmDialogProps {
  open: boolean
  setOpen: (open: boolean) => void
  title: string
  description: string
  handleConfirm: () => void
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({ open, setOpen, title, description, handleConfirm }) => {
  const handleClose = () => setOpen(false)

  const _handleConfirm = () => {
    handleConfirm()
    handleClose()
  }

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      aria-labelledby="alert-dialog-title"
      aria-describedby="alert-dialog-description"
    >
      <DialogTitle id="alert-dialog-title">{title}</DialogTitle>
      <DialogContent>
        <DialogContentText id="alert-dialog-description">{description}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" onClick={handleClose}>
          Cancel
        </Button>
        <Button variant="contained" color="primary" autoFocus onClick={_handleConfirm}>
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default ConfirmDialog
