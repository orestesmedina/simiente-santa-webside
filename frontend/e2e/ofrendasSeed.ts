import { expect, type APIResponse, type Page } from '@playwright/test';
import { apiSend, uniqueSuffix } from './helpers';

/**
 * Utilidades de siembra y limpieza de **ofrendas y diezmos** (F7) para los e2e
 * T733–T736. No es un `*.spec.ts`, así que Playwright no lo trata como prueba.
 *
 * La base es **compartida y no se limpia** entre corridas, así que cada spec
 * empieza con `resetOfrendas` (o con `seedOfrendas`, que lo hace por dentro):
 * borra **todos** los métodos —publicados y borradores— y pone los dos textos en
 * borrador, de modo que las aserciones —«0 borradores visibles», «exactamente 5
 * en el bloque», «sin nada publicado la ruta redirige», «solo textos publicados
 * mantiene viva la sección»— sean deterministas y repetibles. Por eso los specs
 * de F7 se ejecutan con `--workers=1`: si dos archivos mutaran el módulo a la
 * vez, se estorbarían (mismo motivo que F4/F6/F10).
 *
 * Los **números canónicos** del módulo (P7-2/P7-3/P7-4) son sagrados aquí: el
 * IBAN de Costa Rica tiene **22** caracteres (`CR` + 20 dígitos) y el SINPE se
 * guarda y publica normalizado como `+506` + 8 dígitos. Las fixtures se crean
 * escribiendo el número «como lo escribe el equipo» (con espacios y guiones, y
 * sin el `+506` en algún caso) y las aserciones comparan contra el valor
 * **normalizado** que devuelve la API (SC-009).
 *
 * El catálogo cubre, de una vez, los casos que piden las cuatro specs:
 *
 *   IBAN (kicker 01, IBAN primero, FR-004):
 *     1. `iban-nacional`      cuenta CR **con titular y moneda** (colones),
 *                             creada con espacios y guiones → normalizada.
 *     2. `iban-internacional` IBAN **internacional** de **24** caracteres (ES):
 *                             el caso de 320 px y de «nunca elipsis» (SC-009).
 *     3. `iban-sin-titular`   cuenta CR **sin titular** (US1 esc. 5: sin línea).
 *     4. `iban-dolares`       cuenta CR en dólares, con titular **solo en
 *                             español** (fallback `en → es`, SC-006).
 *   SINPE (kicker 02, tras IBAN):
 *     5. `sinpe-con-titular`  `+506` con titular.
 *     6. `sinpe-sin-titular`  sin titular, creado **sin** el `+506`.
 *     7. `sinpe-tercero`      `+506` con titular.
 *     8. `sinpe-cuarto`       sin titular (**nunca** con moneda, Q2).
 *   + `borrador-iban` / `borrador-sinpe`: métodos en borrador (SC-002).
 *
 * Con `{ publishedCount: 2 }` se siembran solo los dos primeros publicados
 * (escenario «exactos con menos, sin huecos» del bloque, SC-014); con
 * `publishedCount: 8` (por defecto) los ocho (tope 5 del bloque).
 *
 * Textos (singletons `intro`/`usage`, FR-021/Q5/Q11): `texts: 'published'`
 * (por defecto) publica los dos con tildes, `ñ`, `¿¡` y emojis en varios
 * párrafos; el introductorio **sí** tiene versión en inglés y el de uso **no**
 * (fallback `en → es` visible, SC-006). `texts: 'usage-draft'` deja el de uso en
 * borrador con métodos publicados (US3 esc. 7: sin hueco). `texts: 'draft'`
 * deja los dos en borrador (escenario «0 publicados»).
 *
 * `seedOfrendasPanelUsers` crea las tres cuentas del panel de T734: una con el
 * permiso **`donaciones`**, una con **otros** permisos de módulo (sin
 * `donaciones`) y una **sin** permisos de módulo (SC-005).
 */

/* ── Tipos ─────────────────────────────────────────────────────────────── */

export interface OfrendaMethodFixture {
  /** Clave estable del catálogo (para localizarla en las aserciones). */
  key: string;
  id: string;
  kind: 'iban' | 'sinpe';
  /** Cómo se escribe el número al crearlo (con espacios/guiones, sin `+506`…). */
  inputNumber: string;
  /** Número **normalizado** que publica la API: el valor exacto y copiable. */
  number: string;
  institution: string;
  holderEs?: string;
  holderEn?: string;
  /** Solo en IBAN; el SINPE **nunca** lo lleva (Q2). */
  currency?: 'CRC' | 'USD';
  sortOrder: number;
  publicationState: 'draft' | 'published';
}

export interface OfrendaTextFixture {
  key: 'intro' | 'usage';
  textEs: string;
  textEn: string | null;
  publicationState: 'draft' | 'published';
}

export type OfrendasTextsScenario = 'published' | 'usage-draft' | 'draft';

export interface SeedOfrendasOptions {
  /** Cuántos **publicados** sembrar (los N primeros del catálogo). Por defecto 8. */
  publishedCount?: number;
  /** Sembrar también los dos borradores de método (SC-002). Por defecto `true`. */
  withDrafts?: boolean;
  /** Estado de los dos textos de la sección. Por defecto `published`. */
  texts?: OfrendasTextsScenario;
}

export interface OfrendasSeed {
  suffix: string;
  /** Publicados en el **orden de la sección** (`iban → sinpe`, `sortOrder`). */
  published: OfrendaMethodFixture[];
  /** Los **5 primeros** publicados: los que pinta el bloque de la portada. */
  portada: OfrendaMethodFixture[];
  /** Borradores de método (invisibles al público). */
  drafts: OfrendaMethodFixture[];
  /** Los textos guardados (o `null` si el escenario los dejó sin escribir). */
  texts: { intro: OfrendaTextFixture | null; usage: OfrendaTextFixture | null };
}

interface AdminDonacionItem {
  id: string;
  kind: 'iban' | 'sinpe';
  number: string;
  institution: string;
  publicationState: 'draft' | 'published';
}

interface AdminTextos {
  intro: OfrendaTextFixture | null;
  usage: OfrendaTextFixture | null;
}

/** El catálogo base, sin `id` (se completa al crear). */
type Fixture = Omit<OfrendaMethodFixture, 'id'>;

export interface PanelUser {
  email: string;
  password: string;
  roleId: string;
}

export interface OfrendasPanelUsers {
  /** Cuenta con el permiso `donaciones` (puede entrar al módulo). */
  donaciones: PanelUser;
  /** Cuenta con **otros** permisos de módulo (no `donaciones`): 403 cruzado. */
  otherModules: PanelUser;
  /** Cuenta sin ningún permiso de módulo. */
  none: PanelUser;
}

/* ── Utilidades internas ───────────────────────────────────────────────── */

async function expectOk(response: APIResponse, label: string): Promise<unknown> {
  expect(response.ok(), `${label} → ${response.status()}`).toBe(true);
  return response.json();
}

/** Listado completo (tope del contrato, 100) de métodos del panel. */
async function listAll(page: Page): Promise<AdminDonacionItem[]> {
  const body = (await expectOk(
    await apiSend(page, 'GET', '/api/v1/admin/donaciones?limit=100'),
    'GET donaciones',
  )) as { items: AdminDonacionItem[] };
  return body.items;
}

/** Los dos textos del panel (singletons; `null` hasta el primer guardado). */
export async function getOfrendaTextos(page: Page): Promise<AdminTextos> {
  return (await expectOk(
    await apiSend(page, 'GET', '/api/v1/admin/donaciones/textos'),
    'GET textos de donaciones',
  )) as AdminTextos;
}

/**
 * Busca un rol por su identificador recorriendo el listado paginado: la base
 * compartida acumula roles de corridas anteriores y el del administrador, creado
 * el primero, queda fuera de la primera página.
 */
async function findRoleById(
  page: Page,
  roleId: string,
): Promise<{ id: string; permissions: string[] } | undefined> {
  const limit = 100;
  for (let offset = 0; offset < 2000; offset += limit) {
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

/**
 * Garantiza que el rol de la cuenta administradora tiene el permiso
 * **`donaciones`** (ya sembrado por F2, P7-6). La base de desarrollo puede venir
 * de una corrida anterior a la activación del módulo; el `Resolver` de sesión lee
 * los permisos vigentes en cada petición, así que basta con completar el rol
 * (idempotente: solo actúa si falta).
 */
export async function ensureDonacionesPermission(page: Page): Promise<void> {
  const session = (await expectOk(
    await apiSend(page, 'GET', '/api/v1/auth/session'),
    'GET session',
  )) as { roleId: string; permissions: string[] };

  if (session.permissions.includes('donaciones')) {
    return;
  }

  const role = await findRoleById(page, session.roleId);
  if (!role) {
    throw new Error('No se encontró el rol de la cuenta administradora');
  }
  const permissions = [...new Set([...role.permissions, 'donaciones'])];
  expect(
    (await apiSend(page, 'PATCH', `/api/v1/admin/roles/${role.id}`, { permissions })).ok(),
    'el rol del administrador con el permiso donaciones',
  ).toBe(true);
}

/* ── Reset ─────────────────────────────────────────────────────────────── */

/**
 * Deja el módulo en un estado conocido: **sin ningún método** (ni publicado ni
 * borrador) y con los dos textos en **borrador** (los que existan se conservan
 * en contenido, solo se retiran). Es la base de la repetibilidad de los cuatro
 * specs de F7 y, además, el escenario «nada publicado» de la redirección de
 * `/ofrendas` (SC-010/P7-21) y del bloque ausente de la portada (SC-014).
 */
export async function resetOfrendas(page: Page): Promise<void> {
  await ensureDonacionesPermission(page);

  const items = await listAll(page);
  for (const item of items) {
    expect(
      (await apiSend(page, 'DELETE', `/api/v1/admin/donaciones/${item.id}`)).status(),
      `DELETE donaciones/${item.id}`,
    ).toBe(204);
  }

  // Los textos no se borran (P7-17): se retiran a borrador conservando su cuerpo.
  const textos = await getOfrendaTextos(page);
  for (const key of ['intro', 'usage'] as const) {
    const texto = textos[key];
    if (texto) {
      await expectOk(
        await apiSend(page, 'PUT', `/api/v1/admin/donaciones/textos/${key}`, {
          textEs: texto.textEs,
          textEn: texto.textEn,
          publicationState: 'draft',
        }),
        `PUT texto ${key} a borrador`,
      );
    }
  }
}

/* ── Catálogo ──────────────────────────────────────────────────────────── */

/** Números canónicos (normalizados) que las specs usan en sus aserciones. */
export const CANONICAL = {
  /** IBAN de Costa Rica: **22** caracteres (FR-015). */
  IBAN_CR: 'CR05015202001234567890',
  /** IBAN internacional: **24** caracteres (ES) — el caso de 320 px (SC-009). */
  IBAN_ES: 'ES9121000418450200051332',
  /** IBAN de Costa Rica sin titular (US1 esc. 5). */
  IBAN_CR_SIN_TITULAR: 'CR01010222003344556677',
  /** IBAN de Costa Rica en dólares. */
  IBAN_CR_DOLARES: 'CR12015100000011223344',
  /** SINPE normalizado (FR-016). */
  SINPE_CON_TITULAR: '+50688888888',
  SINPE_SIN_TITULAR: '+50670001111',
} as const;

/**
 * Catálogo del módulo con el sufijo de la ejecución. Los publicados van primero
 * (en orden manual por grupo: IBAN por `sortOrder` y luego SINPE) y los
 * borradores al final.
 */
function catalog(suffix: string): Fixture[] {
  return [
    /* ── Cuentas IBAN (publicadas) ─────────────────────────────────── */
    {
      key: 'iban-nacional',
      kind: 'iban',
      inputNumber: 'CR05 0152-0200-1234-5678-90',
      number: CANONICAL.IBAN_CR,
      institution: `Banco Nacional de Costa Rica ${suffix}`,
      holderEs: `Iglesia Simiente Santa ${suffix}`,
      holderEn: `Simiente Santa Church ${suffix}`,
      currency: 'CRC',
      sortOrder: 10,
      publicationState: 'published',
    },
    {
      key: 'iban-internacional',
      kind: 'iban',
      // 24 caracteres (ES): el número más largo del catálogo, el que a 320 px
      // debe reacomodarse **sin perder ni un carácter** (SC-007/SC-009).
      inputNumber: 'ES91 2100 0418 4502 0005 1332',
      number: CANONICAL.IBAN_ES,
      institution: `Banco Internacional ${suffix}`,
      holderEs: `Iglesia Simiente Santa ${suffix}`,
      holderEn: `Simiente Santa Church ${suffix}`,
      currency: 'USD',
      sortOrder: 20,
      publicationState: 'published',
    },
    {
      key: 'iban-sin-titular',
      kind: 'iban',
      inputNumber: 'CR01 0102 2200 3344 5566 77',
      number: CANONICAL.IBAN_CR_SIN_TITULAR,
      institution: `Banco Popular ${suffix}`,
      // Sin titular: la ficha se compone **sin** línea (US1 esc. 5).
      currency: 'CRC',
      sortOrder: 30,
      publicationState: 'published',
    },
    {
      key: 'iban-dolares',
      kind: 'iban',
      inputNumber: 'CR12 0151 0000 0011 2233 44',
      number: CANONICAL.IBAN_CR_DOLARES,
      institution: `BAC San José ${suffix}`,
      // Titular **solo en español**: en inglés cae al español (FR-008/SC-006).
      holderEs: `Asociación Simiente ${suffix}`,
      currency: 'USD',
      sortOrder: 40,
      publicationState: 'published',
    },

    /* ── Números de SINPE Móvil (publicados) ───────────────────────── */
    {
      key: 'sinpe-con-titular',
      kind: 'sinpe',
      inputNumber: '8888-8888',
      number: CANONICAL.SINPE_CON_TITULAR,
      institution: `Banco Promerica ${suffix}`,
      holderEs: `Tesorería Simiente Santa ${suffix}`,
      holderEn: `Simiente Santa Treasury ${suffix}`,
      sortOrder: 10,
      publicationState: 'published',
    },
    {
      key: 'sinpe-sin-titular',
      kind: 'sinpe',
      // Escribido **sin** el `+506`: el servidor lo normaliza (FR-016).
      inputNumber: '7000 1111',
      number: CANONICAL.SINPE_SIN_TITULAR,
      institution: `Banco de Costa Rica ${suffix}`,
      sortOrder: 20,
      publicationState: 'published',
    },
    {
      key: 'sinpe-tercero',
      kind: 'sinpe',
      inputNumber: '6123-4567',
      number: '+50661234567',
      institution: `Coopealianza ${suffix}`,
      holderEs: `Misiones Simiente ${suffix}`,
      sortOrder: 30,
      publicationState: 'published',
    },
    {
      key: 'sinpe-cuarto',
      kind: 'sinpe',
      inputNumber: '+506 8777 6655',
      number: '+50687776655',
      institution: `Banco General ${suffix}`,
      // Sin titular y **nunca** con moneda (Q2).
      sortOrder: 40,
      publicationState: 'published',
    },

    /* ── Borradores (invisibles por cualquier vía, SC-002) ─────────── */
    {
      key: 'borrador-iban',
      kind: 'iban',
      inputNumber: 'CR99 0152 0200 1234 5699 99',
      number: 'CR99015202001234569999',
      institution: `Banco en borrador ${suffix}`,
      holderEs: `Cuenta en borrador ${suffix}`,
      currency: 'CRC',
      sortOrder: 90,
      publicationState: 'draft',
    },
    {
      key: 'borrador-sinpe',
      kind: 'sinpe',
      inputNumber: '60111111',
      number: '+50660111111',
      institution: `SINPE en borrador ${suffix}`,
      sortOrder: 100,
      publicationState: 'draft',
    },
  ];
}

/* ── Siembra ───────────────────────────────────────────────────────────── */

/** Crea un método y devuelve su fixture con el `id` real. */
async function createMethod(page: Page, fixture: Fixture): Promise<OfrendaMethodFixture> {
  const body: Record<string, unknown> = {
    kind: fixture.kind,
    number: fixture.inputNumber,
    institution: fixture.institution,
    holderEs: fixture.holderEs ?? null,
    holderEn: fixture.holderEn ?? null,
    sortOrder: fixture.sortOrder,
    publicationState: fixture.publicationState,
  };
  if (fixture.kind === 'iban') {
    body.currency = fixture.currency;
  }

  const result = (await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/donaciones', body),
    `POST donación «${fixture.key}»`,
  )) as { id: string; number: string };

  return { ...fixture, id: result.id, number: result.number };
}

/** Guarda (crea o reemplaza) un texto y devuelve su fixture con el `id` real. */
async function putTexto(
  page: Page,
  key: 'intro' | 'usage',
  fixture: Omit<OfrendaTextFixture, 'key'>,
): Promise<OfrendaTextFixture> {
  const result = (await expectOk(
    await apiSend(page, 'PUT', `/api/v1/admin/donaciones/textos/${key}`, {
      textEs: fixture.textEs,
      textEn: fixture.textEn,
      publicationState: fixture.publicationState,
    }),
    `PUT texto ${key}`,
  )) as OfrendaTextFixture;
  return { ...result, key };
}

/** Los dos textos del catálogo, con el sufijo de la ejecución. */
function textFixtures(suffix: string): {
  intro: Omit<OfrendaTextFixture, 'key'>;
  usage: Omit<OfrendaTextFixture, 'key'>;
} {
  return {
    intro: {
      textEs: `Tu aporte sostiene la obra de la iglesia 🙏\n¿Cómo donar? Transfiere a cualquiera de estas cuentas o envía un SINPE Móvil: es rápido y seguro. ${suffix}`,
      textEn: `Your offering sustains the church's work 🙏\nHow to give? Transfer to any of these accounts or send a SINPE Móvil: quick and safe. ${suffix}`,
      publicationState: 'published',
    },
    usage: {
      // Solo en español: en inglés cae al español sin hueco (FR-008/SC-006).
      textEs: `¿En qué se usan? En la obra social, el mantenimiento del templo y las misiones.\n¡Gracias por tu apoyo! 🎉 ${suffix}`,
      textEn: null,
      publicationState: 'published',
    },
  };
}

/**
 * Siembra determinista del módulo: limpia lo anterior y crea los métodos y los
 * textos del catálogo con un sufijo único por ejecución. Devuelve las fixtures
 * (con sus `id` reales) en el orden de la sección.
 */
export async function seedOfrendas(
  page: Page,
  options: SeedOfrendasOptions = {},
): Promise<OfrendasSeed> {
  const { publishedCount = 8, withDrafts = true, texts = 'published' } = options;

  await resetOfrendas(page);

  const suffix = uniqueSuffix();
  const all = catalog(suffix);
  const publishedFixtures = all.filter((fixture) => fixture.publicationState === 'published');
  const selected = all.filter((fixture) =>
    fixture.publicationState === 'draft'
      ? withDrafts
      : publishedFixtures.indexOf(fixture) < publishedCount,
  );

  const created: OfrendaMethodFixture[] = [];
  for (const fixture of selected) {
    created.push(await createMethod(page, fixture));
  }

  const textsCatalog = textFixtures(suffix);
  const stored: OfrendasSeed['texts'] = { intro: null, usage: null };
  if (texts === 'published') {
    stored.intro = await putTexto(page, 'intro', textsCatalog.intro);
    stored.usage = await putTexto(page, 'usage', textsCatalog.usage);
  } else if (texts === 'usage-draft') {
    stored.intro = await putTexto(page, 'intro', textsCatalog.intro);
    stored.usage = await putTexto(page, 'usage', {
      ...textsCatalog.usage,
      publicationState: 'draft',
    });
  } else {
    stored.intro = await putTexto(page, 'intro', {
      ...textsCatalog.intro,
      publicationState: 'draft',
    });
    stored.usage = await putTexto(page, 'usage', {
      ...textsCatalog.usage,
      publicationState: 'draft',
    });
  }

  // Orden de la sección: `iban` antes que `sinpe` y, dentro de cada grupo,
  // `sortOrder` (FR-004/P7-20). Los `sortOrder` del catálogo son distintos, así
  // que el orden es determinista sin depender del `id`.
  const published = created
    .filter((fixture) => fixture.publicationState === 'published')
    .sort((a, b) => {
      if (a.kind !== b.kind) {
        return a.kind === 'iban' ? -1 : 1;
      }
      return a.sortOrder - b.sortOrder;
    });
  const drafts = created.filter((fixture) => fixture.publicationState === 'draft');

  return {
    suffix,
    published,
    portada: published.slice(0, 5),
    drafts,
    texts: stored,
  };
}

/** Localiza una fixture del catálogo por su clave (falla si no se sembró). */
export function fixtureByKey(seed: OfrendasSeed, key: string): OfrendaMethodFixture {
  const found = [...seed.published, ...seed.drafts].find((fixture) => fixture.key === key);
  if (!found) {
    throw new Error(`No se sembró el método de donación «${key}»`);
  }
  return found;
}

/* ── Cuentas del panel (SC-005, T734) ──────────────────────────────────── */

/** Crea un rol y una cuenta con esos permisos; la contraseña es inicial. */
async function createPanelUser(
  page: Page,
  label: string,
  permissions: string[],
  suffix: string,
): Promise<PanelUser> {
  const role = (await expectOk(
    await apiSend(page, 'POST', '/api/v1/admin/roles', {
      name: `${label} ${suffix}`,
      permissions,
    }),
    `POST rol «${label}»`,
  )) as { id: string };

  const email = `${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}-${suffix}@ejemplo.com`;
  const password = 'Inicial.2026';
  expect(
    (
      await apiSend(page, 'POST', '/api/v1/admin/usuarios', {
        firstName: 'Cuenta',
        lastName: label,
        email,
        phone: '612 000 555',
        roleId: role.id,
        password,
      })
    ).ok(),
    `POST usuario «${label}»`,
  ).toBe(true);

  return { email, password, roleId: role.id };
}

/**
 * Crea las tres cuentas del panel que necesitan las aserciones de permisos de
 * T734 (SC-005): con `donaciones`, con **otros** permisos de módulo y **sin**
 * permisos de módulo. Todas se crean con la contraseña inicial `Inicial.2026`,
 * que el primer acceso obliga a cambiar.
 *
 * Nota: el contrato exige **al menos un permiso** por rol (`RoleCreateInput`),
 * así que la cuenta «sin permisos» lleva un permiso que **no** corresponde a
 * ningún módulo del panel (`noticias`): no ve ningún acceso rápido ni entra en
 * ningún módulo.
 */
export async function seedOfrendasPanelUsers(
  page: Page,
  suffix: string = uniqueSuffix(),
): Promise<OfrendasPanelUsers> {
  return {
    donaciones: await createPanelUser(page, 'Solo donaciones', ['donaciones'], suffix),
    otherModules: await createPanelUser(page, 'Otros modulos', ['eventos', 'noticias'], suffix),
    none: await createPanelUser(page, 'Sin modulos', ['noticias'], suffix),
  };
}
