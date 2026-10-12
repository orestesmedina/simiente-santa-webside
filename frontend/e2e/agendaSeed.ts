import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { expect, type APIResponse, type Page } from '@playwright/test';
import { API_URL, apiSend, uniqueSuffix } from './helpers';

/**
 * Utilidades de siembra y limpieza de la agenda para los e2e de F4 (T438/T439/
 * T442). No es un `*.spec.ts` (Playwright no lo trata como prueba).
 *
 * La base es **compartida y no se limpia** entre corridas, así que cada spec
 * empieza con `resetAgenda`: borra los eventos y actividades (y las categorías
 * que no sembró la migración `000007`) para que las aserciones —en particular
 * «sin nada publicado y próximo» y «el borrador no se ve»— sean deterministas y
 * repetibles. Por eso los specs de F4 se ejecutan con `--workers=1` (igual que
 * los de F10): si dos archivos mutan la agenda a la vez, se estorban.
 */

/** Imagen real del repositorio, usada para las subidas (patrón portada-publica). */
export const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

/** Categorías del catálogo inicial (`000007`): no se borran nunca. */
const SEEDED_CATEGORIES = [
  'Acción social',
  'Predicación',
  'Adoración',
  'Entrevista',
  'Noche de oración',
  'Otro',
];

export interface AdminList<T> {
  items: T[];
}

export interface AgendaItem {
  id: string;
  titleEs: string;
  publicationState: 'draft' | 'published';
  isPast: boolean;
  nextDate: string | null;
}

export interface CategoryItem {
  id: string;
  nameEs: string;
  nameEn: string | null;
}

/**
 * Fecha de «hoy» **en la zona del servidor** (`APP_TZ=America/Costa_Rica`,
 * R4-2): el corte de «próximo» lo decide el backend con esa zona, así que la
 * siembra debe usar la misma fecha de calendario. El formato `en-CA` produce
 * directamente `YYYY-MM-DD`.
 */
export function costaRicaToday(): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'America/Costa_Rica',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date());
}

/** Desplaza una fecha `YYYY-MM-DD` en días (sin conversión de zona horaria). */
export function shiftDays(isoDate: string, days: number): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  const base = new Date(Date.UTC(year, month - 1, day));
  base.setUTCDate(base.getUTCDate() + days);
  return base.toISOString().slice(0, 10);
}

async function expectOk(response: APIResponse, label: string): Promise<unknown> {
  expect(response.ok(), `${label} → ${response.status()}`).toBe(true);
  return response.json();
}

/** Lista completa (tope del contrato) de un recurso del panel de agenda. */
async function listAll<T>(page: Page, path: string, label: string): Promise<T[]> {
  const body = (await expectOk(
    await apiSend(page, 'GET', `${path}?limit=100`),
    `GET ${label}`,
  )) as AdminList<T>;
  return body.items;
}

/**
 * Busca un rol por su identificador recorriendo el listado paginado: la base
 * compartida acumula roles de corridas anteriores (cada e2e de F2 crea los
 * suyos) y el del administrador, creado el primero, queda fuera de la primera
 * página.
 */
async function findRoleById(
  page: Page,
  roleId: string,
): Promise<{ id: string; permissions: string[] } | undefined> {
  const limit = 100;
  for (let offset = 0; offset < 1000; offset += limit) {
    const body = (await expectOk(
      await apiSend(page, 'GET', `/api/v1/admin/roles?limit=${limit}&offset=${offset}`),
      'GET roles',
    )) as { items: { id: string; permissions: string[] }[]; total: number };
    const found = body.items.find((item) => item.id === roleId);
    if (found) {
      return found;
    }
    if (offset + limit >= body.total) {
      return undefined;
    }
  }
  return undefined;
}

/** Los tres permisos de F4 (`eventos`, `actividades`, `categorias`, FR-012). */
const AGENDA_PERMISSIONS = ['eventos', 'actividades', 'categorias'] as const;

/**
 * Garantiza que el rol de la cuenta administradora tiene los tres permisos de
 * F4. La base de desarrollo puede venir de una corrida anterior a la migración
 * `000009` (el permiso `categorias` nace con ella y el administrador inicial lo
 * recibe solo si la inicialización ocurre **después**); el `Resolver` de sesión
 * lee los permisos vigentes en cada petición, así que basta con completar el rol
 * (idempotente: solo actúa si falta alguno).
 */
export async function ensureAgendaPermissions(page: Page): Promise<void> {
  const session = (await expectOk(
    await apiSend(page, 'GET', '/api/v1/auth/session'),
    'GET session',
  )) as { roleId: string; permissions: string[] };

  if (AGENDA_PERMISSIONS.every((code) => session.permissions.includes(code))) {
    return;
  }

  const role = await findRoleById(page, session.roleId);
  if (!role) {
    throw new Error('No se encontró el rol de la cuenta administradora');
  }
  const permissions = [...new Set([...role.permissions, ...AGENDA_PERMISSIONS])];
  expect(
    (await apiSend(page, 'PATCH', `/api/v1/admin/roles/${role.id}`, { permissions })).ok(),
    'el rol del administrador con los permisos de F4',
  ).toBe(true);
}

/**
 * Deja la agenda en un estado conocido: sin eventos, sin actividades y sin
 * categorías añadidas por corridas anteriores (las del catálogo inicial se
 * conservan). Es la base de la repetibilidad de los tres specs de F4.
 */
export async function resetAgenda(page: Page): Promise<void> {
  await ensureAgendaPermissions(page);

  for (const resource of ['eventos', 'actividades'] as const) {
    const items = await listAll<AgendaItem>(page, `/api/v1/admin/${resource}`, resource);
    for (const item of items) {
      expect(
        (await apiSend(page, 'DELETE', `/api/v1/admin/${resource}/${item.id}`)).status(),
        `DELETE ${resource}/${item.id}`,
      ).toBe(204);
    }
  }

  const categories = await listAll<CategoryItem>(page, '/api/v1/admin/categorias', 'categorias');
  for (const category of categories) {
    if (SEEDED_CATEGORIES.includes(category.nameEs)) {
      continue;
    }
    // Sin elementos que la usen, el borrado procede (204). Un 409 significaría
    // que quedó en uso: se deja (no rompe ninguna aserción).
    await apiSend(page, 'DELETE', `/api/v1/admin/categorias/${category.id}`);
  }
}

/** Identificador de una categoría del catálogo por su nombre en español. */
export async function categoryIdByName(page: Page, nameEs: string): Promise<string> {
  const categories = await listAll<CategoryItem>(page, '/api/v1/admin/categorias', 'categorias');
  const category = categories.find((item) => item.nameEs === nameEs);
  if (!category) {
    throw new Error(`No existe la categoría «${nameEs}» en el catálogo`);
  }
  return category.id;
}

/** Sube una imagen al módulo indicado y devuelve su `fileName`. */
export async function uploadAgendaImage(
  page: Page,
  module: 'eventos' | 'actividades',
): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) {
    throw new Error('No hay cookie csrf_token: ¿se ha iniciado sesión en este contexto?');
  }
  const upload = await page.request.post(`${API_URL}/api/v1/admin/${module}/imagenes`, {
    headers: { 'X-CSRF-Token': csrf.value },
    multipart: {
      file: { name: 'foto.jpg', mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
    },
  });
  expect(upload.status(), `la subida de imagen de ${module}`).toBe(201);
  return ((await upload.json()) as { fileName: string }).fileName;
}

export interface AgendaSeed {
  categoryName: string;
  eventTitle: string;
  eventTitleEn: string;
  eventPlace: string;
  eventDetail: string;
  eventFrequency: string;
  eventMapUrl: string;
  eventImageAlt: string;
  eventImageFile: string;
  /** Dos fechas futuras: la próxima y la siguiente (⟲ D-11). */
  eventNextDate: string;
  eventLaterDate: string;
  /** Evento publicado con **todas** sus fechas pasadas: el público no lo ve. */
  pastEventTitle: string;
  /** Evento en borrador con imagen: invisible y su imagen no se sirve. */
  draftEventTitle: string;
  draftImageFile: string;
  /** Evento de hoy con `endTime` ya terminado: se retira solo (R4-2). */
  todayFinishedTitle: string;
  today: string;
  activityTitle: string;
  activityPlace: string;
  activityDetail: string;
  activityImageAlt: string;
  activityDate: string;
  pastActivityTitle: string;
}

/**
 * Siembra la agenda del escenario público: un evento y una actividad publicados
 * con fecha próxima (con imagen, categoría, detalle y enlace de mapa), más los
 * tres casos que **no** deben verse —pasado, borrador y el de hoy ya terminado—
 * y una actividad vencida. Todo con un sufijo único por ejecución.
 */
export async function seedAgenda(page: Page): Promise<AgendaSeed> {
  const suffix = uniqueSuffix();
  const today = costaRicaToday();
  const eventNextDate = shiftDays(today, 7);
  const eventLaterDate = shiftDays(today, 14);
  const pastDate = shiftDays(today, -3);
  const activityDate = shiftDays(today, 8);

  const categoryName = 'Noche de oración';
  const categoryId = await categoryIdByName(page, categoryName);

  // El mapa del sitio del pie solo dibuja el grupo «Explorar» cuando hay algún
  // enlace de sección: se garantiza con una publicación idempotente de «quiénes
  // somos» para que la comprobación de FR-022 (barra, menú móvil y pie) valga
  // también sobre una base recién creada.
  await expectOk(
    await apiSend(page, 'PUT', '/api/v1/admin/portada/quienes-somos', {
      textEs: `Somos una iglesia cercana ${suffix}.`,
      publicationState: 'published',
    }),
    'la publicación de «quiénes somos»',
  );

  const eventImageFile = await uploadAgendaImage(page, 'eventos');
  const draftImageFile = await uploadAgendaImage(page, 'eventos');
  const activityImageFile = await uploadAgendaImage(page, 'actividades');

  const seed: AgendaSeed = {
    categoryName,
    eventTitle: `Vigilia de viernes ${suffix}`,
    eventTitleEn: `Friday vigil ${suffix}`,
    eventPlace: `Templo central ${suffix}`,
    eventDetail: `Noche de oración abierta al barrio.\nTrae tu Biblia ${suffix}.`,
    eventFrequency: `Todos los viernes ${suffix}`,
    eventMapUrl: `https://maps.google.com/?q=simiente-${suffix}`,
    eventImageAlt: `Templo iluminado ${suffix}`,
    eventImageFile,
    eventNextDate,
    eventLaterDate,
    pastEventTitle: `Vigilia pasada ${suffix}`,
    draftEventTitle: `Vigilia en borrador ${suffix}`,
    draftImageFile,
    todayFinishedTitle: `Encuentro de hoy ${suffix}`,
    today,
    activityTitle: `Repartir comida ${suffix}`,
    activityPlace: `Comedor comunal ${suffix}`,
    activityDetail: `Acción social del mes ${suffix}`,
    activityImageAlt: `Voluntarios repartiendo comida ${suffix}`,
    activityDate,
    pastActivityTitle: `Reparto pasado ${suffix}`,
  };

  // Evento publicado y próximo (el protagonista de la sección pública).
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/eventos', {
      titleEs: seed.eventTitle,
      titleEn: seed.eventTitleEn,
      frequencyEs: seed.eventFrequency,
      frequencyEn: `Every Friday ${suffix}`,
      detailEs: seed.eventDetail,
      detailEn: `Prayer night open to the neighbourhood ${suffix}`,
      placeEs: seed.eventPlace,
      placeEn: `Main temple ${suffix}`,
      mapUrl: seed.eventMapUrl,
      categoryId,
      imageFile: eventImageFile,
      imageAltEs: seed.eventImageAlt,
      imageAltEn: `Illuminated temple ${suffix}`,
      publicationState: 'published',
      dates: [
        { date: eventNextDate, startTime: '19:00', endTime: '21:00' },
        { date: eventLaterDate, startTime: '19:00' },
      ],
    }),
    'el evento publicado',
  );

  // Publicado pero con todas sus fechas pasadas: no se muestra (Q3/SC-009).
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/eventos', {
      titleEs: seed.pastEventTitle,
      detailEs: 'Ya pasó.',
      placeEs: 'Templo',
      publicationState: 'published',
      dates: [{ date: pastDate, startTime: '19:00' }],
    }),
    'el evento vencido',
  );

  // Borrador: jamás visible, ni su imagen (SC-002).
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/eventos', {
      titleEs: seed.draftEventTitle,
      detailEs: 'Todavía sin publicar.',
      placeEs: 'Salón',
      imageFile: draftImageFile,
      imageAltEs: 'Borrador',
      dates: [{ date: eventNextDate }],
    }),
    'el evento en borrador',
  );

  // Hoy con hora de fin ya terminada: se retira al terminar (R4-2/SC-009).
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/eventos', {
      titleEs: seed.todayFinishedTitle,
      detailEs: 'Terminó esta madrugada.',
      placeEs: 'Templo',
      publicationState: 'published',
      dates: [{ date: today, startTime: '00:00', endTime: '00:01' }],
    }),
    'el evento de hoy ya terminado',
  );

  // Actividad publicada y próxima (con horario e imagen).
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/actividades', {
      titleEs: seed.activityTitle,
      detailEs: seed.activityDetail,
      placeEs: seed.activityPlace,
      mapUrl: `https://maps.google.com/?q=comedor-${suffix}`,
      categoryId,
      imageFile: activityImageFile,
      imageAltEs: seed.activityImageAlt,
      publicationState: 'published',
      date: activityDate,
      startTime: '08:00',
      endTime: '11:00',
    }),
    'la actividad publicada',
  );

  // Actividad vencida: fuera de «lo que viene».
  await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/actividades', {
      titleEs: seed.pastActivityTitle,
      detailEs: 'Ya se hizo.',
      placeEs: 'Comedor',
      publicationState: 'published',
      date: pastDate,
    }),
    'la actividad vencida',
  );

  return seed;
}
