import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
} from '@mui/material';

import { useAppSelector, useAppDispatch } from '@/store';
import { closeConfirmDialog } from '@/store/slices/uiSlice';

export const ConfirmDialog: React.FC = () => {
  const dispatch = useAppDispatch();
  const { confirmDialogOpen, confirmDialog } = useAppSelector(state => state.ui);

  const handleClose = () => {
    dispatch(closeConfirmDialog());
  };

  const handleConfirm = () => {
    if (confirmDialog?.onConfirm) {
      confirmDialog.onConfirm();
    }
    handleClose();
  };

  if (!confirmDialog) {
    return null;
  }

  return (
    <Dialog
      open={confirmDialogOpen}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle>{confirmDialog.title}</DialogTitle>
      <DialogContent>
        <Typography>{confirmDialog.message}</Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button onClick={handleConfirm} variant="contained" color="error">
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  );
};
