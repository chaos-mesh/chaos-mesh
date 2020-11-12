import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@material-ui/core'

import React from 'react'
import T from 'components/T'

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
      aria-labelledby="dialog-title"
      aria-describedby="dialog-description"
      PaperProps={{ style: { minWidth: 300 } }}
    >
      <DialogTitle id="dialog-title">{title}</DialogTitle>
      <DialogContent>
        <DialogContentText id="dialog-description">{description}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{T('common.cancel')}</Button>
        <Button color="primary" autoFocus onClick={_handleConfirm}>
          {T('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default ConfirmDialog
