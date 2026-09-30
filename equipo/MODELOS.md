# Asignación de modelos por agente

Cada agente puede usar un modelo distinto. Todo se configura en `equipo/config.json`; los roles en `equipo/agentes/` no nombran modelos, solo un **nivel** (`alto`, `medio`, `bajo`).

## Cómo se decide el modelo de un agente

Para cada herramienta, en este orden de prioridad:

1. **`agentes`** — excepción explícita para ese agente.
2. **`niveles`** — el modelo asignado al nivel del agente.
3. **Vacío** — el agente usa el mismo modelo que la sesión principal.

Además:
- **`orquestador`** — modelo de la sesión principal (la que ejecuta los comandos de Spec Kit y coordina). Se escribe en `opencode.json` (`model`) o `.claude/settings.json` (`model`). En Codex se elige al iniciar la sesión.
- **`ligero`** (solo OpenCode) — modelo para tareas auxiliares como generar títulos (`small_model`).

```json
"opencode": {
  "orquestador": "opencode-go/glm-5.3",
  "ligero":      "opencode-go/glm-5.3-flash",
  "niveles": { "alto": "opencode-go/kimi-k3", "medio": "opencode-go/glm-5.3", "bajo": "opencode-go/glm-5.3-flash" },
  "agentes": { "revisor-codigo": "opencode-go/deepseek-v4-pro" }
}
```

Después de cualquier cambio:
```bash
make sincronizar   # regenera la configuración
make modelos       # muestra qué modelo usa cada agente y de dónde sale
```

## Criterios para asignar modelos

| Rol | Volumen de tokens | Qué necesita | Recomendación |
|---|---|---|---|
| `analista-producto`, `arquitecto` | Bajo | Razonamiento, criterio | El modelo más capaz. Se usan poco, así que el costo es bajo y un error aquí es el más caro. |
| `dev-backend`, `dev-frontend` | **Muy alto** | Código, uso de herramientas | Un modelo especializado en código, con buen límite de uso. Aquí se va la mayor parte del consumo. |
| `qa-tester` | Alto | Código de pruebas | Modelo medio. |
| `revisor-codigo`, `seguridad` | Medio | Detectar errores | Un modelo capaz **de una familia distinta a la de los desarrolladores**, para que no comparta sus puntos ciegos. |
| `disenador-ux`, `devops` | Medio | Criterio práctico | Modelo medio. |
| `documentador` | Bajo | Redacción | El más barato y rápido. |
| Orquestador | Medio | Seguir el proceso, delegar | Un modelo confiable siguiendo instrucciones largas. |

**Regla clave: quien revisa no debería usar el mismo modelo que quien escribió.** Así como quien escribe no aprueba, un modelo tiende a no ver sus propios errores.

## Ejemplo con OpenCode Go

El `config.json` del kit trae este ejemplo:

| Agente | Modelo | Motivo |
|---|---|---|
| orquestador | `glm-5.3` | Equilibrio entre costo y fiabilidad |
| analista-producto, arquitecto | `kimi-k3` (nivel alto) | Máxima calidad donde el volumen es bajo |
| dev-backend, dev-frontend | `kimi-k2.7-code` | Especializado en código |
| qa-tester, disenador-ux, devops | `glm-5.3` (nivel medio) | Uso general |
| revisor-codigo | `deepseek-v4-pro` | Familia distinta a los desarrolladores |
| seguridad | `qwen3.8-max` | Otra familia distinta, segunda opinión independiente |
| documentador | `glm-5.3-flash` (nivel bajo) | Barato y rápido |

Este reparto es un **punto de partida**, no una recomendación probada. Midan resultados y ajusten.

### Límites de uso de OpenCode Go

- La suscripción da un límite de gasto por modelo, dividido en ventanas de 5 horas (20%), semanal (50%) y mensual (100%).
- Los modelos caros tienen un límite más bajo. Por eso el modelo más potente va en los roles de poco volumen (analista, arquitecto) y **no** en los desarrolladores.
- Si un modelo agota su límite, cambia ese nivel o agente a otro modelo y ejecuta `make sincronizar`.

### Privacidad (importante con código de clientes)

Revisa la política de datos de cada modelo antes de usarlo con código de clientes. Según la documentación de OpenCode Go, la mayoría no retiene datos, pero hay excepciones: algunos modelos usan los datos para entrenamiento y otros guardan registros durante un tiempo. **No asignes esos modelos a proyectos de clientes** sin su autorización. La lista actualizada está en [opencode.ai/docs/go](https://opencode.ai/docs/go/).

## El orquestador en OpenCode

En OpenCode, el kit crea un agente principal llamado **`orquestador`** y lo deja como **agente por defecto** al abrir OpenCode (`default_agent`). Los integrados `build` y `plan` siguen disponibles con la tecla `Tab`. Sus instrucciones están en `equipo/orquestador.md` (su manual: Spec Kit, el kit, los flujos y las reglas) más `AGENTS.md`, y usa el modelo de `"orquestador"`. En Claude Code, el mismo manual se carga desde `CLAUDE.md`; en Codex, `AGENTS.md` indica leerlo.

```json
"opencode": {
  "orquestador": "opencode-go/glm-5.3",
  "agente_principal": {
    "nombre": "orquestador",
    "ocultar": [],
    "temperatura": 0.2
  }
}
```

- Si falta la sección `agente_principal`, se activa igual con estos valores.
- `"ocultar": ["build", "plan"]` esconde los agentes integrados.
- `"nombre": ""` desactiva el orquestador y vuelve a los agentes integrados de OpenCode.

## Temperatura

La temperatura controla cuánto azar usa el modelo: baja (0–0.2) da respuestas consistentes y repetibles; media (0.3–0.6) da más variedad. Hoy solo **OpenCode** permite fijarla por subagente; en Claude Code y Codex se omite.

### Cómo se decide la temperatura de un agente

En este orden de prioridad:

1. **`sin_temperatura`** — si el modelo del agente está en esta lista, no se envía temperatura y se usa la que exige el modelo.
2. **`temperatura.agentes`** en `config.json` — excepción para ese agente.
3. **`temperatura:` en `equipo/agentes/<rol>.md`** — el valor por defecto del agente.
4. **Nada** — se usa la temperatura por defecto del modelo.

Valores por defecto que trae el kit:

| Agente | Temperatura | Motivo |
|---|---|---|
| `dev-backend`, `dev-frontend`, `devops` | 0.1 | Código consistente y repetible |
| `qa-tester`, `revisor-codigo`, `seguridad` | 0.1 | Un revisor no debe dar veredictos distintos cada vez |
| `arquitecto` | 0.2 | Algo de amplitud para evaluar alternativas |
| `documentador` | 0.3 | Redacción natural sin inventar |
| `analista-producto`, `disenador-ux` | 0.4 | Pensar en más casos límite y opciones |

Ejemplo de excepciones:

```json
"opencode": {
  "temperatura": {
    "agentes": { "revisor-codigo": 0 },
    "sin_temperatura": ["opencode-go/kimi-k3"]
  }
}
```

**Importante:** muchos modelos de razonamiento recomiendan o exigen una temperatura concreta, y otro valor puede empeorar los resultados. Revisa la ficha de cada modelo; si indica un valor fijo, agrégalo a `sin_temperatura`.

`make modelos` muestra la temperatura final de cada agente y de dónde sale (`agente`, `config`, `fija del modelo` o `del modelo`).

## Verificar que los modelos existan

Los catálogos cambian con frecuencia. Para ver los IDs exactos disponibles en tu cuenta:

```bash
opencode models opencode-go
```

`make modelos` marca con ⚠ cualquier modelo configurado que tu instalación de OpenCode no reconozca.

## Evaluar y ajustar

1. Ejecuten la misma funcionalidad de prueba con dos configuraciones distintas.
2. Comparen: ¿cuántos ciclos de corrección necesitó? ¿Qué encontró el revisor? ¿Cuánto consumo?
3. Anoten el resultado en `docs/decisiones/` y ajusten `config.json`.
