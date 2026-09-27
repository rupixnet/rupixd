# ops — cómo está montado el seed

Copias de los archivos de operación del servidor `rupix-seed-1`, para poder reconstruirlo si se pierde.
**Sin secretos:** contraseñas y llaves no están aquí. / *Operation files of the seed server, so it can be rebuilt. No secrets here.*

- `systemd/`: los servicios (nodo, wallet, minero y los demás `rupix*`). Van en `/etc/systemd/system/`; luego `systemctl daemon-reload && systemctl enable --now <servicio>`.
- `rupix-monitor-descalificados.sh`: la alarma de bloques descalificados (v2.1). Va en `/root/`. La línea del cron está en `crontab.txt`.
- `nginx/`: el sitio `explorer.rupix.network`, que pasa las peticiones a `rupixexplorer` en `127.0.0.1:8090`. Los certificados TLS no están aquí: se vuelven a emitir en el servidor nuevo.
- **Mantener al día:** si cambias algo de esto en el servidor, cópialo aquí y commitea.
- `logrotate/`: rotación de los logs de salida del nodo, minero y wallet. Comprime y guarda todas las copias (3650); no se borra nada.
- `retirados/`: servicios que ya no se usan, guardados como memoria (p. ej. el nodo v0.3.0).
