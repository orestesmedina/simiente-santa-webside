import type { EstadoSistema } from '../estado';
import { StatusItem } from './StatusItem';

interface StatusResultProps {
  estado: EstadoSistema;
  fechaConsulta: Date | undefined;
  actualizando: boolean;
}

function formatHora(fecha: Date): string {
  return fecha.toLocaleTimeString('es-ES', { hour12: false });
}

/**
 * Renderiza el estado actual: veredicto, detalle de los dos componentes, hora
 * de la última consulta y explicación si es un error. Los literales son los de
 * ux.md §6. El orden es constante en los cuatro estados.
 */
export function StatusResult({ estado, fechaConsulta, actualizando }: StatusResultProps) {
  if (estado.kind === 'conectado') {
    return (
      <>
        {actualizando && <p className="mb-2 text-slate-600">Actualizando el estado…</p>}
        <p className="text-lg font-semibold" data-testid="veredicto">
          El sistema está funcionando.
        </p>
        <dl className="mt-3" data-testid="detalle">
          <StatusItem nombre="Servidor" valor="servidor-ok" />
          <StatusItem nombre="Base de datos" valor="bd-ok" />
        </dl>
        {fechaConsulta && (
          <p className="mt-3 text-sm text-slate-600">
            Última consulta:{' '}
            <time dateTime={fechaConsulta.toISOString()}>{formatHora(fechaConsulta)}</time>.
          </p>
        )}
      </>
    );
  }

  if (estado.kind === 'bd-no-conectada') {
    return (
      <>
        {actualizando && <p className="mb-2 text-slate-600">Actualizando el estado…</p>}
        <p className="text-lg font-semibold" data-testid="veredicto">
          El sistema está en marcha, pero la base de datos no está conectada.
        </p>
        <dl className="mt-3" data-testid="detalle">
          <StatusItem nombre="Servidor" valor="servidor-ok" />
          <StatusItem nombre="Base de datos" valor="bd-error" />
        </dl>
        {fechaConsulta && (
          <p className="mt-3 text-sm text-slate-600">
            Última consulta:{' '}
            <time dateTime={fechaConsulta.toISOString()}>{formatHora(fechaConsulta)}</time>.
          </p>
        )}
        <p className="mt-3">
          El sitio web está corriendo, pero no puede leer ni guardar información por ahora. Esto
          suele ocurrir cuando la base de datos está detenida o todavía está arrancando.
        </p>
      </>
    );
  }

  const sinRespuesta = estado.motivo === 'sin-respuesta';

  return (
    <>
      {actualizando && <p className="mb-2 text-slate-600">Actualizando el estado…</p>}
      <p className="text-lg font-semibold" data-testid="veredicto">
        No se pudo consultar el estado del sistema.
      </p>
      <dl className="mt-3" data-testid="detalle">
        <StatusItem nombre="Servidor" valor={sinRespuesta ? 'sin-respuesta' : 'no-comprobable'} />
        <StatusItem nombre="Base de datos" valor="no-comprobable" />
      </dl>
      {fechaConsulta && (
        <p className="mt-3 text-sm text-slate-600">
          Última consulta:{' '}
          <time dateTime={fechaConsulta.toISOString()}>{formatHora(fechaConsulta)}</time>.
        </p>
      )}
      {sinRespuesta ? (
        <p className="mt-3">
          No hubo respuesta del servidor, así que no sabemos si el sistema está funcionando. Esto no
          significa que la base de datos esté desconectada: aquí el problema está en el propio
          servidor o en la conexión con él. Espere unos momentos y vuelva a consultar.
        </p>
      ) : (
        <p className="mt-3">
          El servidor respondió, pero con algo que no pudimos interpretar, así que no sabemos si el
          sistema está funcionando. Esto no significa que la base de datos esté desconectada: aquí
          el problema está en el propio servidor o en la conexión con él. Espere unos momentos y
          vuelva a consultar.
        </p>
      )}
    </>
  );
}
