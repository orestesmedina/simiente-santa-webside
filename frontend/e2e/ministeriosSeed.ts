import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { expect, type APIResponse, type Page } from '@playwright/test';
import { API_URL, apiSend, uniqueSuffix } from './helpers';

/**
 * Utilidades de siembra y limpieza de **ministerios** para los e2e de F6
 * (T632–T636). No es un `*.spec.ts`, así que Playwright no lo trata como prueba.
 *
 * La base es **compartida y no se limpia** entre corridas, así que cada spec
 * empieza con `resetMinisterios` (o con `seedMinisterios`, que lo hace por
 * dentro): borra **todos** los ministerios —publicados y borradores— para que las
 * aserciones («0 borradores visibles», «exactamente 5 en el bloque», «sin
 * publicados redirige», «0 apariciones del encargado oculto») sean deterministas
 * y repetibles. Por eso los specs de F6 se ejecutan con `--workers=1`: si dos
 * archivos mutaran la tabla a la vez, se estorbarían (mismo motivo que F4/F10).
 *
 * El catálogo (`FIXTURES`) cubre, de una vez, los casos que piden las cuatro
 * specs (y las tarjetas de `tasks.md`):
 *
 *   1. `alabanza`        publicado con EN, horario/lugar, imagen+alt, enlace A y
 *                        encargado **visible** con celular (US1, SC-009/SC-017).
 *   2. `sonido`          publicado **sin EN** (fallback en→es), sin imagen, sin
 *                        enlace y con encargado **oculto** pero nombre y celular
 *                        reales (SC-014/RG6-1) — el caso crítico del módulo.
 *   3. `accion-social`   publicado con horario de **texto libre largo**
 *                        («Una vez al mes, fecha variable», SC-015), sin enlace,
 *                        sin encargado y sin imagen (SC-017).
 *   4. `matrimonios`     publicado **solo con nombre y descripción** (SC-017).
 *   5. `ninos`           publicado con **tildes, `ñ`, `¿¡` y emojis** en nombre y
 *                        descripción, con imagen+alt y horario/lugar (edge case).
 *   6. `jovenes`         publicado con imagen y encargado **visible** cuyo único
 *                        medio de contacto es el **correo** (Q8).
 *   7. `intercesion`     publicado con el **mismo enlace** que `alabanza`
 *                        (aviso de enlace compartido, D-7/R6-11).
 *   8. `misiones`        publicado creado **sin la clave `leaderVisible`** y con
 *                        encargado con nombre y celular: el alta nace **oculta**
 *                        («Oculto por defecto», P6-3).
 *   + `borrador-*` (×2)   en borrador —uno con imagen—: invisibles por cualquier
 *                        vía y su imagen **no se sirve** (SC-002).
 *
 * Con `{ publishedCount: 2 }` se siembran solo los dos primeros publicados
 * (escenario «exactos con menos, sin huecos» del bloque de portada, SC-016) y con
 * `publishedCount: 8` (por defecto) los ocho (tope 5 del bloque).
 */

/** Imagen real del repositorio, usada para las subidas (patrón portada-publica). */
export const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

export interface MinisterioAdminItem {
  id: string;
  nameEs: string;
  leaderVisible: boolean;
  publicationState: 'draft' | 'published';
  sortOrder: number;
}

/** Un ministerio sembrado, con su `id` real y los datos con los que se creó. */
export interface MinisterioFixture {
  /** Clave estable del catálogo (para localizarlo en las aserciones). */
  key: string;
  id: string;
  nameEs: string;
  nameEn: string | null;
  descriptionEs: string;
  descriptionEn: string | null;
  scheduleEs?: string;
  placeEs?: string;
  leaderNameEs?: string;
  leaderPhone?: string;
  leaderEmail?: string;
  leaderVisible: boolean;
  whatsappUrl?: string;
  sortOrder: number;
  publicationState: 'draft' | 'published';
  /** Nombre del archivo subido, si el ministerio tiene imagen. */
  imageFile?: string;
  imageAltEs?: string;
  /** `true` si el alta se hizo **sin** la clave `leaderVisible` (P6-3). */
  createdWithoutLeaderVisibleKey?: boolean;
}

export interface MinisterioSeed {
  suffix: string;
  /** Publicados sembrados, en el orden manual que devuelve el servidor. */
  published: MinisterioFixture[];
  /** Los **5 primeros** publicados: los que pinta el bloque de la portada. */
  portada: MinisterioFixture[];
  /** Borradores sembrados (invisibles al público). */
  drafts: MinisterioFixture[];
  /** `true` si esta siembra incluyó los borradores. */
  hasDrafts: boolean;
  /** Enlace compartido por `alabanza` e `intercesion` (D-7). */
  sharedLink: string;
  /** Archivo de imagen del borrador: `/api/v1/media/…` debe responder 404. */
  draftImageFile: string;
}

/** El catálogo base, sin `id` (se completan al crear). */
interface Fixture extends Omit<MinisterioFixture, 'id'> {
  /** Sube y asocia una imagen del repositorio. */
  withImage?: boolean;
  /** Alta **sin** la clave `leaderVisible` en el cuerpo (P6-3). */
  omitLeaderVisibleKey?: boolean;
}

async function expectOk(response: APIResponse, label: string): Promise<unknown> {
  expect(response.ok(), `${label} → ${response.status()}`).toBe(true);
  return response.json();
}

/** Lista completa (tope del contrato, 100) de ministerios del panel. */
async function listAll(page: Page): Promise<MinisterioAdminItem[]> {
  const body = (await expectOk(
    await apiSend(page, 'GET', '/api/v1/admin/ministerios?limit=100'),
    'GET ministerios',
  )) as { items: MinisterioAdminItem[] };
  return body.items;
}

/**
 * Deja el módulo en un estado conocido: **sin ningún ministerio** (ni publicado
 * ni borrador). Es la base de la repetibilidad de los cuatro specs de F6 y,
 * además, el escenario «0 publicados» de la redirección de `/ministerios`
 * (SC-010/P6-21) y del bloque ausente de la portada (SC-016/FR-028).
 */
export async function resetMinisterios(page: Page): Promise<void> {
  const items = await listAll(page);
  for (const item of items) {
    expect(
      (await apiSend(page, 'DELETE', `/api/v1/admin/ministerios/${item.id}`)).status(),
      `DELETE ministerios/${item.id}`,
    ).toBe(204);
  }
}

/** Sube una imagen de ministerio y devuelve su `fileName` (FR-020). */
export async function uploadMinistryImage(page: Page): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) {
    throw new Error('No hay cookie csrf_token: ¿se ha iniciado sesión en este contexto?');
  }
  const upload = await page.request.post(`${API_URL}/api/v1/admin/ministerios/imagenes`, {
    headers: { 'X-CSRF-Token': csrf.value },
    multipart: {
      file: { name: 'ministerio.jpg', mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
    },
  });
  expect(upload.status(), 'la subida de imagen de ministerio').toBe(201);
  return ((await upload.json()) as { fileName: string }).fileName;
}

/**
 * Catálogo del módulo con el sufijo de la ejecución. El orden de la lista es el
 * orden manual (`sortOrder`) y, por tanto, el que ve el visitante.
 */
function catalog(suffix: string): Fixture[] {
  const sharedLink = `https://chat.whatsapp.com/alabanza-${suffix}`;
  const longSchedule = `Una vez al mes, fecha variable — se anuncia el domingo anterior (${suffix})`;

  return [
    {
      key: 'alabanza',
      nameEs: `Ministerio de alabanza ${suffix}`,
      nameEn: `Worship ministry ${suffix}`,
      descriptionEs: `Cantamos y dirigimos la adoración de cada culto.\nEnsayo los jueves ${suffix}.`,
      descriptionEn: `We sing and lead worship at every service ${suffix}.`,
      scheduleEs: `Todos los viernes, 7:00 p. m. (${suffix})`,
      placeEs: `Templo principal ${suffix}`,
      leaderNameEs: `Ana Solís ${suffix}`,
      leaderPhone: '+506 7000 1111',
      leaderVisible: true,
      whatsappUrl: sharedLink,
      sortOrder: 10,
      publicationState: 'published',
      withImage: true,
      imageAltEs: `Personas cantando en el templo ${suffix}`,
    },
    {
      key: 'sonido',
      nameEs: `Ministerio de sonido ${suffix}`,
      nameEn: null,
      descriptionEs: `Operamos el audio y la proyección de los cultos ${suffix}.`,
      descriptionEn: null,
      // Encargado con nombre y celular REALES pero con la ficha **no publicada**:
      // el caso crítico SC-014 (0 apariciones en red y HTML).
      leaderNameEs: `Beto Ramírez ${suffix}`,
      leaderPhone: '+506 7000 9876',
      leaderVisible: false,
      sortOrder: 20,
      publicationState: 'published',
    },
    {
      key: 'accion-social',
      nameEs: `Acción social ${suffix}`,
      nameEn: `Social action ${suffix}`,
      descriptionEs: `Repartimos alimentos y visitamos a personas enfermas ${suffix}.`,
      descriptionEn: `We share food and visit the sick ${suffix}.`,
      scheduleEs: longSchedule,
      placeEs: `Barrio del Carmen ${suffix}`,
      leaderVisible: false,
      sortOrder: 30,
      publicationState: 'published',
    },
    {
      key: 'matrimonios',
      nameEs: `Matrimonios ${suffix}`,
      nameEn: null,
      descriptionEs: `Acompañamos a las parejas que se preparan para casarse ${suffix}.`,
      descriptionEn: null,
      leaderVisible: false,
      sortOrder: 40,
      publicationState: 'published',
    },
    {
      key: 'ninos',
      nameEs: `Niños y niñas ¡Bienvenidos! 🎈 ${suffix}`,
      nameEn: null,
      descriptionEs: `¿Te gusta enseñar? ¡Acompáñanos! Cuidamos a los más pequeños con cariño y ñoñería ${suffix}. 🎉`,
      descriptionEn: null,
      scheduleEs: `Domingos durante el culto (${suffix})`,
      placeEs: `Salón de niños ${suffix}`,
      leaderVisible: false,
      sortOrder: 50,
      publicationState: 'published',
      withImage: true,
      imageAltEs: `Niñas y niños jugando ${suffix}`,
    },
    {
      key: 'jovenes',
      nameEs: `Jóvenes ${suffix}`,
      nameEn: `Youth ${suffix}`,
      descriptionEs: `Encuentros, salidas y discipulado para adolescentes ${suffix}.`,
      descriptionEn: `Meetings, outings and discipleship for teenagers ${suffix}.`,
      scheduleEs: `Sábados, 4:00 p. m. (${suffix})`,
      placeEs: `Salón juvenil ${suffix}`,
      // Único medio de contacto: el correo (Q8).
      leaderNameEs: `Carla Méndez ${suffix}`,
      leaderEmail: `carla-${suffix}@ejemplo.com`,
      leaderVisible: true,
      whatsappUrl: `https://chat.whatsapp.com/jovenes-${suffix}`,
      sortOrder: 60,
      publicationState: 'published',
      withImage: true,
      imageAltEs: `Jóvenes en un encuentro ${suffix}`,
    },
    {
      key: 'intercesion',
      nameEs: `Intercesión ${suffix}`,
      nameEn: null,
      descriptionEs: `Oramos por las necesidades de la iglesia y del barrio ${suffix}.`,
      descriptionEn: null,
      // Mismo enlace que `alabanza`: aviso no bloqueante de enlace compartido.
      whatsappUrl: sharedLink,
      leaderVisible: false,
      sortOrder: 70,
      publicationState: 'published',
    },
    {
      key: 'misiones',
      nameEs: `Misiones ${suffix}`,
      nameEn: null,
      descriptionEs: `Apoyamos a quienes sirven fuera de la ciudad ${suffix}.`,
      descriptionEn: null,
      // Alta **sin** `leaderVisible` (P6-3): el servidor lo deja en `false`.
      leaderNameEs: `Diego Vargas ${suffix}`,
      leaderPhone: '+506 7000 5555',
      leaderVisible: false,
      omitLeaderVisibleKey: true,
      sortOrder: 80,
      publicationState: 'published',
    },
    {
      key: 'borrador-alabanza',
      nameEs: `Ensayo de alabanza en borrador ${suffix}`,
      nameEn: null,
      descriptionEs: `Todavía sin publicar ${suffix}.`,
      descriptionEn: null,
      leaderVisible: false,
      sortOrder: 90,
      publicationState: 'draft',
      withImage: true,
      imageAltEs: `Borrador de imagen ${suffix}`,
    },
    {
      key: 'borrador-sonido',
      nameEs: `Inventario de sonido en borrador ${suffix}`,
      nameEn: null,
      descriptionEs: `Todavía sin publicar ${suffix}.`,
      descriptionEn: null,
      leaderVisible: false,
      sortOrder: 100,
      publicationState: 'draft',
    },
  ];
}

export interface SeedOptions {
  /** Cuántos **publicados** sembrar (los N primeros del catálogo). Por defecto 8. */
  publishedCount?: number;
  /** Sembrar también los dos borradores (SC-002). Por defecto `true`. */
  withDrafts?: boolean;
}

/**
 * Siembra determinista del módulo: limpia lo anterior y crea los ministerios del
 * catálogo con un sufijo único por ejecución. Devuelve las fixtures con sus `id`
 * reales para que los specs aserten sobre datos que conocen.
 */
export async function seedMinisterios(
  page: Page,
  options: SeedOptions = {},
): Promise<MinisterioSeed> {
  const { publishedCount = 8, withDrafts = true } = options;

  await resetMinisterios(page);

  const suffix = uniqueSuffix();
  const all = catalog(suffix);
  // El catálogo está ordenado: los `published` primero (en orden manual) y los
  // borradores al final.
  const publishedFixtures = all.filter((fixture) => fixture.publicationState === 'published');
  const selected = all.filter((fixture) =>
    fixture.publicationState === 'draft'
      ? withDrafts
      : publishedFixtures.indexOf(fixture) < publishedCount,
  );

  const created: MinisterioFixture[] = [];
  for (const fixture of selected) {
    const imageFile = fixture.withImage ? await uploadMinistryImage(page) : undefined;
    const body: Record<string, unknown> = {
      nameEs: fixture.nameEs,
      nameEn: fixture.nameEn,
      descriptionEs: fixture.descriptionEs,
      descriptionEn: fixture.descriptionEn,
      scheduleEs: fixture.scheduleEs ?? null,
      placeEs: fixture.placeEs ?? null,
      leaderNameEs: fixture.leaderNameEs ?? null,
      leaderPhone: fixture.leaderPhone ?? null,
      leaderEmail: fixture.leaderEmail ?? null,
      whatsappUrl: fixture.whatsappUrl ?? '',
      sortOrder: fixture.sortOrder,
      imageFile: imageFile ?? null,
      imageAltEs: imageFile ? fixture.imageAltEs : null,
      publicationState: fixture.publicationState,
    };
    // «Oculto por defecto» (P6-3): el caso `misiones` viaja **sin** la clave.
    if (!fixture.omitLeaderVisibleKey) {
      body.leaderVisible = fixture.leaderVisible;
    }

    const result = (await expectOk(
      await apiSend(page, 'POST', '/api/v1/admin/ministerios', body),
      `POST ministerio «${fixture.key}»`,
    )) as { id: string; leaderVisible: boolean };

    created.push({
      ...fixture,
      id: result.id,
      imageFile,
      createdWithoutLeaderVisibleKey: fixture.omitLeaderVisibleKey,
      leaderVisible: result.leaderVisible,
    });
  }

  const published = created
    .filter((fixture) => fixture.publicationState === 'published')
    .sort((a, b) => a.sortOrder - b.sortOrder);
  const drafts = created.filter((fixture) => fixture.publicationState === 'draft');
  const draftWithImage = drafts.find((fixture) => fixture.imageFile);

  return {
    suffix,
    published,
    portada: published.slice(0, 5),
    drafts,
    hasDrafts: drafts.length > 0,
    sharedLink: `https://chat.whatsapp.com/alabanza-${suffix}`,
    draftImageFile: draftWithImage?.imageFile ?? '',
  };
}

/** Localiza una fixture del catálogo por su clave (falla si no se sembró). */
export function fixtureByKey(seed: MinisterioSeed, key: string): MinisterioFixture {
  const found = [...seed.published, ...seed.drafts].find((fixture) => fixture.key === key);
  if (!found) {
    throw new Error(`No se sembró el ministerio «${key}»`);
  }
  return found;
}
