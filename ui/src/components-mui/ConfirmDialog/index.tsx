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
    <Dialog
      open={open}
      onClose={close}
      aria-labelledby="dialog-title"
      aria-describedby="dialog-description"
      PaperProps={{ style: { minWidth: 300 } }}
      {...dialogProps}
    >
      <DialogTitle id="dialog-title">{title}</DialogTitle>
      {children ? (
        <DialogContent>{children}</DialogContent>
      ) : description ? (
        <DialogContent>
          <DialogContentText id="dialog-description">{description}</DialogContentText>
        </DialogContent>
      ) : null}
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

export default ConfirmDialog
