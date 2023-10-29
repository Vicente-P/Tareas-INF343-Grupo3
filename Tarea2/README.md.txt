# Tarea2 INF343 GRUPO3


## Proyect Members 

* Vicente Perez - 202073042-2
* Russel Guevara - 202073121-7

## Instrucciones Tarea

*La aplicacion esta programada para funcionar de forma local.
*Antes de empezar debe asegurar de que mongodb este inicializado. 
*La base de datos estara vacia, por lo que debera añadir productos simulando el stock para esto realizamos un programa llamado productos.go, que añade los distintos productos con una cantidad igual a 98. 
*Para hacer funcionar la aplicacion debe correr los distintos archivos que se mencionaran en diferentes terminales (debe ingresar al directorio correspondiente y ingresar el comando go run nombre_archivo.go) 
*Primero debe correr ventas.go (archivo que recibe una orden de compra y comunica esta con los distintos servicios).
*Despues debe correr los correspondientes servicios: inventario.go (se encarga de actualizar el stock), despacho.go (se encarga de modificar la orden y añadir el despacho), notificacion.go (envia un correo electronico al cliente con la informacion de su orden de compra).
*Finalmente debe correr el archivo cliente.go para realizar la orden.

