---
name: go-backend
description: Convenciones para escribir servicios en Go según la documentación oficial (Effective Go, Go Code Review Comments, la guía de organización de módulos y las buenas prácticas de seguridad de go.dev). Usar siempre que se cree o modifique un archivo .go.
metadata:
  madurez: redactada
---
# Go — convenciones

Fuentes, todas en go.dev: Effective Go, Go Code Review Comments (go.dev/wiki/CodeReviewComments), Organizing a Go module (go.dev/doc/modules/layout), Accessing a relational database (go.dev/doc/database) y Security Best Practices (go.dev/doc/security/best-practices). Revisado el 2026-10-09. Ante una duda que esta página no resuelva, se consulta ahí.

## Versión
- El equipo de Go da soporte a las dos últimas versiones mayores (al 2026-10-09: 1.26 y 1.27). Un proyecto usa una de ellas, declarada en `go.mod`, y se mantiene en su última versión menor.
- Las dependencias se manejan con módulos: `go.mod` y `go.sum` van en git.

## Estructura
La que go.dev recomienda para un servidor:
```
go.mod
cmd/
  <programa>/main.go     # un directorio por ejecutable; main solo arranca
internal/
  <paquete>/             # toda la lógica; otros módulos no pueden importarla
```
- Un servidor no exporta paquetes: su lógica va en `internal/`.
- Los paquetes se organizan por lo que hacen (`orders`, `auth`), no por tipo de archivo (`models`, `utils`, `helpers`).
- `main` se limita a leer la configuración, construir las dependencias y arrancar. Las dependencias se pasan como parámetros; no hay variables globales con estado.

## Formato y nombres
- Todo el código pasa por `gofmt`; los `import`, por `goimports` (biblioteca estándar primero, en su propio grupo). El formato no se discute ni se hace a mano.
- `MixedCaps` o `mixedCaps`, nunca guiones bajos, tampoco en constantes (`maxLength`, no `MAX_LENGTH`).
- Las siglas mantienen sus mayúsculas: `URL`, `ID`, `HTTP` (`userID`, no `userId`).
- Paquetes: una palabra corta, en minúsculas, sin guiones bajos ni plural. El nombre del paquete ya es parte del nombre: `orders.Service`, no `orders.OrderService`.
- Un método que devuelve un campo no lleva `Get`: `Owner()`, no `GetOwner()`.
- Receptor: una o dos letras del tipo (`o` para `Order`), el mismo en todos sus métodos; nunca `this` ni `self`. Receptor puntero si el método modifica el valor o el tipo es grande, y entonces en todos los métodos del tipo.
- Variables locales cortas; cuanto más lejos se usa un nombre de donde se declara, más descriptivo.
- Todo lo exportado lleva un comentario que es una oración completa y empieza con su nombre.

## Errores
- Todo error devuelto se revisa. Nunca se descarta con `_`.
- El error se maneja primero y se sale; el camino normal queda sin sangría, sin `else`.
- Para agregar contexto se envuelve con `%w` (`fmt.Errorf("create order: %w", err)`) y se examina con `errors.Is` y `errors.As`, no comparando textos.
- Los textos de error van en minúsculas y sin punto final.
- Un fallo se devuelve como `error`, no con un valor especial (`-1`, `""`, `nil`).
- `panic` no es manejo de errores: solo para lo que no debería poder ocurrir.

## Interfaces
- Se declaran en el paquete que las **usa**, no en el que las implementa, y son pequeñas (uno o pocos métodos).
- Las funciones devuelven tipos concretos y aceptan interfaces.
- No se declara una interfaz antes de necesitarla.

## Contexto y concurrencia
- `context.Context` es el primer parámetro (`ctx`) de todo lo que hace entrada y salida o puede tardar. No se guarda dentro de una estructura.
- Al lanzar una goroutine debe estar claro cuándo termina. Ninguna queda viva sin forma de detenerla.
- Las funciones son síncronas por defecto: devuelven su resultado; quien llama decide si la corre en una goroutine.
- Las pruebas se ejecutan con el detector de carreras (`go test -race`).

## Base de datos
- Los valores van siempre como parámetros de la consulta. Nunca se arma el SQL con `fmt.Sprintf` ni uniendo textos.
- Se usan las variantes con contexto (`QueryContext`, `ExecContext`).
- Las filas se cierran (`defer rows.Close()`) y al terminar de recorrerlas se revisa `rows.Err()`.
- Una transacción lleva `defer tx.Rollback()` justo después de abrirla, y `tx.Commit()` al final.
- `sql.DB` es un grupo de conexiones: se crea una vez y se comparte.

## Seguridad
- `govulncheck ./...` sin vulnerabilidades conocidas, y Go en su última versión menor.
- Aleatorios para claves y tokens con `crypto/rand`, nunca `math/rand`.
- HTML generado con `html/template`, que escapa según el contexto; nunca con `text/template` ni uniendo textos.
- Un servidor HTTP lleva tiempos límite (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`): el valor por defecto es sin límite.
- Toda entrada se valida al llegar. A quien llama nunca se le devuelve el error interno.
- Secretos por variables de entorno; nunca en el código ni en los registros.

## Pruebas
- Paquete `testing` de la biblioteca estándar. Las pruebas van junto al código, en archivos `_test.go` del mismo directorio.
- Tabla de casos con subpruebas (`t.Run`).
- Un fallo dice qué se probó, qué se obtuvo y qué se esperaba: `Foo(%q) = %d; want %d`.
- Las dependencias se reemplazan por dobles a través de la interfaz que declara quien las usa. Para HTTP, `net/http/httptest`.
- Las funciones de apoyo llaman a `t.Helper()`.

## Comandos
| Para | Comando |
|---|---|
| Comprobar el formato | `test -z "$(gofmt -l .)"` |
| Análisis estático | `go vet ./...` |
| Pruebas | `go test -race ./...` |
| Cobertura | `go test -cover ./...` |
| Vulnerabilidades | `govulncheck ./...` |

Los que el proyecto tiene de verdad están en su perfil.

## Lo que no va en git
- Los binarios compilados, los ejecutables de prueba (`*.test`) y los perfiles de cobertura (`*.out`).
- `go.mod` y `go.sum` sí van en git, siempre (go.dev/ref/mod). `vendor/` va en git solo si el proyecto decidió versionar sus dependencias.

## Si el proyecto tiene sus propias convenciones
- Esta página es el estándar. Si el proyecto tiene una skill de convenciones propias, **en nombres, estructura, patrones y bibliotecas manda esa**, para que el código nuevo se parezca al que ya hay. Lo que ella no diga se hace como dice esta página.
- **Seguridad** y los parámetros en las consultas no admiten excepción. Si cumplirlos exige cambiar algo que tu tarea no cubre, detente y avisa.
- No reformatees ni reorganices lo que tu tarea no toca. Lo que encuentres fuera del estándar va en la entrega, como deuda.
