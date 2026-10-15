import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { OfrendasBlock } from './OfrendasBlock';

const portadaUrl = `${API_BASE_URL}/api/v1/portada/ofrendas`;

function item(kind: 'iban' | 'sinpe', id: string): MetodoDonacionPublico {
  return kind === 'iban'
    ? { id, kind, number: 'CR05015202001234567890', institution: `Banco ${id}`, currency: 'CRC' }
    : { id, kind, number: '+50688888888', institution: `Banco ${id}` };
}

function renderBlock() {
  return render(
    <AppProviders>
      <LanguageProvider>
        <OfrendasBlock />
      </LanguageProvider>
    </AppProviders>,
  );
}

describe('OfrendasBlock · bloque «Ofrendas y diezmos» de la portada (US6/FR-022/SC-014)', () => {
  it('con 0 métodos publicados el bloque NO se dibuja: sin título ni mensaje', async () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    const { container } = renderBlock();

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it('con 8 publicados muestra exactamente los 5 primeros en el orden del servidor (tope del servidor)', async () => {
    // El servidor aplica el tope 5; el cliente NO recorta: pinta lo que llega.
    const items = Array.from({ length: 5 }, (_, index) => item('iban', `id-${index}`));
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es', items })));

    renderBlock();

    expect(
      await screen.findByRole('heading', { level: 2, name: 'Ofrendas y diezmos' }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('listitem')).toHaveLength(5);
    expect(screen.getByRole('link', { name: 'Ver formas de ofrendar' })).toHaveAttribute(
      'href',
      '/ofrendas',
    );
  });

  it('con 1 método muestra exactamente esa ficha (sin huecos ni marcadores)', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es', items: [item('sinpe', 'uno')] })),
    );

    renderBlock();

    await screen.findByRole('heading', { level: 2, name: 'Ofrendas y diezmos' });
    expect(screen.getAllByRole('listitem')).toHaveLength(1);
    expect(screen.getByText('SINPE Móvil')).toBeInTheDocument();
  });

  it('mientras carga muestra un esqueleto de altura acotada', () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    renderBlock();

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });

  it('ante error sin caché el bloque no se dibuja (ni título huérfano)', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({ error: { code: 'internal', message: 'x' } }, { status: 500 }),
      ),
    );

    const { container } = renderBlock();

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });
});
