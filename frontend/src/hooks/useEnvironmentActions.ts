import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAppDispatch } from '@/store';
import { environmentApi } from '@/api/environments';
import { removeEnvironment } from '@/store/slices/environmentSlice';
import { showSuccess, showError } from '@/store/slices/notificationSlice';
import { openConfirmDialog } from '@/store/slices/uiSlice';

export const useEnvironmentActions = () => {
  const dispatch = useAppDispatch();
  const queryClient = useQueryClient();

  const restartMutation = useMutation({
    mutationFn: ({ id, force }: { id: string; force?: boolean }) => 
      environmentApi.restart(id, force),
    onSuccess: () => {
      dispatch(showSuccess(`Environment restart initiated`));
    },
    onError: (error: any) => {
      dispatch(showError(`Failed to restart environment: ${error.message}`));
    },
  });

  const upgradeMutation = useMutation({
    mutationFn: ({ id, version }: { id: string; version: string }) => 
      environmentApi.upgrade(id, { version }),
    onSuccess: (_, { version }) => {
      dispatch(showSuccess(`Environment upgrade to version ${version} initiated`));
      queryClient.invalidateQueries({ queryKey: ['environments'] });
    },
    onError: (error: any) => {
      dispatch(showError(`Failed to upgrade environment: ${error.message}`));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => environmentApi.delete(id),
    onSuccess: (_, id) => {
      dispatch(removeEnvironment(id));
      dispatch(showSuccess('Environment deleted successfully'));
      queryClient.invalidateQueries({ queryKey: ['environments'] });
    },
    onError: (error: any) => {
      dispatch(showError(`Failed to delete environment: ${error.message}`));
    },
  });

  const restart = (id: string, force: boolean = false) => {
    if (force) {
      dispatch(openConfirmDialog({
        title: 'Force Restart Environment',
        message: 'Are you sure you want to force restart this environment? This may cause data loss.',
        onConfirm: () => restartMutation.mutate({ id, force }),
      }));
    } else {
      restartMutation.mutate({ id, force });
    }
  };

  const upgrade = (id: string, version: string) => {
    dispatch(openConfirmDialog({
      title: 'Upgrade Environment',
      message: `Are you sure you want to upgrade this environment to version ${version}?`,
      onConfirm: () => upgradeMutation.mutate({ id, version }),
    }));
  };

  const deleteEnvironment = (id: string) => {
    dispatch(openConfirmDialog({
      title: 'Delete Environment',
      message: 'Are you sure you want to delete this environment? This action cannot be undone.',
      onConfirm: () => deleteMutation.mutate(id),
    }));
  };

  return {
    restart,
    upgrade,
    deleteEnvironment,
    isRestarting: restartMutation.isPending,
    isUpgrading: upgradeMutation.isPending,
    isDeleting: deleteMutation.isPending,
  };
};
