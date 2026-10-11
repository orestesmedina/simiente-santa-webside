import { describe, expect, it } from 'vitest';
import type { PortadaPublica } from '../../api/portada';
import { PUBLIC_SECTIONS, isSectionPresent, presentSections } from './sections';

const vacia: PortadaPublica = { lang: 'es' };

const llena: PortadaPublica = {
  lang: 'es',
  identity: { name: 'Iglesia Simiente Santa' },
  about: { text: 'Somos una familia' },
  schedule: [{ id: 'a', dayOfWeek: 0, startTime: '10:00', name: 'Culto', place: 'Templo' }],
  whatsapp: [{ id: 'b', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' }],
  socials: [{ id: 'c', network: 'facebook', url: 'https://facebook.com/simiente' }],
  contact: { address: 'San José', email: 'hola@ejemplo.com', phone: '+506 8888 8888' },
};

describe('presentSections (SC-012)', () => {
  it('sin datos no hay ninguna sección', () => {
    expect(presentSections(undefined)).toEqual([]);
  });

  it('una respuesta vacía no pinta ninguna sección', () => {
    expect(presentSections(vacia)).toEqual([]);
  });

  it('devuelve las secciones con contenido publicado en el orden de la página', () => {
    expect(presentSections(llena).map((section) => section.id)).toEqual([
      'who',
      'schedule',
      'whatsapp',
      'contact',
      'social',
    ]);
  });

  it('una colección vacía no cuenta como sección presente', () => {
    const soloListasVacias: PortadaPublica = {
      lang: 'es',
      identity: { name: 'Iglesia' },
      schedule: [],
      whatsapp: [],
      socials: [],
    };
    expect(presentSections(soloListasVacias)).toEqual([]);
    expect(isSectionPresent('schedule', soloListasVacias)).toBe(false);
  });

  it('la identidad no es un ancla de la navegación (encabeza la página)', () => {
    expect(PUBLIC_SECTIONS.map((section) => section.id)).not.toContain('identity');
  });
});
