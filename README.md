# draftea-challenge

## Guía
1. En **/docs** se encuentra toda la documentación de la arquitectura definida para el challenge.
2. Para simplificar el reto, los distintos componentes y servicios estarán centralizados en el mismo repositorio; En el caso real, se debe respectar la separación definida previamente.
3. Se agregaron servicio simulados de event bus pub/sub, cache local y simulación de operaciones en payments repository.

## Requisitos
1. go 1.24+
2. Los test cases cubren:
    - Producción de eventos
    - Casos de retry y backoff
    - Orchestrator SAGA 
    - Casos de concurrencia al crear pagos sobre ```POST /payments```
3. Ejecutar los test ```go test ./...```

