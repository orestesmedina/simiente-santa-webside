import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { expect, type Page } from '@playwright/test';
import { API_URL, uniqueSuffix } from './helpers';

/**
 * Utilidades de siembra y limpieza de **medios** (F9) para los e2e T940–T943.
 * No es un `*.spec.ts`, así que Playwright no lo trata como prueba.
 *
 * ## Por qué se siembra **por SQL** (patrón nuevo declarado en RG9-7)
 *
 * F9 **no tiene alta manual** de episodios (FR-019): la única puerta de entrada
 * de contenido es «Buscar episodios nuevos», que lee los feeds reales de YouTube
 * y del podcast. Por eso los e2e **no pueden crear fichas por API** (lo harían
 * dependiendo de la red y del contenido publicado hoy en esos servicios) y los
 * fixtures se insertan **directamente en PostgreSQL** —el contenedor `db` de
 * `docker compose`, con el mismo SQL documentado en `quickstart.md` §5 y §8—.
 * Es el **coste declarado** del plan: los datos de partida de los e2e de F9 no
 * se crean por la interfaz. `mediosSeed.ts` **nunca** pulsa «Buscar episodios
 * nuevos» (la importación se cubre en Go con dobles, T920/T921, y a mano en
 * `quickstart.md` §3–§4).
 *
 * ## Determinismo (y el segundo coste declarado: la base es compartida)
 *
 * La base de desarrollo trae **episodios reales** importados de una búsqueda
 * anterior, y la sección `/medios`, el bloque de la portada y el listado del
 * panel **no tienen filtro por dueño**: cualquier ficha publicable de la base
 * aparece en las aserciones. Para que los totales («exactamente 5 en el bloque»,
 * «12 y luego 14 con Cargar más», «0 publicables → redirección») sean
 * deterministas y **no dependan de esos datos**:
 *
 *   1. se **ocultan temporalmente** las fichas publicables que no son nuestras
 *      (`visibility = 'hidden'`, el mismo estado que usa el equipo) y se anota
 *      su `id` en un archivo temporal (`/tmp`) además de en memoria;
 *   2. al terminar, `resetMedios` **borra lo nuestro** y **restaura** a `visible`
 *      las fichas anotadas.
 *
 * Nada se borra ni se altera de forma permanente, y el archivo temporal permite
 * que una corrida interrumpida se repare sola en la siguiente (el reset lee el
 * archivo si no tiene la lista en memoria). Las filas propias usan **ids fijos**
 * (los mismos en cada corrida): se borran antes de insertar, así que no se
 * acumulan y las aserciones pueden nombrarlas por su título.
 *
 * ## La carátula tiene que existir de verdad
 *
 * `GET /api/v1/media/<file>` solo sirve el archivo si una ficha **publicable** lo
 * referencia (`IsFilePublished`, R9-8). Para que las fichas con carátula carguen
 * una imagen real (y no el marcador neutro del `onError`), la siembra sube una
 * vez la imagen del repositorio con `POST /api/v1/admin/portada/imagenes` (el
 * endpoint de F3) y usa el nombre devuelto como `cover_file`. **Solo las fichas
 * publicables llevan carátula**: en una ficha oculta o bloqueada el archivo no se
 * sirve (404 por diseño) y el panel mostraría una imagen rota. El archivo subido
 * queda en el volumen de subidas (mismo comportamiento que las siembras de
 * F7/F10, que también suben su imagen).
 *
 * ## Catálogo (qué caso cubre cada fixture)
 *
 * **Publicables** (en el orden de la sección: más reciente primero; los 8 del
 * catálogo base, y con `publishedCount > 8` se añaden «de relleno» para el
 * escenario de «Cargar más»):
 *
 *   1. `dos-enlaces`    los **dos** enlaces (video + audio, `origin='both'`),
 *                       carátula y descripción es/en (SC-006/SC-011).
 *   2. `solo-video`     **un solo** enlace (video): la ficha no deja línea vacía.
 *   3. `solo-audio`     **un solo** enlace (audio).
 *   4. `sin-caratula`   **sin** carátula → marcador neutro (D-4).
 *   5. `sin-ingles`     descripción **sin versión en inglés** → fallback `en → es`
 *                       visible (FR-025/SC-011).
 *   6. `sin-descripcion` **sin** descripción → la ficha **omite** la línea
 *                       (FR-020/precisión 3).
 *   7. `vivo-fuera`     «ya no aparece en el feed» **abierta** y sigue **visible**
 *                       (`availability_result='available'`): SC-014/FR-017
 *                       (quickstart §8(a), id `3333…`).
 *   8. `inconcluso`     «ya no aparece en el feed» con `inconclusive`: nunca
 *                       oculta por falta de certeza (SC-014).
 *   + `relleno-NN`      solo con `publishedCount > 8` (tope 5 del bloque y
 *                       «Cargar más»).
 *
 * **No publicables** (siempre sembradas: son el objeto del panel y la prueba de
 * que nada de eso sale al público, SC-002):
 *
 *   · `par-union-base` / `par-union-otra`: par «posible duplicado» **abierto**
 *     (base `visible` + adicional `hidden`, `duplicate_of_id` recíproco) para el
 *     flujo de **unión** con doble confirmación (quickstart §5, ids `1111…`/`2222…`).
 *   · `par-distintos-base` / `par-distintos-otra`: otro par abierto, para resolver
 *     **«no son el mismo sermón»** (US3 esc. 2) sin gastar el anterior.
 *   · `par-mixto-base` / `par-mixto-otra`: par abierto **y** marca de
 *     disponibilidad abierta en la base → los **dos ejes** conviven y se
 *     resuelven por separado (P9-19).
 *   · `oculta-equipo`: oculta por el equipo, sin marcas (SC-002/SC-007).
 *   · `marca-unavailable`: el **estado resultante de la precedencia A3** (fase 5):
 *     oculta con «no disponible» **abierta** (`availability_result='unavailable'`)
 *     → el panel la presenta y «Mostrar de nuevo» funciona (quickstart §8(b), id
 *     `4444…`). La precedencia en sí la fija T921.
 *   · `marca-resuelta`: marca de disponibilidad **ya resuelta** → la fila queda
 *     «Resuelta» y no se reabre por la misma causa (FR-017/US3 esc. 6).
 *   · `absorbida`: absorbida por una unión (`merged_into_id`): invisible por
 *     cualquier vía (FR-019/SC-002).
 *
 * ## Cuentas del panel (SC-009)
 *
 * `seedMediosPanelUsers` crea las tres cuentas de T941: una con el permiso
 * **`medios`**, una con **otros** permisos de módulo (sin `medios`) y una **sin**
 * permisos de módulo (se le da `noticias`, que no abre ningún módulo del panel,
 * porque el contrato exige al menos un permiso por rol).
 */

/* ── Tipos ─────────────────────────────────────────────────────────────── */

export type ProveedorMedio = 'youtube' | 'podcast';
export type OrigenMedioFixture = 'video' | 'audio' | 'both';
export type VisibilidadFixture = 'visible' | 'hidden';
export type MarcaDisponibilidadFixture = 'unavailable' | 'missing_from_feed';
export type ResultadoDisponibilidadFixture = 'available' | 'unavailable' | 'inconclusive';

export interface FuenteFixture {
  provider: ProveedorMedio;
  /** `yt:videoId` (11 caracteres) o `<guid>` del podcast (R9-5). */
  externalId: string;
  /** Título tal como lo trae **ese** feed (procedencia, FR-003). */
  feedTitle: string;
  url: string;
}

export interface MedioFixture {
  /** Clave estable del catálogo (para localizarla en las aserciones). */
  key: string;
  id: string;
  title: string;
  /** Instante RFC 3339 (el día público se resuelve en `America/Costa_Rica`). */
  publishedAt: string;
  descriptionEs: string | null;
  descriptionEn: string | null;
  origin: OrigenMedioFixture;
  visibility: VisibilidadFixture;
  /** `true` si la ficha lleva carátula (solo las publicables la llevan). */
  hasCover: boolean;
  sources: FuenteFixture[];
  /** Par «posible duplicado» **abierto** con la ficha indicada (si lo tiene). */
  duplicateOfKey?: string;
  /** Marca de disponibilidad **abierta** (si la tiene). */
  availabilityMark?: MarcaDisponibilidadFixture;
  /** Resultado de la comprobación que motivó la marca (si lo tiene). */
  availabilityResult?: ResultadoDisponibilidadFixture;
  /** Marca de disponibilidad **ya resuelta** (si la tiene). */
  availabilityResolved?: boolean;
  /** `true` si está absorbida por una unión (`merged_into_id`). */
  merged?: boolean;
}

export interface SeedMediosOptions {
  /**
   * Cuántos **publicables** sembrar, tomados del principio del catálogo (más
   * reciente primero). Por defecto **8** (los del catálogo base). `0` deja la
   * sección sin nada publicable (escenario de redirección, SC-016) y por encima
   * de 8 añade fichas «de relleno» (escenario de «Cargar más», FR-006).
   */
  publishedCount?: number;
}

export interface MediosSeed {
  /** Sufijo único de la corrida (para textos que deban ser irrepetibles). */
  suffix: string;
  /** Nombre del archivo de carátula subido (ausente si la subida no se hizo). */
  coverFile?: string;
  /** Publicables en el **orden de la sección** (más reciente primero). */
  publicables: MedioFixture[];
  /** Los **5 primeros** publicables: los que pinta el bloque de la portada. */
  portada: MedioFixture[];
  /** Fichas no publicables (ocultas, marcadas, en par abierto o absorbidas). */
  ocultas: MedioFixture[];
  /** Todas las fichas sembradas, por clave. */
  porClave: Record<string, MedioFixture>;
}

export interface PanelUser {
  email: string;
  password: string;
  roleId: string;
}

export interface MediosPanelUsers {
  /** Cuenta con el permiso `medios` (puede entrar al módulo). */
  medios: PanelUser;
  /** Cuenta con **otros** permisos de módulo (no `medios`): 403 cruzado. */
  otherModules: PanelUser;
  /** Cuenta sin ningún permiso de módulo. */
  none: PanelUser;
}

/* ── Acceso a PostgreSQL (contenedor `db`, RG9-7) ──────────────────────── */

/** `docker-compose.yml` del proyecto (se busca hacia arriba desde el cwd). */
function composeFile(): string {
  let dir = process.cwd();
  for (let depth = 0; depth < 4; depth += 1) {
    const candidate = join(dir, 'docker-compose.yml');
    if (existsSync(candidate)) {
      return candidate;
    }
    dir = dirname(dir);
  }
  throw new Error(
    'No se encontró docker-compose.yml: los e2e de F9 siembran por SQL en el contenedor `db` (RG9-7)',
  );
}

/**
 * Ejecuta SQL en la base de desarrollo (`docker compose exec -T db psql`). Con
 * `ON_ERROR_STOP=1` un error aborta y `execFileSync` lanza: la siembra nunca
 * sigue con la base a medias. Los valores se interpolan ya escapados
 * (`sqlText`), nunca desde datos del usuario.
 */
function psql(sql: string): string {
  return execFileSync(
    'docker',
    [
      'compose',
      '-f',
      composeFile(),
      'exec',
      '-T',
      'db',
      'psql',
      '-U',
      'app',
      '-d',
      'app',
      '-v',
      'ON_ERROR_STOP=1',
      '-A',
      '-t',
      '-c',
      sql,
    ],
    { encoding: 'utf8' },
  );
}

/** Literal de texto SQL (comillas simples duplicadas). */
function sqlText(value: string): string {
  return `'${value.replace(/'/g, "''")}'`;
}

/** Literal de texto o `NULL`. */
function sqlNullable(value: string | null): string {
  return value === null ? 'NULL' : sqlText(value);
}

/** Literal de instante (`TIMESTAMPTZ`). */
function sqlTimestamp(value: string): string {
  return `${sqlText(value)}::timestamptz`;
}

/** Lista `ARRAY[...]::uuid[]` con los ids indicados. */
function sqlUuidArray(ids: string[]): string {
  return `ARRAY[${ids.map((id) => `${sqlText(id)}::uuid`).join(', ')}]`;
}

/* ── Catálogo ──────────────────────────────────────────────────────────── */

/** Ids fijos del catálogo (mismos en cada corrida: se borran antes de insertar). */
const IDS = {
  parUnionBase: '11111111-1111-4111-8111-111111111111',
  parUnionOtra: '22222222-2222-4222-8222-222222222222',
  vivoFuera: '33333333-3333-4333-8333-333333333333',
  marcaUnavailable: '44444444-4444-4444-8444-444444444444',
  dosEnlaces: '55555555-5555-4555-8555-555555555555',
  soloVideo: '66666666-6666-4666-8666-666666666666',
  soloAudio: '77777777-7777-4777-8777-777777777777',
  sinCaratula: '88888888-8888-4888-8888-888888888888',
  sinIngles: '99999999-9999-4999-8999-999999999999',
  sinDescripcion: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
  inconcluso: 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
  parDistintosBase: 'cccccccc-cccc-4ccc-8ccc-cccccccccccc',
  parDistintosOtra: 'dddddddd-dddd-4ddd-8ddd-dddddddddddd',
  parMixtoBase: 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee',
  parMixtoOtra: 'f0f0f0f0-f0f0-40f0-80f0-f0f0f0f0f0f0',
  ocultaEquipo: 'f1f1f1f1-f1f1-41f1-81f1-f1f1f1f1f1f1',
  marcaResuelta: 'f3f3f3f3-f3f3-43f3-83f3-f3f3f3f3f3f3',
  absorbida: 'f4f4f4f4-f4f4-44f4-84f4-f4f4f4f4f4f4',
} as const;

/** Id de una ficha «de relleno» (`nn` = 1…24), para el escenario de «Cargar más». */
function rellenoId(nn: number): string {
  return `f4000000-0000-4000-8000-0000000000${String(nn).padStart(2, '0')}`;
}

/** Tope de fichas «de relleno» que el catálogo sabe generar. */
const MAX_RELLENO = 24;

/** Base del enlace de YouTube con un `videoId` de 11 caracteres (CHECK de la BD). */
function youtubeUrl(videoId: string): string {
  return `https://www.youtube.com/watch?v=${videoId}`;
}

/** Enlace de un episodio del podcast (CHECK `podcasters.spotify.com`). */
function podcastUrl(slug: string): string {
  return `https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/${slug}`;
}

/**
 * Los **8 publicables** del catálogo base, en el orden de la sección (más
 * reciente primero). Las fechas son de 2026-09/2026-10 (recientes respecto a los
 * datos reales de la base, y distintas entre sí para que el orden sea estable).
 */
function publicablesBase(): MedioFixture[] {
  return [
    {
      key: 'dos-enlaces',
      id: IDS.dosEnlaces,
      title: 'Prédica de prueba con video y audio',
      publishedAt: '2026-10-05T15:00:00Z',
      descriptionEs: 'Prédica con los dos enlaces: ¡qué día! 🙌 Se abre en YouTube y en Spotify.',
      descriptionEn: 'Sermon with both links: what a day! 🙌 It opens in YouTube and Spotify.',
      origin: 'both',
      visibility: 'visible',
      hasCover: true,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2edosEnl01',
          feedTitle: 'Prédica de prueba con video y audio',
          url: youtubeUrl('e2edosEnl01'),
        },
        {
          provider: 'podcast',
          externalId: 'e2e-guid-dos-enlaces',
          feedTitle: 'Prédica de prueba con video y audio (podcast)',
          url: podcastUrl('predica-de-prueba-con-video-y-audio-e2e001'),
        },
      ],
    },
    {
      key: 'solo-video',
      id: IDS.soloVideo,
      title: 'Sermón de prueba solo en video',
      publishedAt: '2026-10-04T15:00:00Z',
      descriptionEs: 'Este sermón solo tiene video: la ficha no deja ninguna línea vacía.',
      descriptionEn: null,
      origin: 'video',
      visibility: 'visible',
      hasCover: true,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2esoloVid1',
          feedTitle: 'Sermón de prueba solo en video',
          url: youtubeUrl('e2esoloVid1'),
        },
      ],
    },
    {
      key: 'solo-audio',
      id: IDS.soloAudio,
      title: 'Podcast de prueba solo en audio',
      publishedAt: '2026-10-03T15:00:00Z',
      descriptionEs: 'Este episodio solo tiene audio: un único enlace, sin huecos.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'visible',
      hasCover: true,
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-solo-audio',
          feedTitle: 'Podcast de prueba solo en audio',
          url: podcastUrl('podcast-de-prueba-solo-en-audio-e2e002'),
        },
      ],
    },
    {
      key: 'sin-caratula',
      id: IDS.sinCaratula,
      title: 'Episodio de prueba sin carátula',
      publishedAt: '2026-10-02T15:00:00Z',
      descriptionEs: 'Sin carátula en el feed: la ficha pinta el marcador neutro, nunca un hueco.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'visible',
      hasCover: false,
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-sin-caratula',
          feedTitle: 'Episodio de prueba sin carátula',
          url: podcastUrl('episodio-de-prueba-sin-caratula-e2e003'),
        },
      ],
    },
    {
      key: 'sin-ingles',
      id: IDS.sinIngles,
      title: 'Episodio de prueba sin versión en inglés',
      publishedAt: '2026-10-01T15:00:00Z',
      descriptionEs: 'Solo hay versión en español: en inglés se ve este mismo texto.',
      descriptionEn: null,
      origin: 'both',
      visibility: 'visible',
      hasCover: true,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2esinIngl1',
          feedTitle: 'Episodio de prueba sin versión en inglés',
          url: youtubeUrl('e2esinIngl1'),
        },
        {
          provider: 'podcast',
          externalId: 'e2e-guid-sin-ingles',
          feedTitle: 'Episodio de prueba sin versión en inglés (podcast)',
          url: podcastUrl('episodio-de-prueba-sin-version-en-ingles-e2e004'),
        },
      ],
    },
    {
      key: 'sin-descripcion',
      id: IDS.sinDescripcion,
      title: 'Episodio de prueba sin descripción',
      publishedAt: '2026-09-30T15:00:00Z',
      descriptionEs: null,
      descriptionEn: null,
      origin: 'video',
      visibility: 'visible',
      hasCover: true,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2esinDesc1',
          feedTitle: 'Episodio de prueba sin descripción',
          url: youtubeUrl('e2esinDesc1'),
        },
      ],
    },
    {
      key: 'vivo-fuera',
      id: IDS.vivoFuera,
      title: 'Vivo pero fuera del feed',
      publishedAt: '2026-09-29T15:00:00Z',
      descriptionEs: 'Su enlace sigue disponible.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'visible',
      hasCover: true,
      // SQL de `quickstart.md` §8(a): salió de la ventana del feed pero su enlace
      // responde → sigue VISIBLE y solo queda marcado (FR-017/SC-014).
      availabilityMark: 'missing_from_feed',
      availabilityResult: 'available',
      sources: [
        {
          provider: 'podcast',
          externalId: 'guid-vivo-1',
          feedTitle: 'Vivo pero fuera del feed',
          url: 'https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/Potencial-sin-Resultados-La-Leccin-de-la-Higuera-e3jsj3n',
        },
      ],
    },
    {
      key: 'inconcluso',
      id: IDS.inconcluso,
      title: 'Comprobación sin respuesta',
      publishedAt: '2026-09-28T15:00:00Z',
      descriptionEs: 'La comprobación de su enlace no se pudo completar: sigue visible.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'visible',
      hasCover: true,
      // SC-014: la comprobación que no se completa **nunca** oculta (solo marca).
      availabilityMark: 'missing_from_feed',
      availabilityResult: 'inconclusive',
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-inconcluso',
          feedTitle: 'Comprobación sin respuesta',
          url: podcastUrl('comprobacion-sin-respuesta-e2e005'),
        },
      ],
    },
  ];
}

/**
 * Ficha «de relleno» `nn` (1…24): publicable y sin marcas, para engordar la
 * sección por encima de la tanda de 12 («Cargar más», FR-006) y para el tope 5
 * del bloque. Su fecha es más antigua que la de los 8 del catálogo base, así que
 * el bloque de la portada sigue mostrando exactamente los mismos 5.
 */
function relleno(nn: number): MedioFixture {
  const publishedAt = new Date(Date.UTC(2026, 8, 27 - (nn - 1), 15, 0, 0)).toISOString();
  // El `videoId` de YouTube debe tener **11** caracteres (CHECK de la BD).
  const videoId = `e2erellen${String(nn).padStart(2, '0')}`;
  return {
    key: `relleno-${String(nn).padStart(2, '0')}`,
    id: rellenoId(nn),
    title: `Episodio de relleno ${String(nn).padStart(2, '0')}`,
    publishedAt,
    descriptionEs: `Episodio de relleno ${String(nn).padStart(2, '0')} para las tandas de «Cargar más».`,
    descriptionEn: null,
    origin: 'video',
    visibility: 'visible',
    hasCover: true,
    sources: [
      {
        provider: 'youtube',
        externalId: videoId,
        feedTitle: `Episodio de relleno ${String(nn).padStart(2, '0')}`,
        url: youtubeUrl(videoId),
      },
    ],
  };
}

/**
 * Las fichas **no publicables** del catálogo (ocultas, en par abierto o
 * absorbidas): son el objeto del panel y la prueba de que nada de eso sale al
 * público (SC-002).
 */
function noPublicables(): MedioFixture[] {
  return [
    {
      // SQL de `quickstart.md` §5: par «posible duplicado» abierto (base visible,
      // adicional oculta, `duplicate_of_id` recíproco).
      key: 'par-union-base',
      id: IDS.parUnionBase,
      title: 'Sermón de prueba (video)',
      publishedAt: '2026-10-06T15:00:00Z',
      descriptionEs: 'Ficha base del par dudoso.',
      descriptionEn: null,
      origin: 'video',
      visibility: 'visible',
      hasCover: false,
      duplicateOfKey: 'par-union-otra',
      sources: [
        {
          provider: 'youtube',
          externalId: 'aaaaaaaaaaa',
          feedTitle: 'Sermón de prueba (video)',
          url: youtubeUrl('aaaaaaaaaaa'),
        },
      ],
    },
    {
      key: 'par-union-otra',
      id: IDS.parUnionOtra,
      title: 'Otro título para el mismo día',
      publishedAt: '2026-10-06T18:00:00Z',
      descriptionEs: 'Ficha adicional del par dudoso.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'hidden',
      hasCover: false,
      duplicateOfKey: 'par-union-base',
      sources: [
        {
          provider: 'podcast',
          externalId: 'guid-de-prueba-1',
          feedTitle: 'Otro título para el mismo día',
          url: podcastUrl('otro-titulo-para-el-mismo-dia-e3zzzzz'),
        },
      ],
    },
    {
      key: 'par-distintos-base',
      id: IDS.parDistintosBase,
      title: 'Sermón distinto (base)',
      publishedAt: '2026-09-20T15:00:00Z',
      descriptionEs: 'Par marcado que el equipo resolverá como «no son el mismo sermón».',
      descriptionEn: null,
      origin: 'video',
      visibility: 'visible',
      hasCover: false,
      duplicateOfKey: 'par-distintos-otra',
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2edistint1',
          feedTitle: 'Sermón distinto (base)',
          url: youtubeUrl('e2edistint1'),
        },
      ],
    },
    {
      key: 'par-distintos-otra',
      id: IDS.parDistintosOtra,
      title: 'Sermón distinto (adicional)',
      publishedAt: '2026-09-20T18:00:00Z',
      descriptionEs: 'La adicional del par que se resolverá como distinto.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'hidden',
      hasCover: false,
      duplicateOfKey: 'par-distintos-base',
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-distintos',
          feedTitle: 'Sermón distinto (adicional)',
          url: podcastUrl('sermon-distinto-adicional-e2e006'),
        },
      ],
    },
    {
      // Los DOS ejes activos en la misma ficha (P9-19): par abierto **y** marca
      // de disponibilidad abierta; cada uno se resuelve por su cuenta.
      key: 'par-mixto-base',
      id: IDS.parMixtoBase,
      title: 'Sermón con las dos marcas',
      publishedAt: '2026-09-19T15:00:00Z',
      descriptionEs: 'Ficha con par abierto y marca de disponibilidad abierta a la vez.',
      descriptionEn: null,
      origin: 'video',
      visibility: 'visible',
      hasCover: false,
      duplicateOfKey: 'par-mixto-otra',
      availabilityMark: 'missing_from_feed',
      availabilityResult: 'available',
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2emixto001',
          feedTitle: 'Sermón con las dos marcas',
          url: youtubeUrl('e2emixto001'),
        },
      ],
    },
    {
      key: 'par-mixto-otra',
      id: IDS.parMixtoOtra,
      title: 'Sermón con las dos marcas (adicional)',
      publishedAt: '2026-09-19T18:00:00Z',
      descriptionEs: 'La adicional del par mixto.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'hidden',
      hasCover: false,
      duplicateOfKey: 'par-mixto-base',
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-mixto',
          feedTitle: 'Sermón con las dos marcas (adicional)',
          url: podcastUrl('sermon-con-las-dos-marcas-adicional-e2e007'),
        },
      ],
    },
    {
      key: 'oculta-equipo',
      id: IDS.ocultaEquipo,
      title: 'Episodio oculto por el equipo',
      publishedAt: '2026-09-18T15:00:00Z',
      descriptionEs: 'Oculto a mano por el equipo, sin ninguna marca.',
      descriptionEn: null,
      origin: 'video',
      visibility: 'hidden',
      hasCover: false,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2eoculta01',
          feedTitle: 'Episodio oculto por el equipo',
          url: youtubeUrl('e2eoculta01'),
        },
      ],
    },
    {
      // SQL de `quickstart.md` §8(b): el **estado resultante de la precedencia
      // A3** (fase 5): se comprobó que su enlace ya no está disponible → oculta y
      // marcada «no disponible» **abierta**, y el equipo puede «Mostrar de nuevo».
      key: 'marca-unavailable',
      id: IDS.marcaUnavailable,
      title: 'Borrado de verdad',
      publishedAt: '2026-09-17T15:00:00Z',
      descriptionEs: 'Su enlace ya no existe.',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'hidden',
      hasCover: false,
      availabilityMark: 'unavailable',
      availabilityResult: 'unavailable',
      sources: [
        {
          provider: 'podcast',
          externalId: 'guid-muerto-1',
          feedTitle: 'Borrado de verdad',
          url: 'https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/episodio-que-no-existe-abcdef123',
        },
      ],
    },
    {
      key: 'marca-resuelta',
      id: IDS.marcaResuelta,
      title: 'Episodio con la marca ya resuelta',
      publishedAt: '2026-09-16T15:00:00Z',
      descriptionEs: 'Su marca de disponibilidad ya se resolvió: no se reabre por la misma causa.',
      descriptionEn: null,
      origin: 'video',
      visibility: 'hidden',
      hasCover: false,
      availabilityMark: 'missing_from_feed',
      availabilityResult: 'available',
      availabilityResolved: true,
      sources: [
        {
          provider: 'youtube',
          externalId: 'e2eresuelt1',
          feedTitle: 'Episodio con la marca ya resuelta',
          url: youtubeUrl('e2eresuelt1'),
        },
      ],
    },
    {
      key: 'absorbida',
      id: IDS.absorbida,
      title: 'Episodio absorbido por una unión',
      publishedAt: '2026-09-15T15:00:00Z',
      descriptionEs: 'Absorbido por una unión: no aparece en ninguna lista (FR-019).',
      descriptionEn: null,
      origin: 'audio',
      visibility: 'hidden',
      hasCover: false,
      merged: true,
      sources: [
        {
          provider: 'podcast',
          externalId: 'e2e-guid-absorbida',
          feedTitle: 'Episodio absorbido por una unión',
          url: podcastUrl('episodio-absorbido-por-una-union-e2e008'),
        },
      ],
    },
  ];
}

/** Ids de **todas** las fichas que la siembra puede llegar a insertar. */
function catalogoIds(): string[] {
  const base = [...publicablesBase(), ...noPublicables()].map((fixture) => fixture.id);
  const rellenoIds = Array.from({ length: MAX_RELLENO }, (_, index) => rellenoId(index + 1));
  return [...base, ...rellenoIds];
}

/* ── Carátula (subida real, servida por el sitio) ──────────────────────── */

/** Imagen real del repositorio, usada para la carátula (patrón F7/F10). */
const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

/** Lee la cookie CSRF de la sesión abierta (misma que `src/api/client.ts`). */
async function csrfToken(page: Page): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) {
    throw new Error('No hay cookie csrf_token: ¿se ha iniciado sesión en este contexto?');
  }
  return csrf.value;
}

/**
 * Sube una vez la imagen del repositorio con el endpoint de imágenes de F3 y
 * devuelve el nombre de archivo (que las fichas publicables usan como
 * `cover_file`). Sin sesión de panel no se puede subir: los specs siembran con
 * un contexto de administrador.
 */
async function subirCaratula(page: Page): Promise<string> {
  const upload = await page.request.post(`${API_URL}/api/v1/admin/portada/imagenes`, {
    headers: { 'X-CSRF-Token': await csrfToken(page) },
    multipart: {
      file: {
        name: 'caratula-e2e.jpg',
        mimeType: 'image/jpeg',
        buffer: readFileSync(FIXTURE_IMAGE),
      },
    },
  });
  expect(upload.status(), 'la subida de la carátula de los fixtures').toBe(201);
  return ((await upload.json()) as { fileName: string }).fileName;
}

/* ── Determinismo: ocultar/restaurar lo que no es nuestro ─────────────── */

/** Archivo temporal con los ids que la siembra ocultó (auto-reparación). */
const SNAPSHOT_FILE = join(tmpdir(), 'simiente-e2e-f9-medios-ocultos.json');

let ocultadasEnMemoria: string[] | null = null;

function guardarOcultadas(ids: string[]): void {
  ocultadasEnMemoria = ids;
  if (ids.length > 0) {
    writeFileSync(SNAPSHOT_FILE, JSON.stringify(ids), { mode: 0o600 });
  }
}

function leerOcultadas(): string[] {
  if (ocultadasEnMemoria) {
    return ocultadasEnMemoria;
  }
  if (!existsSync(SNAPSHOT_FILE)) {
    return [];
  }
  try {
    return JSON.parse(readFileSync(SNAPSHOT_FILE, 'utf8')) as string[];
  } catch {
    return [];
  }
}

/** Devuelve a `visible` las fichas que la siembra ocultó, y olvida el registro. */
function restaurarOcultadas(): void {
  const ids = leerOcultadas();
  if (ids.length > 0) {
    psql(
      `UPDATE media_items SET visibility = 'visible' WHERE id = ANY(${sqlUuidArray(ids)}::uuid[]);`,
    );
  }
  ocultadasEnMemoria = null;
  if (existsSync(SNAPSHOT_FILE)) {
    rmSync(SNAPSHOT_FILE, { force: true });
  }
}

/**
 * Oculta (temporalmente) toda ficha **visible que no sea nuestra**, para que los
 * totales de la sección y del bloque no dependan del contenido real que la base
 * de desarrollo tenga hoy. Devuelve los ids ocultados para poder restaurarlos.
 */
function ocultarLoAjeno(): void {
  const rows = psql(
    `SELECT id FROM media_items WHERE visibility = 'visible' AND NOT (id = ANY(${sqlUuidArray(
      catalogoIds(),
    )}::uuid[]));`,
  );
  const ids = rows
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line !== '');
  guardarOcultadas(ids);
  if (ids.length > 0) {
    psql(
      `UPDATE media_items SET visibility = 'hidden' WHERE id = ANY(${sqlUuidArray(ids)}::uuid[]);`,
    );
  }
}

/* ── Reset ─────────────────────────────────────────────────────────────── */

/**
 * Deja la base como estaba antes de la siembra: borra **todo** lo nuestro (las
 * fuentes primero, por la FK `ON DELETE RESTRICT`) y **restaura** a `visible` las
 * fichas ajenas que la siembra hubiera ocultado. Es la base de la repetibilidad
 * de los cuatro specs de F9 y se llama también al principio de `seedMedios`.
 *
 * La auditoría **no** se toca: es de solo lectura y acumulativa (FR-022), y los
 * registros que dejan las operaciones del panel se quedan, como en F7.
 */
export function resetMedios(): void {
  const ids = sqlUuidArray(catalogoIds());
  psql(`DELETE FROM media_episode_sources WHERE item_id = ANY(${ids}::uuid[]);`);
  psql(
    `UPDATE media_items SET duplicate_of_id = NULL, merged_into_id = NULL WHERE id = ANY(${ids}::uuid[]);`,
  );
  psql(`DELETE FROM media_items WHERE id = ANY(${ids}::uuid[]);`);
  restaurarOcultadas();
}

/* ── Siembra ───────────────────────────────────────────────────────────── */

/** Sentencia `INSERT` de las fichas indicadas (con el `cover_file` indicado). */
function insertItems(fixtures: MedioFixture[], coverFile: string | undefined): string {
  const values = fixtures.map((fixture) => {
    const cover = fixture.hasCover && coverFile ? sqlText(coverFile) : 'NULL';
    return [
      `${sqlText(fixture.id)}::uuid`,
      sqlText(fixture.title),
      sqlTimestamp(fixture.publishedAt),
      sqlNullable(fixture.descriptionEs),
      sqlNullable(fixture.descriptionEn),
      sqlText(fixture.origin),
      sqlText(fixture.visibility),
      cover,
      fixture.duplicateOfKey ? `${sqlText(porClaveId(fixture.duplicateOfKey))}::uuid` : 'NULL',
      fixture.availabilityMark ? sqlText(fixture.availabilityMark) : 'NULL',
      fixture.availabilityResult ? sqlText(fixture.availabilityResult) : 'NULL',
      fixture.availabilityMark ? sqlTimestamp(fixture.publishedAt) : 'NULL',
      fixture.availabilityResolved ? sqlTimestamp(fixture.publishedAt) : 'NULL',
      fixture.merged ? `${sqlText(IDS.parUnionBase)}::uuid` : 'NULL',
      // `created_at` fijo (y no `now()`): el orden de desempate y la fecha de
      // importación que muestra el panel quedan deterministas entre corridas.
      sqlTimestamp(fixture.publishedAt),
    ].join(', ');
  });

  return `INSERT INTO media_items (
    id, title, published_at, description_es, description_en, origin, visibility,
    cover_file, duplicate_of_id, availability_mark, availability_result,
    availability_checked_at, availability_mark_resolved_at, merged_into_id, created_at
  ) VALUES
    (${values.join('),\n    (')});`;
}

/** Id de una ficha del catálogo por su clave (falla si no existe). */
function porClaveId(key: string): string {
  const found = [...publicablesBase(), ...noPublicables()].find((fixture) => fixture.key === key);
  if (!found) {
    throw new Error(`El catálogo de medios no tiene la clave «${key}»`);
  }
  return found.id;
}

/** Sentencia `INSERT` de las fuentes de las fichas indicadas. */
function insertSources(fixtures: MedioFixture[]): string {
  const rows = fixtures.flatMap((fixture) =>
    fixture.sources.map((source) =>
      [
        `${sqlText(fixture.id)}::uuid`,
        sqlText(source.provider),
        sqlText(source.externalId),
        sqlText(source.feedTitle),
        sqlText(source.url),
      ].join(', '),
    ),
  );

  return `INSERT INTO media_episode_sources (item_id, provider, external_id, feed_title, url) VALUES
    (${rows.join('),\n    (')});`;
}

/**
 * Siembra determinista de F9: limpia lo anterior, **oculta temporalmente** lo
 * ajeno, sube la carátula y crea el catálogo (fichas y fuentes) por SQL.
 * Devuelve las fixtures con sus claves, en el orden de la sección.
 */
export async function seedMedios(page: Page, options: SeedMediosOptions = {}): Promise<MediosSeed> {
  const { publishedCount = 8 } = options;
  if (publishedCount < 0 || publishedCount > 8 + MAX_RELLENO) {
    throw new Error(
      `publishedCount debe estar entre 0 y ${8 + MAX_RELLENO} (el catálogo tiene 8 publicables y ${MAX_RELLENO} de relleno)`,
    );
  }

  resetMedios();
  ocultarLoAjeno();

  const coverFile = await subirCaratula(page);

  const publicables: MedioFixture[] = [...publicablesBase()];
  for (let nn = 1; nn <= publishedCount - 8; nn += 1) {
    publicables.push(relleno(nn));
  }
  const publicablesSembrados = publicables.slice(0, publishedCount);
  const ocultas = noPublicables();
  const todos = [...publicablesSembrados, ...ocultas];

  psql(insertItems(todos, coverFile));
  psql(insertSources(todos));

  const porClave: Record<string, MedioFixture> = {};
  for (const fixture of todos) {
    porClave[fixture.key] = { ...fixture, hasCover: fixture.hasCover && Boolean(coverFile) };
  }

  return {
    suffix: uniqueSuffix(),
    coverFile,
    publicables: publicablesSembrados.map((fixture) => porClave[fixture.key]),
    portada: publicablesSembrados.slice(0, 5).map((fixture) => porClave[fixture.key]),
    ocultas: ocultas.map((fixture) => porClave[fixture.key]),
    porClave,
  };
}

/** Localiza una fixture del catálogo por su clave (falla si no se sembró). */
export function fixtureByKey(seed: MediosSeed, key: string): MedioFixture {
  const found = seed.porClave[key];
  if (!found) {
    throw new Error(`No se sembró la ficha de medios «${key}»`);
  }
  return found;
}

/* ── Cuentas del panel (SC-009, T941) ─────────────────────────────────── */

/** Crea un rol y una cuenta con esos permisos; la contraseña es inicial. */
async function createPanelUser(
  page: Page,
  label: string,
  permissions: string[],
  suffix: string,
): Promise<PanelUser> {
  const roleResponse = await page.request.post(`${API_URL}/api/v1/admin/roles`, {
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': await csrfToken(page) },
    data: { name: `${label} ${suffix}`, permissions },
  });
  expect(roleResponse.ok(), `POST rol «${label}» → ${roleResponse.status()}`).toBe(true);
  const role = (await roleResponse.json()) as { id: string };

  const email = `${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}-${suffix}@ejemplo.com`;
  const password = 'Inicial.2026';
  const userResponse = await page.request.post(`${API_URL}/api/v1/admin/usuarios`, {
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': await csrfToken(page) },
    data: {
      firstName: 'Cuenta',
      lastName: label,
      email,
      phone: '612 000 555',
      roleId: role.id,
      password,
    },
  });
  expect(userResponse.ok(), `POST usuario «${label}» → ${userResponse.status()}`).toBe(true);

  return { email, password, roleId: role.id };
}

/**
 * Crea las tres cuentas del panel que necesitan las aserciones de permisos de
 * T941 (SC-009): con `medios`, con **otros** permisos de módulo y **sin**
 * permisos de módulo. Todas nacen con la contraseña inicial `Inicial.2026`, que
 * el primer acceso obliga a cambiar.
 *
 * Nota: el contrato exige **al menos un permiso** por rol (`RoleCreateInput`),
 * así que la cuenta «sin permisos» lleva `noticias`, que no abre ningún módulo
 * del panel.
 */
export async function seedMediosPanelUsers(page: Page, suffix: string): Promise<MediosPanelUsers> {
  return {
    medios: await createPanelUser(page, 'Solo medios', ['medios'], suffix),
    otherModules: await createPanelUser(page, 'Otros modulos', ['eventos', 'noticias'], suffix),
    none: await createPanelUser(page, 'Sin modulos', ['noticias'], suffix),
  };
}
