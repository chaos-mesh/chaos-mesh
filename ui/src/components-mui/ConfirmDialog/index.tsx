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
} from '@material-ui/core'

import T from 'components/T'

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
    <Dialog open={open} onClose={close} PaperProps={{ sx: { minWidth: 300 } }} {...dialogProps}>
      <DialogTitle sx={{ p: 4.5 }}>{title}</DialogTitle>
      <DialogContent sx={{ p: 4.5 }}>
        {children ? children : description ? <DialogContentText>{description}</DialogContentText> : null}
      </DialogContent>

      {onConfirm && (
        <DialogActions sx={{ p: 4.5 }}>
          <Button onClick={close}>{T('common.cancel')}</Button>
          <Button variant="contained" color="primary" autoFocus disableFocusRipple onClick={handleConfirm}>
            {T('common.confirm')}
          </Button>
        </DialogActions>
      )}
    </Dialog>
  )
}

export default ConfirmDialog
