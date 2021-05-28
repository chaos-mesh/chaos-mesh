import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogProps,
  DialogTitle,
} from '@material-ui/core'
import React, { useImperativeHandle, useState } from 'react'

import T from 'components/T'

export interface ConfirmDialogHandles {
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

interface ConfirmDialogProps {
  open: boolean
  close?: () => void
  title: string | JSX.Element
  description?: string
  onConfirm?: () => void
  dialogProps?: Omit<DialogProps, 'open'>
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  close,
  title,
  description,
  onConfirm,
  children,
  dialogProps,
}) => {
  const handleConfirm = () => {
    typeof onConfirm === 'function' && onConfirm()
    typeof close === 'function' && close()
  }

  return (
    <Dialog
      open={open}
      onClose={close}
      aria-labelledby="dialog-title"
      aria-describedby="dialog-description"
      PaperProps={{ style: { minWidth: 300 } }}
      {...dialogProps}
    >
      <DialogTitle id="dialog-title">{title}</DialogTitle>
      <DialogContent>
        {children ? children : <DialogContentText id="dialog-description">{description}</DialogContentText>}
      </DialogContent>
      {onConfirm && (
        <DialogActions>
          <Button size="small" onClick={close}>
            {T('common.cancel')}
          </Button>
          <Button variant="contained" color="primary" size="small" autoFocus disableFocusRipple onClick={handleConfirm}>
            {T('common.confirm')}
          </Button>
        </DialogActions>
      )}
    </Dialog>
  )
}

export default React.forwardRef(ConfirmDialog)
