# MIA_Proyecto2_201900042

## Proyecto 2 -  Manejo e Implementación de Archivos

Proyecto dedicado a la simulación de un disco duro junto con la creación de un sistema de archivos EXT2.

### Frontend

Realizado con la librería de React.js para poder escribir los comandos, puede encontrar el código fuente [aqui](./Frontend/miap2_frontend/)

Para ejecutarlo necesita hacer lo siguiente:

  1. Escribir en la consola, ubicandose en la carpeta del código frontend lo siguiente: `npm install`
  2. Posteriormente escriba el siguiente comando para poder levantar la aplicación: `npm run dev`

### Backend

Realizado con el lenguaje de programación Golang para la lógica de la ejecución de los comandos escritos en el frontend, puede encontrar el código fuente [aqui](./Backend/)

Para ejecutarlo necesita hacer lo siguiente:
  
  1. Escribir en la consola, ubicandose en la carpeta del código Backend lo siguiente: `sudo go run main.go` ó al ejecutable (en linux) `sudo ./main`

### Comandos

- ## Mkdisk
Este comando creará un archivo binario que simulará un disco duro, estos
archivos binarios tendrán la extensión dsk y su contenido al inicio será 0. Contiene los siguientes parametros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>size    |Obligatorio |Este parámetro recibirá un número que indicará el tamaño del disco a crear. Debe ser positivo y mayor que cero, si no se mostrará un error.|
|>path    |Obligatorio |Este parámetro será la ruta en el que se creará el archivo que representará el disco duro. Si las carpetas de la ruta no existen deberán crearse.|
|>fit     |Opcional    |Indicará el ajuste que utilizará el disco para crear las particiones dentro del disco Podrá tener los siguientes valores: `BF`: Indicará el mejor ajuste (Best Fit), `FF`: Utilizará el primer ajuste (First Fit), `WF`: Utilizará el peor ajuste (Worst Fit). Ya que es opcional, se tomará el primer ajuste(FF) si no está especificado en el comando.|
|>unit    |Opcional    |Este parámetro recibirá una letra que indicará las unidades que utilizará el parámetro size. Podrá tener los siguientes valores: `K` que indicará que se utilizarán Kilobytes (1024 bytes), `M` en el que se utilizarán Megabytes (1024 * 1024 bytes) Este parámetro es opcional, si no se encuentra se creará un disco con tamaño en Megabytes. |

Ejemplos:

```bash
#Crea un disco de 3000 Kb en la carpeta home
mkdisk >Size=3000 >unit=K >path=/home/user/Disco1.dsk
#No es necesario utilizar comillas para la ruta en este caso ya que la ruta no tiene ningún espacio en blanco
mkdisk >path=/home/user/Disco2.dsk >Unit=K >size=3000
#Se ponen comillas por la carpeta “mis discos” ya que tiene espacios en blanco, se crea si no está no existe
mkdisk >size=5 >unit=M >path="/home/mis discos/Disco3.dsk"
#Creará un disco de 10 Mb ya que no hay parámetro unit
mkdisk >size=10 >path="/home/mis discos/Disco4.dsk"
```

- ## Rmdisk
Este comando elimina un archivo que representa a un disco duro mostrando
un mensaje de confirmación para eliminar. Tendrá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>path    |Obligatorio |Este parámetro será la ruta en el que se eliminará el archivo que representará el disco duro. |

Ejemplos:
```bash
#Elimina Disco4.dsk
rmdisk >path="/home/mis discos/Disco4.dsk”
```

- ## Fdisk
Este comando administra las particiones en el archivo que representa al disco duro. Tendrá los siguientes parámetros:
|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>size    |Obligatorio al crear |Este parámetro recibirá un número que indicará el tamaño de la partición a crear. Debe ser positivo y mayor a cero, de lo contrario se mostrará un mensaje de error. |
|>unit    |Opcional |Este parámetro recibirá una letra que indicará las unidades que utilizará el parámetro size. Podrá tener los siguientes valores: `B`: indicará que se utilizarán bytes, `K`: indicará que se utilizarán Kilobytes(1024 bytes), `M`:indicará que se utilizarán Megabytes(1024 * 1024 bytes). Este parámetro es opcional, si no se encuentra se creará una partición en Kilobytes. Si se utiliza un valor diferente mostrará un mensaje de error. |
|>path    |Obligatorio |Este parámetro será la ruta en la que se encuentra el disco en el que se creará la partición. Este archivo ya debe existir, si no se mostrará un error. |
|>type    |Opcional |Indicará que tipo de partición se creará. Ya que es opcional, se tomará como primaria en caso de que no se indique. Podrá tener los siguientes valores: `P`: en este caso se creará una partición primaria. `E`: en este caso se creará una partición extendida. `L`: Con este valor se creará una partición lógica. Las particiones lógicas sólo pueden estar dentro de la extendida sin sobrepasar su tamaño. Deberá tener en cuenta las restricciones de teoría de particiones: La suma de primarias y extendidas debe ser como máximo 4. Solo puede haber una partición extendida por disco. No se puede crear una partición lógica si no hay una extendida. Si se utiliza otro valor diferente a los anteriores deberá mostrar un mensaje de error.|
|>fit    |Opcional |Indicará el ajuste que utilizará la partición para asignar espacio. Podrá tener los siguientes valores: `BF`: Indicará el mejor ajuste (Best Fit), `FF`: Utilizará el primer ajuste (First Fit), `WF`: Utilizará el peor ajuste (Worst Fit) Ya que es opcional, se tomará el peor ajuste(WF) si no está especificado en el comando. Si se utiliza otro valor que no sea alguno de los anteriores mostrará un mensaje de error. |
|>name    |Obligatorio |Indicará el nombre de la partición. El nombre no debe repetirse dentro de las particiones de cada disco. Si se va a eliminar, la partición ya debe existir, si no existe debe mostrar un mensaje de error.|

Ejemplos:

```bash
#Crea una partición primaria llamada Particion1 de 300kb
#con el peor ajuste en el disco Disco1.dsk
fdisk >Size=300 >path=/home/Disco1.dsk >name=Particion1
#Crea una partición extendida dentro de Disco2 de 300kb
#Tiene el peor ajuste
fdisk >type=E >path=/home/Disco2.dsk >Unit=K >name=Particion2 >size=300
#Crea una partición lógica con el mejor ajuste, llamada Partición 3,
#de 1 Mb en el Disco3
fdisk >size=1 >type=L >unit=M >fit=BF >path="/mis discos/Disco3.dsk" >name="Particion3"
#Intenta crear una partición extendida dentro de Disco2 de 200 kb
#Debería mostrar error ya que ya existe una partición extendida
#dentro de Disco2
fdisk >type=E >path=/home/Disco2.dsk >name=Part3 >Unit=K >size=200
```

- ## Mount
Este comando montará una partición del disco en el sistema. 

Cada partición se identificará por un id que tendrá la siguiente estructura utilizando el número de carnet:
*Últimos dos dígitos del Carnet + Número + Letra Ejemplo: carnet = 201900042

Id´s = 421A, 421B, 421C, 422A, 423A
|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>path    |Obligatorio|Este parámetro será la ruta en la que se encuentra el disco que se montará en el sistema. Este archivo ya debe existir.|
|>name    |Obligatorio|Indica el nombre de la partición a cargar. Si no existe debe mostrar error |

Ejemplos:

```bash
#Monta las particiones de Disco1.dsk
#carnet = 201900042
mount >path=/home/Disco1.dsk >name=Part1 #id=421a
mount >path=/home/Disco2.dsk >name=Part1 #id=422a
mount >path=/home/Disco3.dk >name=Part2 #id=423a
```

- ## Mkfs

Este comando realiza un formateo completo de la partición, se formatea
como ext2. También creará un archivo en la raíz llamado users.txt que tendrá los usuarios y contraseñas del sistema de archivos.

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>id      |Obligatorio|Indicará el id que se generó con el comando mount.Si no existe mostrará error. Se utilizará para saber la partición y el disco que se utilizará para hacer el sistema de archivos.|
|>type    |Opcional|Indicará que tipo de formateo se realizará. Podrá tener los siguientes valores: `Full`: en este caso se realizará un formateo completo. Ya que es opcional, se tomará como un formateo completo si no se especifica esta opción.|

Ejemplos:

```bash
#Realiza un formateo completo de la partición en el id 421A en ext2
mkfs >type=full >id=421A
#Realiza un formateo completo de la partición que ocupa el id 062A
mkfs >id=422A
```

- ## Login

Este comando se utiliza para iniciar sesión en el sistema. No se puede iniciar otra sesión sin haber hecho un `LOGOUT` antes, si no, debe mostrar un mensaje de error indicando que debe cerrar sesión. Recibirá los Siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>id      |Obligatorio|Indicará el id de la partición montada de la cual van a iniciar sesión. De lograr iniciar sesión todas las acciones se realizarán sobre este id.|
|>user      |Obligatorio|Especifica el nombre del usuario que iniciará sesión. Si no se encuentra mostrará un mensaje indicando que el usuario no existe. Va a distinguir mayúsculas de minúsculas.|
|>pwd      |Obligatorio|Indicará la contraseña del usuario que inicia sesión. Si no coincide debe mostrar un mensaje de autenticación fallida. Va a distinguir entre mayúsculas y minúsculas.|

Ejemplos:

```bash
#Se loguea en el sistema como usuario root
login >user=root >pwd=123 >id=422A
#Debe dar error porque ya hay un usuario logueado
login >user="mi usuario" >pwd="mi pwd" >id=422A
```

- ## Logout

Este comando se utiliza para cerrar sesión. Debe haber una sesión activa
anteriormente para poder utilizarlo, si no, debe mostrar un mensaje de error. Este comando no recibe parámetros.

Ejemplo:

```bash
#Termina la sesión del usuario
Logout
```

- ## Mkgrp
Este comando creará un grupo para los usuarios de la partición y se guardará en el archivo users.txt de la partición, este comando solo lo puede utilizar el usuario root. Si otro usuario lo intenta ejecutar, deberá mostrar un mensaje de error, si el grupo a ingresar ya existe deberá mostrar un mensaje de error. Distinguirá entre mayúsculas y minúsculas. Recibirá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>name    |Obligatorio|Indicará el nombre que tendrá el grupo|

Ejemplo:

```bash
#Crea el grupo usuarios en la partición de la sesión actual
mkgrp >name=usuarios
mkgrp >name="grupo 1"
#Debe mostrar mensaje de error ya que el grupo ya existe
mkgrp >name="grupo 1"
```

- ## Rmgrp

Este comando eliminará un grupo para los usuarios de la partición. Solo lo
puede utilizar el usuario root, si lo utiliza alguien más debe mostrar un error. Recibirá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>name    |Obligatorio|Indicará el nombre del grupo a eliminar. Si el grupo no se encuentra dentro de la partición debe mostrar un error.|

Ejemplo:

```bash
#Elimina el grupo de usuarios en la partición de la sesión actual
rmgrp >name=usuarios
#Debe mostrar mensaje de error ya que el grupo no existe porque ya fue eliminado
rmgrp >name=usuarios
```

- ## Mkuser
Este comando crea un usuario en la partición. Solo lo puede ejecutar el
usuario root, si lo utiliza otro usuario deberá mostrar un error. Recibirá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>user    |Obligatorio|Indicará el nombre del usuario a crear, si ya existe, deberá mostrar un error indicando que ya existe el usuario. Máximo: 10 caracteres.|
|>pwd    |Obligatorio|Indicará la contraseña del usuario Máximo 10 Caracteres|
|>grp    |Obligatorio|Indicará el grupo al que pertenece el usuario. Debe de existir en la partición en la que se está creando el usuario, si no debe mostrar un mensaje de error. Máximo 10 Caracteres|

Ejemplo:

```bash
#Crea usuario user1 en el grupo ‘usuarios’
mkusr >user=user1 >pwd=usuario >grp=usuarios
#Debe mostrar mensaje de error ya que el usuario ya existe
#independientemente que este en otro grupo
mkusr >user=user1 >pwd=usuario >grp=usuarios2
```

- ## Rmusr

Este comando elimina un usuario en la partición. Solo lo puede ejecutar el
usuario root, si lo utiliza otro usuario deberá mostrar un error. Recibirá los
siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>user    |Obligatorio|Indicará el nombre del usuario a eliminar. Si el usuario no se encuentra dentro de la partición debe mostrar un error.|

Ejemplo:

```bash
#Elimina el usuario user1
rmusr >user=user1
#Debe mostrar mensaje de error porque el user1 ya no existe
rmusr >user=user1
```

- ## Mkfile
Este comando permitirá crear un archivo, el propietario será el usuario que
actualmente ha iniciado sesión. Tendrá los permisos 664. El usuario
deberá tener el permiso de escritura en la carpeta padre, si no debe mostrar un error. Tendrá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>path    |Obligatorio|Este parámetro será la ruta del archivo que se creará. Si lleva espacios en blanco deberá encerrarse entre comillas. Si ya existe debe mostrar un mensaje si se desea sobreescribir el archivo. Si no existen las carpetas padres, debe mostrar error, a menos que se utilice el parámetro r, que se explica  posteriormente.|
|>r    |Opcional|Si se utiliza este parámetro y las carpetas especificadas por el parámetro path no existen, entonces deben crearse las carpetas padres. Si ya existen, no deberá crear las carpetas. No recibirá ningún valor, si lo recibe debe mostrar error.|
|>size    |Opcional|Este parámetro indicará el tamaño en bytes del archivo, El contenido serán números del 0 al 9 cuantas veces sea necesario hasta cumplir el tamaño ingresado. Si no se utiliza este parámetro, el tamaño será 0 bytes. Si es negativo debe mostrar error.|
|>cont    |Opcional|Indicará un archivo en el disco duro de la computadora que tendrá el contenido del archivo. Se utilizará para cargar contenido en el archivo. La ruta ingresada debe existir, sino mostrará un mensaje de error.|

Si se ingresan los parámetros cont y size, tendra mayor prioridad el parametro cont

Ejemplo:

```bash
#Crea el archivo a.txt
#Si no existen las carpetas home user o docs se crean
#El tamaño del archivo es de 15 bytes #El contenido sería:
#012345678901234
mkfile >size=15 >path=/home/user/docs/a.txt >r
#Crea "archivo 1.txt" la carpeta "mis documentos" ya debe existir
#el tamaño es de 0 bytes
mkfile >path="/home/mis documentos/archivo 1.txt"
#Crea el archivo b.txt
#El contenido del archivo será el mismo que el archivo b.txt
#que se encuentra en el disco duro de la computadora.
mkfile >path=/home/user/docs/b.txt >r >cont=/home/Documents/b.txt
```

- ## Mkdir

Este comando es similar a mkfile, pero no crea archivos, sino carpetas. El
propietario será el usuario que actualmente ha iniciado sesión. Tendrá los
permisos 664. El usuario deberá tener el permiso de escritura en la carpeta
padre, si no debe mostrar un error. Tendrá los siguientes parámetros:

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>path    |Obligatorio|Este parámetro será la ruta de la carpeta que se creará. Si lleva espacios en blanco deberá encerrarse entre comillas. Si no existen las carpetas padres, debe mostrar error, a menos que se utilice el parámetro r.|
|>r    |Opcional|Si se utiliza este parámetro y las carpetas padres en el parámetro path no existen, entonces deben crearse. Si ya existen, no realizará nada. No recibirá ningún valor, si lo recibe debe mostrar error.|

Ejemplo:

```bash
#Crea la carpeta usac
#Si no existen las carpetas home user o docs se crean
mkdir >r >path=/home/user/docs/usac
#Crea la carpeta "archivos diciembre"
#La carpeta padre ya debe existir
mkdir >path="/home/mis documentos/archivos diciembre"
```

- ## Rep

Recibirá el nombre del reporte que se desea y lo generará con graphviz en
el apartado de reportes.

|parametro|categoria   | descripcion|
|---------|---------   |------------|
|>name    |Obligatorio|Nombre del reporte a generar. Tendrá los siguientes valores: `disk`, `tree`, `file`, `sb`. Si recibe otro valor que no sea alguno de los anteriores, debe mostrar un error.|
|>path    |Obligatorio|Si recibe otro valor que no sea alguno de los anteriores, debe mostrar un error. Indica una carpeta y el nombre que tendrá el reporte. Si no existe la carpeta, deberá crearla. Si lleva espacios se encerrará entre comillas|
|>id    |Obligatorio|Indica el id de la partición que se utilizará. Si el reporte es sobre la información del disco, se utilizará el disco al que pertenece la partición. Si no existe debe mostrar un error.|
|>ruta    |Opcional|Funcionará para el reporte file. Será el nombre del archivo o carpeta del que se mostrará el reporte. Si no existe muestra error.|


