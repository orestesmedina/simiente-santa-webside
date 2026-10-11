import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { PublicLayout } from './PublicLayout';

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;

function renderLayout() {
  server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<PublicLayout />}>
            <Route index element={<p>contenido</p>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </AppProviders>,
  );
}

afterEach(() => {
  window.localStorage.clear();
  delete document.documentElement.dataset.theme;
});

describe('PublicLayout · integración del chrome público (T1014)', () => {
  it('conserva el skip-link, el main y el landmark de contenido', () => {
    renderLayout();

    expect(screen.getByRole('link', { name: 'Saltar al contenido principal' })).toHaveAttribute(
      'href',
      '#contenido',
    );
    expect(screen.getByRole('main')).toHaveAttribute('id', 'contenido');
    expect(screen.getByText('contenido')).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toBeInTheDocument();
  });

  it('monta el ThemeProvider y aplica data-theme en <html>', () => {
    renderLayout();

    expect(['light', 'dark']).toContain(document.documentElement.dataset.theme);
  });

  it('mantiene el orden de landmarks: banner → main → contentinfo', () => {
    const { container } = renderLayout();

    const banner = screen.getByRole('banner');
    const main = screen.getByRole('main');
    const contentinfo = screen.getByRole('contentinfo');
    expect(banner.compareDocumentPosition(main) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(
      main.compareDocumentPosition(contentinfo) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(container.querySelectorAll('main')).toHaveLength(1);
  });
});
