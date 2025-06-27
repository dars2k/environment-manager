import { useEffect, useRef } from 'react';
import { useSnackbar } from 'notistack';
import { useAppSelector, useAppDispatch } from '@/store';
import { removeNotification } from '@/store/slices/notificationSlice';

export const useNotifications = () => {
  const { enqueueSnackbar } = useSnackbar();
  const dispatch = useAppDispatch();
  const notifications = useAppSelector(state => state.notifications.notifications);
  const displayedNotifications = useRef<Set<string>>(new Set());

  useEffect(() => {
    notifications.forEach((notification) => {
      // Skip if notification has already been displayed
      if (displayedNotifications.current.has(notification.id)) {
        return;
      }

      // Mark notification as displayed
      displayedNotifications.current.add(notification.id);

      enqueueSnackbar(notification.message, {
        variant: notification.type,
        autoHideDuration: notification.duration || 5000,
        onExited: () => {
          dispatch(removeNotification(notification.id));
          // Remove from displayed set when notification is removed
          displayedNotifications.current.delete(notification.id);
        },
      });
    });
  }, [notifications, enqueueSnackbar, dispatch]);
};
