/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogProps,
  DialogTitle,
} from '@mui/material'

interface ConfirmDialogProps {
  open: boolean
  close?: () => void
  title: React.ReactNode
  description?: React.ReactNode
  cancelText?: string
  confirmText?: string
  onConfirm?: () => void
  dialogProps?: Omit<DialogProps, 'open'>
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  close,
  title,
  description,
  cancelText = 'Cancel',
  confirmText = 'Confirm',
  onConfirm,
  children,
  dialogProps,
}) => {
  const handleConfirm = () => {
    typeof onConfirm === 'function' && onConfirm()
    typeof close === 'function' && close()
  }

  return (
    <Dialog open={open} onClose={close} {...dialogProps}>
      <DialogTitle sx={{ p: 4 }}>{title}</DialogTitle>
      {(children || description) && (
        <DialogContent sx={{ p: 4 }}>
          {description ? <DialogContentText>{description}</DialogContentText> : children}
        </DialogContent>
      )}

      {onConfirm && (
        <DialogActions>
          <Button color="secondary" onClick={close}>
            {cancelText}
          </Button>
          <Button autoFocus onClick={handleConfirm}>
            {confirmText}
          </Button>
        </DialogActions>
      )}
    </Dialog>
  )
}

export default ConfirmDialog
