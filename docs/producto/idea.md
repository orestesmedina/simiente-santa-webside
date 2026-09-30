# Idea: [Sitio web Iglesia Simiente Santa]
Soy lider de una iglesia evangélica, actualmente no contamos con un sitio web, por lo que necesitamos una pagina web, moderna que cumpla con los estándares de hoy en dia a nivel de desarrollo backend, frontend (animaciones modernas y sobretodo que no sea un diseño típico y comun de IA). Necesitamos un sitio que sea limpio pero moderno, apto para todo público ya que tenemos personas de todas las edades que podrian acceder al sitio, personas jovenes que conocen de tecnología, asi como adultos mucho más mayores que saber lo básico de internet y de navegación.
El objetivo principal del sitio web es que nos demos a conocer en nuestra provicia y por qué no internacionalmente, servicios dominicales, nuestros diferentes ministerios (sus encargados y grupos de distribucion por whatsapps) canales de información por whatsapp, o nuestras diferentes redes sociales. También que descubran nuestras próximas actividades o eventos (acción social, predicas, adoraciones, entrevistas, noches de oración, etc) o noticias, ver información, imagenes o videos de nuestros ultimos eventos recientes. También donde pueden visualizar nuestras grabaciones o podcast (esto está subido en Youtube y Spotify) sin necesidad de tener instaladas esas aplicaciones. Y también una sección de cómo pueden ayudarnos con donaciones, numeros de cuentas IBAN y SINPE Movil y en qué utilizamos esas donaciones.

Y por ultimo pero no menos importanes es una sección de grupos de conexión, esto son grupos, clases o actividades que crea una persona y que las imparte un determinado cantidad de dias a la semana, o mes en donde se enseñan habilidades que pueden o no tener relación a la religio por ejemplo un taller de maquillaje, un taller de fotografia, un taller de asados, un taller de ingles, etc. Dónde o cómo se pueden inscribir, fechas, lugar y toda la info necesaria para que las personas interesadas se puedan registrar, ya sea por correo a la persona encargada, por correo o llenando un formulario.



**Autor:** Orestes · **Cliente / interesado:** Simiente Santa · **Fecha:** 2026-09-28

## 1. Problema

Nuestra iglesia evengelica no cuenta con un sitio web en donde sus miembros o personas nuevas descubra todas nuestras actividades, redes sociales y grupos de conexión o donde hacer donaciones


## 2. Usuarios

Usuarios comunes: serian todos los miembros o personas externas a la iglesia que van a ver la información, eventos, grabaciones. Basicamente es nuestro publico meta.
Usuarios del sistema: son los que pueden administrar lo que ven los usuarios comunes, esto por medio de roles y permisos
Usuarios administradores, son los que pueden administrar lo que ven los usuarios comunes y tambien otorgan permisos a los usuarios del sistema, pueden crear mas usuarios o eliminarlos o desactivarlos.

| Usuario | Qué necesita hacer |
|---|---|
| Usuarios comunes | Navegar por el sitio y ver la información de actividades, eventos, grupos de conexiones, ministerios y dónde o cómo donar |
| Usuarios del sistema | Crear, modificar, eliminar actividades, eventos |
| Usuarios del sistema | Crear, modificar, eliminar videos o podcast |
| Usuarios del sistema | Crear, modificar, eliminar grupos de conexiones|
| Usuarios del sistema | Crear, modificar, eliminar ministerios |
| Usuarios administradores | Todo lo que hace un usuario del sistema y además administrar esos usuarios |

## 3. Cómo se resuelve hoy

solo de vez en cuando se hace un afiche que se comparte en los grupos de whatsapp o facebook pero la gran mayoria del tiempo se olvida entonces nuestras actividades no asisten personas nuevas porque no saben nada.


## 4. Qué sería un éxito


- Las personas miembro de la iglesia tienen un canal unificado donde pueden ver y conocer de las actividades y grupos de la iglesia
- peronas nuevas pueden conocer actividades de la iglesia y por qué no deciden hacerse miembros de la congregación
- Los eventos, grupos de conexión, ministerios, videos y podcast sean dinamicos no algo estático que no se pueden administrar.

## 5. Lo mínimo que debe hacer (MVP)


1. [Desplegar la información de la iglesia, actividades, eventos, noticias, podcast y videos y seccion de donaciones]
2. [Permitir administrar esas secciones mediante un sistema administrador]
3. [Administrar usuarios, permisos y roles]

## 6. Fuera de alcance (por ahora)


- creación de banner (imagenes) con la publicidad de nuestros eventos o actividades con nuestra marca de identidad para redes sociales o compartir en grupos.
- Permitir que los usuarios comunes se puedan registrar para llevar cursos o charlas impartidas por la iglesia.

## 7. Restricciones

- El sitio debe de adaptarse a cualquier dispositivo.
- El sistema debe de manejar permisos, roles y estados.


## 8. Preguntas abiertas


- Podemos hacer que los eventos, actividades se puedan crear por medio de un msj de texto ya sea por Telegram o Whatsapp?
- Podemos generar  imagenes de nuestros eventos o actividades que nos sirva para publicar en nuestras redes sociales?
- Podemos automatizar las publicaciones desde nuestro sitio web y que cuando un usuario del sistema decida, hacer la publicación en nuestra cuenta de instagram, whatsapp, facebook?
- Podemos hacer que cuando subamos un capitulo a spotify o youtube, ese se agrege a nuestro sitio web?
- Podemos crear una sección donde los miembros puedan registrar y llevar un curso virtual en la plataforma?