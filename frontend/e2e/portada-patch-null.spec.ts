import { expect, test } from '@playwright/test';

import { ADMIN, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * Guardia de regresión e2e del bloqueante B1 (ciclo 2 de F3,
 * `revision-2026-10-10-qa.md`): confirma en el STACK REAL que los PATCH del
 * panel distinguen `null` (limpiar) de campo ausente (no tocar), que la
 * limpieza se audita y que se ve de inmediato en la portada pública. Nació
 * como traza viva de QA y se versiona como prueba permanente al quedar el
 * defecto corregido (`d5d4610`).
 */
test('B1 cerrado: null limpia endTime/nameEn en vivo, audita y no toca lo ausente', async ({
  page,
}) => {
  test.setTimeout(60_000);
  await ensureInitialized(page.request);
  await loginViaUi(page, ADMIN.email, ADMIN.password);
  await page.waitForURL(/\/panel/, { timeout: 20_000 });

  const suf = uniqueSuffix();

  // 1) Crear servicio publicado con hora de fin y traducción al inglés.
  const created = await apiSend(page, 'POST', '/api/v1/admin/portada/horario', {
    nameEs: `QA-B1c2 ${suf}`,
    nameEn: 'QA-B1c2 service',
    placeEs: 'Templo',
    dayOfWeek: 0,
    startTime: '10:00',
    endTime: '12:00',
    publicationState: 'published',
  });
  expect(created.status()).toBe(201);
  const item = (await created.json()) as {
    id: string;
    nameEs: string;
    endTime: string | null;
    nameEn: string | null;
    publicationState: string;
  };
  expect(item.endTime).toBe('12:00');
  expect(item.publicationState).toBe('published');

  try {
    // 2) PATCH con {"endTime": null}: limpiar ES un cambio → 200 (no 400).
    const patched = await apiSend(page, 'PATCH', `/api/v1/admin/portada/horario/${item.id}`, {
      endTime: null,
    });
    expect(patched.status()).toBe(200);
    const saved = (await patched.json()) as {
      endTime: string | null;
      nameEn: string | null;
      publicationState: string;
    };
    // null limpió la hora de fin (la respuesta omite el campo vacío)…
    expect((saved.endTime ?? null) === null).toBe(true);
    // …sin tocar lo ausente: el inglés y el estado publicado persisten.
    expect(saved.nameEn).toBe('QA-B1c2 service');
    expect(saved.publicationState).toBe('published');

    // 3) PATCH con {"nameEn": null} vacía la traducción y no despublica.
    const patched2 = await apiSend(page, 'PATCH', `/api/v1/admin/portada/horario/${item.id}`, {
      nameEn: null,
    });
    expect(patched2.status()).toBe(200);
    const saved2 = (await patched2.json()) as { nameEn: string | null; publicationState: string };
    expect((saved2.nameEn ?? null) === null).toBe(true);
    expect(saved2.publicationState).toBe('published');

    // 4) Estado admin tras refetch: limpieza persistida (no devolvía el valor).
    const state = (await (await apiSend(page, 'GET', '/api/v1/admin/portada')).json()) as {
      schedule: { items: Array<{ id: string; endTime?: string | null }> };
    };
    const persisted = state.schedule.items.find((s) => s.id === item.id);
    expect(persisted).toBeDefined();
    expect((persisted!.endTime ?? null) === null).toBe(true);

    // 5) La portada pública, ya publicada, refleja la limpieza al recargar
    //    (FR-014/SC-003 intacto): sin hora de fin para este servicio.
    const pub = await page.request.get('http://localhost:8080/api/v1/portada?lang=es', {
      headers: { 'Cache-Control': 'no-cache' },
    });
    expect(pub.status()).toBe(200);
    const payload = (await pub.json()) as {
      schedule?: Array<{ id: string; endTime?: string | null }>;
    };
    const pubItem = (payload.schedule ?? []).find((s) => s.id === item.id);
    expect(pubItem).toBeDefined();
    expect((pubItem!.endTime ?? null) === null).toBe(true);

    // 6) La limpieza quedó AUDITADA como update (FR-017). El payload de
    //    `auditoria/acciones` identifica el objeto por `targetLabel` (contiene
    //    el nombre único del servicio) para acciones de contenido.
    const audit = await page.request.get(
      'http://localhost:8080/api/v1/admin/auditoria/acciones?limit=100',
    );
    expect(audit.status()).toBe(200);
    const actions = (await audit.json()) as {
      items?: Array<{ action: string; targetLabel?: string | null }>;
    };
    const rows = actions.items ?? [];
    expect(
      rows.some(
        (r) => r.action === 'home.schedule.update' && r.targetLabel?.includes(item.nameEs!),
      ),
    ).toBe(true);
  } finally {
    // Limpieza: la prueba no deja basura en el panel.
    await apiSend(page, 'DELETE', `/api/v1/admin/portada/horario/${item.id}`);
  }

  // 7) Contratrama de WhatsApp: {"nameEn": null} limpia también en canales.
  const ch = await apiSend(page, 'POST', '/api/v1/admin/portada/whatsapp', {
    nameEs: `QA-B1c2 canal ${suf}`,
    nameEn: 'QA-B1c2 channel',
    kind: 'direct',
    destination: '04121234567',
    publicationState: 'published',
  });
  expect(ch.status()).toBe(201);
  const chan = (await ch.json()) as { id: string; nameEn: string | null };
  try {
    const patchedCh = await apiSend(page, 'PATCH', `/api/v1/admin/portada/whatsapp/${chan.id}`, {
      nameEn: null,
    });
    expect(patchedCh.status()).toBe(200);
    const savedCh = (await patchedCh.json()) as { nameEn: string | null; publicationState: string };
    expect((savedCh.nameEn ?? null) === null).toBe(true);
    expect(savedCh.publicationState).toBe('published');
  } finally {
    await apiSend(page, 'DELETE', `/api/v1/admin/portada/whatsapp/${chan.id}`);
  }
});
