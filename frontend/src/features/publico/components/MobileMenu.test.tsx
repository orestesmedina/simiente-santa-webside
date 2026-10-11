import { act, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it } from 'vitest';
import { installMatchMedia } from '../../../test/matchMedia';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { MobileMenu } from './MobileMenu';
import { useMobileMenu } from './useMobileMenu';

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '#quienes-somos' },
  { id: 'schedule', anchor: 'horario', messageKey: 'nav.sections.schedule', href: '#horario' },
  { id: 'contact', anchor: 'contacto', messageKey: 'nav.sections.contact', href: '#contacto' },
];

/**
 * Compone el menú igual que `PublicNav`: un `<nav>` nombrado con el botón de
 * `MobileMenu` y **una sola** lista de enlaces que hace de panel.
 */
function Harness() {
  const { t } = useLanguage();
  const menu = useMobileMenu();
  const { open, panelId, panelRef, close } = menu;

  return (
    <nav aria-label={t('nav.label')}>
      <MobileMenu menu={menu} />
      <ul id={panelId} ref={panelRef} data-open={open} className="public-nav-list">
        {links.map((link) => (
          <li key={link.id}>
            <a href={link.href} onClick={() => close(false)}>
              {t(link.messageKey)}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}

function renderMenu() {
  return render(
    <LanguageProvider>
      <Harness />
    </LanguageProvider>,
  );
}

afterEach(() => {
  delete document.documentElement.dataset.navOpen;
});

describe('MobileMenu · patrón disclosure accesible (T1013/D17)', () => {
  it('sincroniza aria-expanded y aria-controls con la lista, y tiene nombre localizado', async () => {
    renderMenu();

    const button = screen.getByRole('button', { name: 'Abrir menú de secciones' });
    const list = document.querySelector('ul.public-nav-list');
    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(button).toHaveAttribute('aria-controls', list?.id);

    await userEvent.click(button);

    expect(screen.getByRole('button', { name: 'Cerrar menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
  });

  it('al abrir el foco entra en el primer enlace y NO es un diálogo (sin trampa de foco)', async () => {
    renderMenu();

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));

    expect(screen.getByRole('link', { name: 'Quiénes somos' })).toHaveFocus();
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(document.querySelector('[aria-modal]')).toBeNull();
  });

  it('Escape cierra el panel y devuelve el foco al botón', async () => {
    renderMenu();
    const button = screen.getByRole('button', { name: 'Abrir menú de secciones' });

    await userEvent.click(button);
    await userEvent.keyboard('{Escape}');

    expect(screen.getByRole('button', { name: 'Abrir menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
    expect(button).toHaveFocus();
  });

  it('el propio botón cierra el panel', async () => {
    renderMenu();

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));
    await userEvent.click(screen.getByRole('button', { name: 'Cerrar menú de secciones' }));

    expect(screen.getByRole('button', { name: 'Abrir menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
  });

  it('el scrim cierra el panel y devuelve el foco al botón', async () => {
    renderMenu();
    const button = screen.getByRole('button', { name: 'Abrir menú de secciones' });

    await userEvent.click(button);
    await userEvent.click(screen.getByTestId('mobile-menu-scrim'));

    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(button).toHaveFocus();
  });

  it('activar una sección cierra el panel', async () => {
    renderMenu();

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));
    await userEvent.click(screen.getByRole('link', { name: 'Horario de servicios' }));

    expect(screen.getByRole('button', { name: 'Abrir menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
  });

  it('al cruzar a escritorio (≥ 861 px) el panel se cierra solo', async () => {
    const media = installMatchMedia({ '(min-width: 861px)': false });
    renderMenu();

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));
    expect(screen.getByRole('button', { name: 'Cerrar menú de secciones' })).toBeInTheDocument();

    act(() => media.set('(min-width: 861px)', true));

    expect(screen.getByRole('button', { name: 'Abrir menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
  });

  it('bloquea el scroll de fondo con data-nav-open mientras está abierto', async () => {
    renderMenu();
    expect(document.documentElement.dataset.navOpen).toBeUndefined();

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));
    expect(document.documentElement.dataset.navOpen).toBe('true');

    await userEvent.click(screen.getByTestId('mobile-menu-scrim'));
    expect(document.documentElement.dataset.navOpen).toBeUndefined();
  });

  it('no contiene el selector de idioma (D2) ni una segunda lista de enlaces (D15)', async () => {
    renderMenu();
    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });

    await userEvent.click(screen.getByRole('button', { name: 'Abrir menú de secciones' }));

    expect(within(nav).queryAllByRole('radio')).toHaveLength(0);
    expect(within(nav).getAllByRole('link', { name: 'Quiénes somos' })).toHaveLength(1);
    expect(nav.querySelectorAll('ul')).toHaveLength(1);
  });

  it('con el menú cerrado la lista lleva la clase que la colapsa (display:none a ≤ 860 px)', () => {
    renderMenu();

    const list = document.querySelector('ul.public-nav-list');
    expect(list).not.toBeNull();
    expect(list).toHaveAttribute('data-open', 'false');
  });
});
