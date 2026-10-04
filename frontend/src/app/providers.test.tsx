import { useQuery } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { AppProviders } from './providers';

function Probe() {
  const { data } = useQuery({
    queryKey: ['probe'],
    queryFn: () => Promise.resolve('ok'),
  });

  return <span>{data ?? 'cargando'}</span>;
}

describe('AppProviders', () => {
  it('envuelve la aplicación con el QueryClientProvider', async () => {
    render(
      <AppProviders>
        <Probe />
      </AppProviders>,
    );

    expect(await screen.findByText('ok')).toBeInTheDocument();
  });
});
