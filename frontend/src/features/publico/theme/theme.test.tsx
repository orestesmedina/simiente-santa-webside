import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { installMatchMedia } from '../../../test/matchMedia';
import { ThemeProvider } from './ThemeProvider';
import { THEME_STORAGE_KEY, useTheme } from './useTheme';

const DARK_QUERY = '(prefers-color-scheme: dark)';

function Consumer() {
  const { theme, preference, setTheme } = useTheme();
  return (
    <div>
      <p data-testid="theme">{theme}</p>
      <p data-testid="preference">{preference ?? 'system'}</p>
      <button type="button" onClick={() => setTheme('dark')}>
        poner oscuro
      </button>
      <button type="button" onClick={() => setTheme('light')}>
        poner claro
      </button>
    </div>
  );
}

function renderTheme() {
  return render(
    <ThemeProvider>
      <Consumer />
    </ThemeProvider>,
  );
}

describe('tema del sitio público (US5, FR-023)', () => {
  beforeEach(() => {
    window.localStorage.clear();
    delete document.documentElement.dataset.theme;
    document.documentElement.style.removeProperty('color-scheme');
  });

  afterEach(() => {
    window.localStorage.clear();
    vi.restoreAllMocks();
  });

  it('sin elección guardada sigue la preferencia del sistema (oscuro)', () => {
    installMatchMedia({ [DARK_QUERY]: true });

    renderTheme();

    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
    expect(screen.getByTestId('preference')).toHaveTextContent('system');
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(document.documentElement.style.getPropertyValue('color-scheme')).toBe('dark');
  });

  it('sin preferencia oscura del sistema entra en claro', () => {
    installMatchMedia({ [DARK_QUERY]: false });

    renderTheme();

    expect(screen.getByTestId('theme')).toHaveTextContent('light');
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('la elección manual cambia el tema al instante y persiste en ss.theme', async () => {
    installMatchMedia({ [DARK_QUERY]: false });
    renderTheme();

    await userEvent.click(screen.getByRole('button', { name: 'poner oscuro' }));

    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
    expect(screen.getByTestId('preference')).toHaveTextContent('dark');
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('una re-visita con elección guardada respeta la manual, aunque el sistema diga lo contrario', () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, 'light');
    installMatchMedia({ [DARK_QUERY]: true });

    renderTheme();

    expect(screen.getByTestId('theme')).toHaveTextContent('light');
    expect(screen.getByTestId('preference')).toHaveTextContent('light');
  });

  it('el cambio en caliente del sistema solo aplica sin elección manual', () => {
    const media = installMatchMedia({ [DARK_QUERY]: false });
    renderTheme();
    expect(screen.getByTestId('theme')).toHaveTextContent('light');

    act(() => media.set(DARK_QUERY, true));

    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
  });

  it('con elección manual, el cambio del sistema no pisa la elección', async () => {
    const media = installMatchMedia({ [DARK_QUERY]: true });
    renderTheme();

    await userEvent.click(screen.getByRole('button', { name: 'poner claro' }));
    expect(screen.getByTestId('theme')).toHaveTextContent('light');

    act(() => media.set(DARK_QUERY, true));

    expect(screen.getByTestId('theme')).toHaveTextContent('light');
  });

  it('con localStorage caído usa el tema del sistema sin excepciones', () => {
    installMatchMedia({ [DARK_QUERY]: true });
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('sin almacenamiento');
    });

    expect(() => renderTheme()).not.toThrow();

    expect(screen.getByTestId('theme')).toHaveTextContent('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
  });
});
