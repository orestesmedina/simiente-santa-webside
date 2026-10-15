import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { API_BASE_URL } from '../api/client';

/**
 * Handlers por defecto de la API simulada.
 *
 * La consulta pública de agenda (`GET /api/v1/proximamente`) la dispara el
 * *chrome* público en **todas** las páginas desde F4 (`useSiteLinks`, ⟲ D-23:
 * una sola petición alimenta página, barra, menú móvil y mapa del sitio). Sin
 * un handler por defecto, cualquier prueba que monte `PublicLayout` fallaría
 * por petición no gestionada (`onUnhandledRequest: 'error'`). Devuelve el
 * documento **vacío** (sin bloques) para no alterar ninguna aserción existente;
 * cada prueba lo sobrescribe con `server.use(...)` cuando necesita contenido.
 */
export const server = setupServer(
  http.get(`${API_BASE_URL}/api/v1/proximamente`, () => HttpResponse.json({ lang: 'es' })),
  // El bloque «Lo que viene» de la portada (F4/US7) consulta su propio
  // extracto en cuanto se monta la portada: sin handler por defecto, cualquier
  // prueba que renderice `HomePage` fallaría por petición no gestionada. Sin
  // ítems, como una agenda vacía; cada prueba lo sobrescribe si necesita datos.
  http.get(`${API_BASE_URL}/api/v1/portada/proximamente`, () => HttpResponse.json({ lang: 'es' })),
  // El bloque «Ministerios» de la portada (F6/T629) consulta su extracto al
  // montarse la portada **y** el chrome público lo consulta para decidir la
  // presencia de la entrada de navegación (T628, compartida por `queryClient`):
  // sin handler por defecto, cualquier prueba que monte `PublicLayout` fallaría
  // por petición no gestionada. Sin ítems —«no hay nada publicado»—, de modo
  // que no añade enlaces ni secciones a las aserciones existentes; cada prueba
  // lo sobrescribe cuando necesita contenido.
  http.get(`${API_BASE_URL}/api/v1/portada/ministerios`, () => HttpResponse.json({ lang: 'es' })),
  // F7: la presencia de la entrada «Ofrendas y diezmos» se deduce del
  // **documento de la sección** (`usePublicOfrendas`, P7-8), compartido por
  // `queryClient` con la página `/ofrendas`; y el bloque de la portada consulta
  // su propio extracto al montarse `HomePage` (`usePortadaOfrendas`, T728). Sin
  // handlers por defecto, cualquier prueba que monte `PublicLayout` o la portada
  // fallaría por petición no gestionada. Sin contenido publicado (documento solo
  // `lang`) —«no hay nada»—, de modo que no añade enlaces ni bloques; cada
  // prueba lo sobrescribe cuando necesita datos.
  http.get(`${API_BASE_URL}/api/v1/ofrendas`, () => HttpResponse.json({ lang: 'es' })),
  http.get(`${API_BASE_URL}/api/v1/portada/ofrendas`, () => HttpResponse.json({ lang: 'es' })),
);
