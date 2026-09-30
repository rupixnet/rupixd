# Workflows retirados

Nada se borra: aquí quedan los workflows que ya no corren, para memoria.

- `tests-heredado-kaspa.yaml` (retirado el 30-sep-2026): el `Tests` heredado de Kaspa. Estuvo rojo en cada push durante semanas por piezas que no eran nuestras (los "stability tests" de Kaspa, `go get -d`, y una subida a Codecov), mientras `go test ./...` pasaba. Un CI que siempre está rojo no avisa de nada. Se reemplazó por `tests.yaml` con lo que sí es de Rupix: `gofmt`, `go vet`, `staticcheck`, `go build` y `go test ./...` en Linux y macOS.
- `race-heredado-kaspa.yaml` (retirado el 30-sep-2026): corría cada noche sobre una rama `master` y ramas `v*-dev` que en Rupix no existen (la nuestra es `main`), así que fallaba cada madrugada sin probar nada. Reemplazado por `race.yaml` sobre `main`.
