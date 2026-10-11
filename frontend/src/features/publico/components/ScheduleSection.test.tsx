import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ScheduleItemPublic } from '../../../api/portada';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { ScheduleSection } from './ScheduleSection';

const items: ScheduleItemPublic[] = [
  {
    id: 'srv-1',
    dayOfWeek: 0,
    startTime: '10:00',
    endTime: '12:00',
    name: 'Culto dominical',
    place: 'Templo central',
  },
  {
    id: 'srv-2',
    dayOfWeek: 3,
    startTime: '18:00',
    endTime: null,
    name: 'Oración',
    place: 'Salón',
  },
];

function renderSection(list: ScheduleItemPublic[] = items) {
  return render(
    <LanguageProvider>
      <ScheduleSection items={list} />
    </LanguageProvider>,
  );
}

describe('ScheduleSection · representación dual (M2) con el sistema (T1022)', () => {
  it('la tabla conserva el nombre accesible «Horario de servicios» y el rango de hora', () => {
    renderSection();

    const table = screen.getByRole('table', { name: 'Horario de servicios' });
    expect(within(table).getByText('10:00 a. m. − 12:00 m.')).toBeInTheDocument();
    expect(within(table).getByText('6:00 p. m.')).toBeInTheDocument();
    expect(within(table).getByText('Domingo')).toBeInTheDocument();
    expect(within(table).getByText('Miércoles')).toBeInTheDocument();
  });

  it('sigue existiendo la lista de tarjetas de la representación móvil', () => {
    const { container } = renderSection();

    const list = container.querySelector('#horario ul') as HTMLElement;
    expect(list).not.toBeNull();
    expect(list.querySelectorAll('li')).toHaveLength(2);
    expect(within(list).getByText('Culto dominical')).toBeInTheDocument();
  });

  it('usa tokens semánticos y el espaciado de sección', () => {
    const { container } = renderSection();

    const section = container.querySelector('section') as HTMLElement;
    expect(section.className).toContain('py-section');
    expect(section.querySelector('.bg-surface-raised')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 2, name: 'Horario de servicios' })).toHaveAttribute(
      'id',
      'horario-title',
    );
  });

  it('un solo servicio se renderiza sin romper la composición', () => {
    renderSection([items[0]]);

    expect(screen.getAllByText('Culto dominical').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Templo central').length).toBeGreaterThan(0);
  });
});
