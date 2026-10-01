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
  "orquestador": "opencode-go/deepseek-v4.1-flash",
  "ligero":      "opencode-go/mimo-v2.6-flash",
  "niveles": { "alto": "opencode-go/mimo-v2.6-pro", "medio": "opencode-go/glm-5.3-flash", "bajo": "opencode-go/mimo-v2.6-flash" },
  "agentes": { "revisor-codigo": "opencode-go/mimo-v2.6-pro" }
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
| `dev-backend`, `dev-frontend` | **Muy alto** | Código, uso de herramientas | Un modelo especializado en código y de bajo costo por token. Aquí se va la mayor parte del presupuesto. |
| `qa-tester` | Alto | Código de pruebas | Modelo medio. |
| `revisor-codigo`, `seguridad` | Medio | Detectar errores | Un modelo capaz **de una familia distinta a la de los desarrolladores**, para que no comparta sus puntos ciegos. |
| `disenador-ux`, `devops` | Medio | Criterio práctico | Modelo medio. |
| `documentador` | Bajo | Redacción | El más barato y rápido. |
| Orquestador | Medio | Seguir el proceso, delegar | Un modelo confiable siguiendo instrucciones largas. |

**Regla clave: quien revisa no debería usar el mismo modelo que quien escribió.** Así como quien escribe no aprueba, un modelo tiende a no ver sus propios errores.

## OpenCode Go: cómo funciona el presupuesto

Verificado con la consola de OpenCode y la [documentación](https://opencode.ai/docs/es/go/) (septiembre 2026):

- Hay **un solo presupuesto** que la consola muestra en porcentaje, con tres ventanas: **Rolling** (5 horas, 20% del mensual), **Weekly** (50%) y **Monthly** (100%).
- Cada modelo tiene un **precio por token** y un **límite mensual** propio ($15, $30 o $60). Ambos determinan cuánto del presupuesto consume: un modelo caro y con límite bajo (ej. GLM-5.3, Kimi K3) lo agota muchísimo más rápido que uno barato con límite alto.
- **Repartir el trabajo entre muchos modelos no da más capacidad.** Lo que la estira es usar modelos con buen **rendimiento por token** donde hay más volumen (orquestador, desarrolladores, QA).
- **"Extra Usage"** en la consola: si tienes crédito, al agotar el presupuesto se cobra de ese crédito en lugar de bloquearse.

### Rendimiento por modelo

Tokens aproximados que rinde cada modelo en una ventana de 5 horas, con una carga típica de agente de código (60% lectura de caché, 32% entrada, 8% salida), y señales de calidad de programación de fuentes independientes (septiembre 2026):

| Modelo | Tokens / 5 h | Calidad de código | Uso recomendado |
|---|---|---|---|
| `deepseek-v4.1-flash` | ~122 M fuera de pico · ~61 M en pico | La mejor del plan en uso real (KiloBench 75%), DeepSWE 74 | Orquestador, desarrolladores |
| `mimo-v2.6-pro` | ~14 M | DeepSWE 72, Terminal-Bench 90 | Analista, arquitecto, revisor |
| `glm-5.3-flash` | ~113 M | DeepSWE 63, Terminal-Bench 84 | QA, UX, DevOps |
| `mimo-v2.6-flash` | ~174 M | DeepSWE 68; en pruebas prácticas falla más en ejecución real | Documentador, tareas ligeras |
| `glm-5.3` | ~3 M | DeepSWE 67, Terminal-Bench 88 | Solo roles de muy poco volumen (seguridad) |
| `kimi-k3` | ~1.3 M | La más alta en pruebas reales (KiloBench 73%) | Solo tareas cortas y críticas |
| `kimi-k2.7-code` | ~16 M | Resultados inconsistentes entre benchmarks | No recomendado por costo |
| `minimax-m3` | ~53 M | SWE-bench alto (dato del fabricante), bajo en uso real (KiloBench 48%) | No recomendado |

Los benchmarks de modelos recientes son escasos y a veces contradictorios: úsalos como orientación y valida con tu propio proyecto.

### Horario pico de DeepSeek

Los modelos DeepSeek cuestan el **doble en hora pico**: 01:00–04:00 y 06:00–10:00 UTC, de lunes a viernes. En Costa Rica (UTC−6) eso es **7:00–10:00 p.m. y 12:00–4:00 a.m.** (de domingo a jueves en la noche); el horario laboral diurno es tarifa normal. Es el mismo ID de modelo; solo cambia el precio.

**Recomendación:** hacer el trabajo pesado (implementación y validación) de día, y dejar para la noche lo liviano (revisar specs y planes, aprobar, escribir ideas y roadmaps). No hace falta cambiar de modelo por horario: incluso en pico, `deepseek-v4.1-flash` rinde más que las alternativas de calidad similar.

### Distribución que trae el kit

| Rol | Modelo | Motivo |
|---|---|---|
| Orquestador | `deepseek-v4.1-flash` | Mejor en uso real y mucha capacidad para el rol de más volumen |
| `dev-backend`, `dev-frontend` | `deepseek-v4.1-flash` | Mejor calidad por costo para programar |
| `qa-tester`, `devops`, `disenador-ux` (medio) | `glm-5.3-flash` | Barato y fiable para trabajo repetitivo |
| `revisor-codigo` | `mimo-v2.6-pro` | Alta calidad y otra familia que los desarrolladores |
| `seguridad` | `glm-5.3` | Poco volumen; tercera familia para una revisión independiente |
| `analista-producto`, `arquitecto` (alto) | `mimo-v2.6-pro` | La mejor calidad con capacidad razonable |
| `documentador` (bajo), tareas ligeras | `mimo-v2.6-flash` | El más barato |

Si el presupuesto sigue agotándose rápido, revisa primero el orquestador: abre una sesión nueva por funcionalidad y verifica que esté delegando en los subagentes.

### Cómo medir

1. Anota el % de **Rolling** y **Weekly** antes y después de construir una funcionalidad.
2. Si quieres comparar modelos, construye la misma funcionalidad pequeña con dos configuraciones y compara consumo, ciclos de corrección y hallazgos del revisor.
3. Ajusta `equipo/config.json` y ejecuta `make sincronizar && make modelos`.

### Privacidad (importante con código de clientes)

Revisa la política de datos de cada modelo antes de usarlo con código de clientes. Según la documentación de OpenCode Go, la mayoría no retiene datos, pero hay excepciones: algunos modelos usan los datos para entrenamiento y otros guardan registros durante un tiempo. **No asignes esos modelos a proyectos de clientes** sin su autorización. La lista actualizada está en [opencode.ai/docs/go](https://opencode.ai/docs/go/).

## El orquestador en OpenCode

En OpenCode, el kit crea un agente principal llamado **`orquestador`** y lo deja como **agente por defecto** al abrir OpenCode (`default_agent`). Los integrados `build` y `plan` siguen disponibles con la tecla `Tab`. Sus instrucciones están en `equipo/orquestador.md` (su manual: Spec Kit, el kit, los flujos y las reglas) más `AGENTS.md`, y usa el modelo de `"orquestador"`. En Claude Code, el mismo manual se carga desde `CLAUDE.md`; en Codex, `AGENTS.md` indica leerlo.

```json
"opencode": {
  "orquestador": "opencode-go/deepseek-v4.1-flash",
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
