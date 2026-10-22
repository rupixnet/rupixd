# instalar-nodo.ps1 (Rupix, 17-oct-2026): un nodo de la testnet de Rupix en un comando, en Windows.
#
#   irm https://rupix.network/instalar.ps1 | iex
#
# Que hace, en orden, y se detiene al primer fallo:
#   1. Encuentra la ultima release publicada.
#   2. Baja el zip de Windows y la lista SHA256SUMS-win64.txt que publica el CI.
#   3. VERIFICA el sha256 de cada programa contra esa lista; si uno no coincide, no instala nada.
#   4. Instala rupixd, rupixctl, rupixwallet y rupixminer en %LOCALAPPDATA%\Rupix\bin (sin administrador)
#      y agrega esa carpeta a tu PATH de usuario.
#   5. Pregunta como quieres que se llame tu nodo en el mapa (https://explorer.rupix.network/mapa.html).
#   6. Crea la tarea programada "Rupix testnet": arranca el nodo al iniciar sesion, en segundo plano, y lo
#      reinicia si se cae. La enciende ahora mismo.
#   7. Corre `rupixwallet --testnet verificar` para que veas tu pantalla de OK.
#
# Antes de tocar nada muestra el plan y espera tu "si". Si esta maquina YA tiene un nodo (tarea, datos o
# rupixd corriendo) SE NIEGA, salvo en modo actualizar (solo reemplaza los programas; no toca tarea ni datos).
#
# Opciones, por variables de entorno (porque con `irm | iex` no se pasan parametros):
#   $env:RUPIX_NOMBRE = "TuNombre"     nombre en el mapa (sin espacios). Vacio = "nodo".
#   $env:RUPIX_VERSION = "v0.6.7"      una release concreta. Por defecto, la ultima.
#   $env:RUPIX_SI = "1"                no pedir confirmacion.
#   $env:RUPIX_ACTUALIZAR = "1"        solo reemplazar los programas de una instalacion existente.
#
# Es texto plano: leelo antes de correrlo (abre https://rupix.network/instalar.ps1 en el navegador).
# No pide contrasenas, no toca wallets, no borra nada. Repo: https://github.com/rupixnet/rupixd
# No confies, verifica.

$ErrorActionPreference = "Stop"
$UltimaConocida = "v0.6.7"   # se actualiza en cada release; ultimo recurso si GitHub no contesta
$Repo = "rupixnet/rupixd"
$Seed1 = "178.104.69.148:17211"
$Tarea = "Rupix testnet"
$Dir = Join-Path $env:LOCALAPPDATA "Rupix"
$Bin = Join-Path $Dir "bin"
$Datos = Join-Path $Dir "testnet"

$Nombre = $env:RUPIX_NOMBRE
$Version = $env:RUPIX_VERSION
$Si = $env:RUPIX_SI -eq "1"
$Actualizar = $env:RUPIX_ACTUALIZAR -eq "1"

function Decir($t) { Write-Host ""; Write-Host "==> $t" -ForegroundColor Yellow }
function Fallo($t) { Write-Host ""; Write-Host "FALLO: $t" -ForegroundColor Red; throw "Rupix: $t" }

# 1) Sistema
if (-not [Environment]::Is64BitOperatingSystem) { Fallo "las releases de Windows son de 64 bits" }
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

# 1b) ¿Ya hay un nodo aqui?
$Existente = @()
if (Get-ScheduledTask -TaskName $Tarea -ErrorAction SilentlyContinue) { $Existente += "tarea '$Tarea'" }
if (Test-Path $Datos) { $Existente += "datos en $Datos" }
if (Get-Process rupixd -ErrorAction SilentlyContinue) { $Existente += "un rupixd corriendo" }
if ($Existente.Count -gt 0 -and -not $Actualizar) {
    Fallo ("esta maquina YA tiene un nodo de Rupix (" + ($Existente -join ", ") + "). No instalo encima. Para reemplazar solo los programas: `$env:RUPIX_ACTUALIZAR='1' y vuelve a correrlo")
}

# 2) Version
if (-not $Version) {
    try {
        $r = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue -UseBasicParsing
    } catch { $r = $_.Exception.Response }
    $loc = $null
    try { $loc = $r.Headers["Location"] } catch {}
    if ($loc -and $loc -match "/tag/(v[0-9][0-9.]*)$") { $Version = $Matches[1] }
    if (-not $Version) { $Version = $UltimaConocida }
}
if ($Version -notmatch "^v[0-9]") { Fallo "version rara: $Version" }
Decir "Rupix $Version para Windows (64 bits)"

# 2b) El plan
Write-Host ""
Write-Host "Esto es lo que va a pasar:"
Write-Host "  1. Bajar rupix-$Version-win64.zip y SHA256SUMS-win64.txt de github.com/$Repo"
Write-Host "  2. Verificar el sha256 de rupixd, rupixctl, rupixwallet y rupixminer; si uno falla, no se instala nada"
if ($Actualizar) {
    Write-Host "  3. ACTUALIZAR: detener el nodo, reemplazar los 4 programas en $Bin, arrancarlo. Tarea y datos NO se tocan"
} else {
    Write-Host "  3. Instalar los 4 programas en $Bin y agregar esa carpeta a tu PATH (sin administrador)"
    Write-Host "  4. Preguntar el nombre de tu nodo para el mapa (viaja en el saludo de la red; nunca tu IP)"
    Write-Host "  5. Crear la tarea programada '$Tarea' (arranca al iniciar sesion, en segundo plano) y encenderla"
    Write-Host "  6. Correr 'rupixwallet --testnet verificar' para que veas tu pantalla de OK"
}
Write-Host "No pide contrasenas, no toca wallets, no borra nada."
if (-not $Si) {
    $resp = Read-Host "`n¿Seguimos? Escribe si para continuar"
    if ($resp -notin @("si","SI","Si","s","S","yes","y")) { Write-Host "Cancelado. No se toco nada."; return }
}

# 3) Descargar y verificar
$Tmp = Join-Path $env:TEMP ("rupix-" + [guid]::NewGuid().ToString("N").Substring(0,8))
New-Item -ItemType Directory -Path $Tmp | Out-Null
$Base = "https://github.com/$Repo/releases/download/$Version"
$Zip = "rupix-$Version-win64.zip"; $Lista = "SHA256SUMS-win64.txt"
Decir "Bajando $Zip y $Lista"
Invoke-WebRequest -Uri "$Base/$Zip" -OutFile (Join-Path $Tmp $Zip) -UseBasicParsing
try { Invoke-WebRequest -Uri "$Base/$Lista" -OutFile (Join-Path $Tmp $Lista) -UseBasicParsing }
catch { Fallo "no pude bajar $Base/$Lista (releases anteriores a v0.6.5 no la tienen)" }
Expand-Archive -Path (Join-Path $Tmp $Zip) -DestinationPath (Join-Path $Tmp "bin") -Force
Decir "Verificando sha256 de cada programa contra la lista publicada"
$esperados = @{}
Get-Content (Join-Path $Tmp $Lista) | ForEach-Object {
    if ($_ -match "^([0-9a-fA-F]{64})\s+\*?(.+)$") { $esperados[(Split-Path -Leaf $Matches[2]).ToLower()] = $Matches[1].ToLower() }
}
$programas = @("rupixd","rupixctl","rupixwallet","rupixminer")
foreach ($p in $programas) {
    $f = Get-ChildItem -Path (Join-Path $Tmp "bin") -Recurse -File | Where-Object { $_.Name -ieq "$p.exe" -or $_.Name -ieq $p } | Select-Object -First 1
    if (-not $f) { Fallo "falta $p en el zip" }
    $esp = $esperados[$f.Name.ToLower()]; if (-not $esp) { $esp = $esperados[$p] }
    if (-not $esp) { Fallo "la lista publicada no trae el sha256 de $($f.Name)" }
    $real = (Get-FileHash -Algorithm SHA256 -Path $f.FullName).Hash.ToLower()
    if ($real -ne $esp) { Fallo "$($f.Name) NO coincide con la lista publicada. No se instala nada. Reporta esto en github.com/$Repo/issues" }
    Write-Host "$($f.Name): OK"
}

# 4) Instalar
if ($Actualizar) { Stop-ScheduledTask -TaskName $Tarea -ErrorAction SilentlyContinue; Get-Process rupixd -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue; Start-Sleep -Seconds 2 }
New-Item -ItemType Directory -Path $Bin -Force | Out-Null
foreach ($p in $programas) {
    $f = Get-ChildItem -Path (Join-Path $Tmp "bin") -Recurse -File | Where-Object { $_.Name -ieq "$p.exe" -or $_.Name -ieq $p } | Select-Object -First 1
    Copy-Item $f.FullName (Join-Path $Bin "$p.exe") -Force
}
$pathUsuario = [Environment]::GetEnvironmentVariable("Path", "User")
if (($pathUsuario -split ";") -notcontains $Bin) {
    [Environment]::SetEnvironmentVariable("Path", ($pathUsuario.TrimEnd(";") + ";" + $Bin), "User")
    $env:Path += ";$Bin"
}
$ver = & (Join-Path $Bin "rupixd.exe") --version 2>$null
Decir "Instalado en ${Bin}: $ver"
Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue

if ($Actualizar) {
    if (Get-ScheduledTask -TaskName $Tarea -ErrorAction SilentlyContinue) { Start-ScheduledTask -TaskName $Tarea; Write-Host "   tarea '$Tarea' arrancada de nuevo (sin tocar su configuracion)" }
    Decir "Actualizacion lista. Tarea y datos quedaron como estaban."
    return
}

# 5) Nombre en el mapa
if (-not $Nombre) { $Nombre = Read-Host "`n¿Como quieres que se llame tu nodo en el mapa? (sin espacios; vacio = 'nodo')" }
$Nombre = ($Nombre -replace '[ <>&"/]', '')
if ($Nombre.Length -gt 32) { $Nombre = $Nombre.Substring(0,32) }
$UA = ""; if ($Nombre) { $UA = " --uacomment=$Nombre" }

# 6) Tarea programada: arranca al iniciar sesion, en segundo plano, se reinicia si se cae
$argumentos = "--testnet --utxoindex --rpclisten=127.0.0.1:17210 --addpeer=$Seed1 --appdir=`"$Datos`"$UA"
$accion = New-ScheduledTaskAction -Execute (Join-Path $Bin "rupixd.exe") -Argument $argumentos -WorkingDirectory $Bin
$disparo = New-ScheduledTaskTrigger -AtLogOn -User $env:USERNAME
$ajustes = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -Hidden -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
Register-ScheduledTask -TaskName $Tarea -Action $accion -Trigger $disparo -Settings $ajustes -Description "Nodo de la testnet de Rupix ($Version). rupix.network" -Force | Out-Null
Start-ScheduledTask -TaskName $Tarea
Decir "Tarea '$Tarea' encendida (arranca sola al iniciar sesion). Para verla: Programador de tareas, o Get-ScheduledTask '$Tarea'"
Write-Host "   Para detener el nodo: Stop-ScheduledTask '$Tarea'   Para quitarlo: Unregister-ScheduledTask '$Tarea'"

# 7) Verificar
Decir "Esperando 25 s a que el nodo arranque y verificando"
Start-Sleep -Seconds 25
& (Join-Path $Bin "rupixwallet.exe") --testnet verificar
Write-Host ""
Write-Host "Mientras sincroniza es normal ver AVISO en 'sincronizado'. Tu luz: https://explorer.rupix.network/mapa.html"
Decir "Listo. Abre una PowerShell nueva para que el PATH tome efecto. Guia: https://rupix.network/guia.html · No confies, verifica."
