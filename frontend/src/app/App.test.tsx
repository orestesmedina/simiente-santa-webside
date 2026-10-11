import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { server } from '../test/server';
import { App } from './App';

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;

describe('App', () => {
  it('renderiza la aplicación en su estado inicial (portada pública en /)', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({ lang: 'es', identity: { name: 'Iglesia Simiente Santa' } }),
      ),
    );

    render(<App />);

    expect(
      await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' }),
    ).toBeInTheDocument();
  });
});
