import { useSystemStatus } from '../hooks/useSystemStatus';
import { StatusResult } from '../components/StatusResult';

/**
 * Página inicial ("Estado del sistema"): consulta el estado al montar y ofrece
 * el único botón para volver a consultarlo. No hace `fetch` directo: usa el
 * hook `useSystemStatus`.
 */
export function StatusPage() {
  const { estado, fechaConsulta, consultando, refetch } = useSystemStatus();

  return (
    <article className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Estado del sistema</h1>
        <p className="mt-2 text-slate-700">
          Este sitio web está en construcción. Por ahora, esta página muestra si el sistema de la
          Iglesia Simiente Santa está funcionando correctamente.
        </p>
      </div>

      <section>
        <h2 className="text-lg font-semibold">Estado actual</h2>
        <div role="status" aria-live="polite" className="mt-2 rounded border border-slate-200 p-4">
          {estado === undefined ? (
            <p>Consultando el estado del sistema…</p>
          ) : (
            <StatusResult
              estado={estado}
              fechaConsulta={fechaConsulta}
              actualizando={consultando}
            />
          )}
        </div>
        <button
          type="button"
          onClick={refetch}
          disabled={consultando}
          className="mt-4 min-h-11 rounded bg-slate-900 px-4 py-2 font-medium text-white hover:bg-slate-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
        >
          Volver a consultar el estado
        </button>
      </section>
    </article>
  );
}
