-- 000009_seed_permission_categorias (up)
-- F4 (cambio de alcance aprobado el 2026-10-11): el catálogo de permisos de
-- F2 (000002) crece con un décimo permiso, `categorias`, que autoriza la
-- gestión del catálogo de categorías de agenda (/api/v1/admin/categorias).
-- Los permisos `eventos` y `actividades` NO autorizan este módulo (son
-- independientes; research R4-4).
--
-- Es solo un INSERT: 000002 está aplicada y NO se edita (§VI). El
-- administrador inicial lo recibe automáticamente (el init de F2 lee el
-- catálogo dinámicamente y crea el rol «Administrador» con todos los permisos).
INSERT INTO permissions (code, label) VALUES
    ('categorias', 'Categorías de agenda');
