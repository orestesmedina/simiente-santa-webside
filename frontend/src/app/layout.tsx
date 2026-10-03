import { Link, Outlet } from 'react-router-dom';

/**
 * Layout mínimo y semántico de la aplicación: encabezado con navegación
 * principal y región `main` donde renderiza la ruta activa. El foco visible
 * se deja al estilo por defecto del navegador (accesibilidad de la skill).
 */
export function AppLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-white text-slate-900">
      <header className="border-b border-slate-200">
        <nav aria-label="Navegación principal" className="mx-auto w-full max-w-3xl px-4 py-3">
          <Link
            to="/"
            className="rounded text-sm font-medium text-slate-700 underline-offset-4 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
          >
            Inicio
          </Link>
        </nav>
      </header>
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
