import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@/test/test-utils';
import { ConfirmDialog } from '../ConfirmDialog';
import { createTestStore } from '@/test/test-utils';
import { openConfirmDialog } from '@/store/slices/uiSlice';

describe('ConfirmDialog', () => {
  it('renders nothing when confirmDialog is null', () => {
    const { container } = render(<ConfirmDialog />);
    // No dialog content rendered
    expect(container.firstChild).toBeNull();
  });

  it('renders dialog when confirmDialog state is set', () => {
    const store = createTestStore();
    store.dispatch(
      openConfirmDialog({
        title: 'Delete Item',
        message: 'Are you sure you want to delete this?',
        onConfirm: vi.fn(),
      })
    );
    render(<ConfirmDialog />, { store });
    expect(screen.getByText('Delete Item')).toBeInTheDocument();
    expect(screen.getByText('Are you sure you want to delete this?')).toBeInTheDocument();
  });

  it('shows Cancel and Confirm buttons', () => {
    const store = createTestStore();
    store.dispatch(
      openConfirmDialog({
        title: 'Test Dialog',
        message: 'Test message',
        onConfirm: vi.fn(),
      })
    );
    render(<ConfirmDialog />, { store });
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
  });

  it('calls onConfirm when Confirm button is clicked', () => {
    const onConfirm = vi.fn();
    const store = createTestStore();
    store.dispatch(
      openConfirmDialog({
        title: 'Confirm Action',
        message: 'This is a test',
        onConfirm,
      })
    );
    render(<ConfirmDialog />, { store });
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it('closes dialog when Cancel is clicked', () => {
    const store = createTestStore();
    store.dispatch(
      openConfirmDialog({
        title: 'Test',
        message: 'Test message',
        onConfirm: vi.fn(),
      })
    );
    render(<ConfirmDialog />, { store });
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    // After cancel, dialog should close (store state updated)
    expect(store.getState().ui.confirmDialogOpen).toBe(false);
  });

  it('closes dialog after confirm', () => {
    const onConfirm = vi.fn();
    const store = createTestStore();
    store.dispatch(
      openConfirmDialog({
        title: 'Test',
        message: 'Test message',
        onConfirm,
      })
    );
    render(<ConfirmDialog />, { store });
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }));
    expect(store.getState().ui.confirmDialogOpen).toBe(false);
  });

  it('handles onConfirm being undefined gracefully', () => {
    const store = createTestStore();
    // Manually set state with no onConfirm
    store.dispatch(
      openConfirmDialog({
        title: 'Test',
        message: 'No callback',
        onConfirm: undefined as any,
      })
    );
    render(<ConfirmDialog />, { store });
    // Should not throw
    expect(() => {
      fireEvent.click(screen.getByRole('button', { name: /confirm/i }));
    }).not.toThrow();
  });
});
