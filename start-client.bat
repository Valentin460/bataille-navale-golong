@echo off
echo ╔════════════════════════════════════════════════════════════════╗
echo ║              🚢 BATAILLE NAVALE - CLIENT 🚢                    ║
echo ╚════════════════════════════════════════════════════════════════╝
echo.
echo Choisissez le mode de démarrage :
echo.
echo 1. Démarrer avec adresses adverses prédéfinies
echo 2. Démarrer et saisir les adversaires manuellement
echo 3. Quitter
echo.
set /p choice=Votre choix (1-3) : 

if "%choice%"=="1" goto predefined
if "%choice%"=="2" goto manual
if "%choice%"=="3" exit /b 0

echo Choix invalide
pause
exit /b 1

:predefined
echo.
set /p port=Port du serveur (défaut: 8080) : 
if "%port%"=="" set port=8080

set /p opponents=Adresses adverses (séparées par des virgules) : 

if "%opponents%"=="" (
    go run main.go --port=%port%
) else (
    go run main.go --port=%port% --opponents=%opponents%
)
goto end

:manual
echo.
set /p port=Port du serveur (défaut: 8080) : 
if "%port%"=="" set port=8080

go run main.go --port=%port%
goto end

:end
pause
