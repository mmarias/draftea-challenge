# draftea-challenge

## Guía
1. En **/docs** se encuentra toda la documentación de la arquitectura definida para el challenge
2. Para simplificar el reto, los distintos compoenentes y servicios estarán centralizados en el mismo repositorio; En el caso real, se debe respectar la separación definida previamente
3. Pruebas cubiertas basadas en lo solicitado en el challenge **Escenarios de Flujo de Pago**
  - Ruta Feliz: Pago exitoso de principio a fin.
  - Saldo Insuficiente: Flujo de rechazo inmediato.
  - Tiempo de Espera de Pasarela Externa: Reintento y fallback.
  - Pagos Concurrentes: Manejo de condiciones de carrera.
  - Recuperación del Sistema: Reinicio y reconstrucción del estado.
4. Para ejecutar los test definidos, ejecutar el comando ```make test```
