import { expect, test } from '@playwright/test';

/**
 * E2E del flujo crítico de F1 (SC-005): al abrir la página inicial, el estado
 * del sistema se muestra sin ninguna acción adicional.
 *
 * El stack completo debe estar levantado (`make up`): nginx sirve la SPA, que
 * consulta `GET /healthz` del backend, que a su vez hace `Ping` a PostgreSQL.
 * Los escenarios de error A/B (detener `db` / `backend`) se validan
 * **manualmente** con `quickstart.md` §4 — no se automatizan aquí porque
 * exigirían manipular contenedores.
 */
test('la página inicial muestra el estado del sistema sin acciones adicionales', async ({
  page,
}) => {
  await page.goto('/');

  // Sin pulsar nada: la consulta se dispara al montar (ux.md §1).
  // El ping de BD responde en ≤2 s (D8), así que el estado debe estar visible
  // en <3 s (SC-005). `toBeVisible` reintenta hasta el timeout de Playwright;
  // el límite de 3 s es explícito.
  const veredicto = page.getByText('El sistema está funcionando.');
  await expect(veredicto).toBeVisible({ timeout: 3_000 });

  // Detalle de los dos componentes (ux.md §2): servidor en marcha y BD conectada.
  const detalle = page.getByTestId('detalle');
  await expect(detalle).toContainText('Servidor');
  await expect(detalle).toContainText('en marcha');
  await expect(detalle).toContainText('Base de datos');
  await expect(detalle).toContainText('conectada');

  // Región de estado en vivo (ux.md §5).
  await expect(page.getByRole('status')).toBeVisible();

  // El único control de la página sigue disponible (ux.md §2).
  await expect(page.getByRole('button', { name: 'Volver a consultar el estado' })).toBeEnabled();
});
