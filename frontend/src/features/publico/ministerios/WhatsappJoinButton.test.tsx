import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { WhatsappJoinButton } from './WhatsappJoinButton';

/**
 * Botón de unirse (FR-003/FR-015, ux.md §4.1/§6.3): solo existe con enlace, se
 * sirve **tal cual** el valor guardado (sin transformar), abre en pestaña nueva
 * con `rel="noopener"` y lleva el nombre accesible completo por ministerio.
 */

function renderButton(props: { url: string; name: string }) {
  return render(
    <LanguageProvider>
      <WhatsappJoinButton {...props} />
    </LanguageProvider>,
  );
}

describe('WhatsappJoinButton · unirse en un paso (FR-003)', () => {
  it('enlaza al valor guardado sin transformarlo y abre fuera con noopener', () => {
    renderButton({ url: 'https://chat.whatsapp.com/ABC123?ref=grupo', name: 'Alabanza' });

    const link = screen.getByRole('link', { name: 'Unirse al grupo de WhatsApp de Alabanza' });
    expect(link).toHaveAttribute('href', 'https://chat.whatsapp.com/ABC123?ref=grupo');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('el texto visible es el del diccionario y el icono es decorativo', () => {
    renderButton({ url: 'https://ejemplo.com/grupo', name: 'Escuela dominical' });

    expect(screen.getByText('Unirse al grupo de WhatsApp')).toBeVisible();
    // El icono no aporta texto accesible (el nombre lo da el `aria-label`).
    expect(
      screen.getByRole('link', { name: 'Unirse al grupo de WhatsApp de Escuela dominical' }),
    ).toBeVisible();
  });

  it('en inglés usa las cadenas del diccionario (SC-006)', () => {
    window.localStorage.setItem('ss.lang', 'en');
    renderButton({ url: 'https://chat.whatsapp.com/ABC123', name: 'Worship' });

    expect(screen.getByRole('link', { name: 'Join the WhatsApp group of Worship' })).toBeVisible();
    expect(screen.getByText('Join the WhatsApp group')).toBeVisible();
    window.localStorage.clear();
  });
});
