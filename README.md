# draftea-challenge

## Guía
1. En **/docs** se encuentra toda la documentación de la arquitectura definida para el challenge.
2. Para simplificar el reto, los distintos componentes y servicios estarán centralizados en el mismo repositorio; En el caso real, se debe respectar la separación definida previamente.
3. Se agregaron servicio simulados de event bus pub/sub, cache local y simulación de operaciones en payments repository.

## Pruebas Unitarias
1. go 1.24+
2. Los casos de prueba cubren:
    - Producción de eventos
    - Casos de reintentos y backoff
    - Orchestrator SAGA
    - Casos de concurrencia al crear pagos sobre ```POST /payments```
3. Ejecutar las pruebas ```go test ./...```

## Prueba de Integración (happy path)
Si enciende la aplicación ```go run cmd/api/main.go``` y ejecuta el siguiente curl
```curl --location 'localhost:8080/payments' \
--header 'x-idempotent-key: 1234' \
--header 'Content-Type: application/json' \
--data '{ 
    "wallet_id": "wal-1234",
    "service_id": "abc-1234",
    "amount": 1230,
    "currency": "USD",
    "method": "balance"
}'
```
Se simula el consumo completo p2p del procesamiento de un pago exitoso.

## Consideraciones Futuras de Rendimiento y Escalabilidad

1.  **API Gateway (`cmd/api`):**
    *   **Consideraciones:** Implementar escalado horizontal (múltiples instancias detrás de un balanceador de carga), manejo eficiente de solicitudes no bloqueantes y limitación de tasas para gestionar un alto número de solicitudes concurrentes.
2.  **Consumidores (`cmd/*_consumer`):**
    *   **Consideraciones:** Escalar los consumidores horizontalmente y asegurar que toda la lógica del consumidor sea idempotente para manejar de forma segura los reintentos y evitar efectos secundarios.
3.  **Memcache (`internal/infraestructure/memcache`):**
    *   **Consideraciones:** Para entornos distribuidos, migrar de una caché local en memoria a una solución de caché distribuida (por ejemplo, Redis, clúster de Memcached) para garantizar la consistencia y escalabilidad entre múltiples instancias de servicio.
4. **Eventos**
   *    **Consideraciones:** Se puede implementar event sourcing, con un DynamoDB, para almacenar los eventos fallidos o reconstruir, en forma de auditoría, cada uno de los steps SAGA.

**Rendimiento General y Observabilidad:**

*   **Trazado Distribuido y Monitorización:** Integrar el trazado distribuido OpenTelemetry, Jaeger para una visibilidad de extremo a extremo de los flujos de transacciones entre servicios. Establecer una monitorización y alertas exhaustivas para métricas clave (latencia, tasas de error, profundidades de cola, utilización de recursos).
*   Implementación de prometheus para popular métricas e integrar todo con Grafana.
*   **Manejo Robusto de Errores:** Implementar mecanismos de reintento, circuit breakers y múltiples proveedores en Payments Gateway para mejorar la resiliencia del sistema frente a fallos transitorios y prevenir errores en cascada.
