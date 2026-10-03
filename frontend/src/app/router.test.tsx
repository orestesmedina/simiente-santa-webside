import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

describe('AppRoutes', () => {
  it('renderiza el destino de la ruta inicial dentro del layout', () => {
    render(
      <AppProviders>
        <MemoryRouter initialEntries={['/']}>
          <AppRoutes />
        </MemoryRouter>
      </AppProviders>,
    );

    expect(screen.getByRole('heading', { name: 'Estado del sistema' })).toBeInTheDocument();
    expect(screen.getByRole('main')).toBeInTheDocument();
  });
});
