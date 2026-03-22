import { describe, it, expect } from 'vitest';
import { render, screen } from '@/test/test-utils';
import { Logo, LogoWithText } from '../Logo';

describe('Logo', () => {
  it('renders logo image', () => {
    render(<Logo />);
    const img = screen.getByAltText(/environment manager logo/i);
    expect(img).toBeInTheDocument();
  });

  it('applies custom size', () => {
    render(<Logo size={60} />);
    const img = screen.getByAltText(/environment manager logo/i);
    // size is applied as inline style
    expect(img).toHaveAttribute('alt', 'Environment Manager Logo');
    expect((img as HTMLImageElement).style.width).toBe('60px');
  });

  it('applies filter when non-default color is passed', () => {
    render(<Logo color="#ff0000" />);
    const img = screen.getByAltText(/environment manager logo/i);
    // Filter style should be applied
    expect(img.style.filter).not.toBe('');
  });

  it('does not apply filter for default color', () => {
    render(<Logo color="#1976d2" />);
    const img = screen.getByAltText(/environment manager logo/i);
    expect(img.style.filter).toBe('');
  });
});

describe('LogoWithText', () => {
  it('renders logo with text', () => {
    render(<LogoWithText />);
    expect(screen.getByText('Environment')).toBeInTheDocument();
    expect(screen.getByText('Manager')).toBeInTheDocument();
  });

  it('renders logo image within LogoWithText', () => {
    render(<LogoWithText />);
    expect(screen.getByAltText(/environment manager logo/i)).toBeInTheDocument();
  });

  it('applies custom size to logo', () => {
    render(<LogoWithText size={50} />);
    const img = screen.getByAltText(/environment manager logo/i);
    expect((img as HTMLImageElement).style.width).toBe('50px');
  });
});
