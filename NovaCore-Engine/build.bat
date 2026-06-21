@echo off
setlocal EnableDelayedExpansion

set "ROOT=%CD%"
set /p VERSION=<version.txt
set MAIN_CLASS=dev.novastep.core.Main
set "SRC_DIR=%ROOT%\src\main\java"
set "BUILD_CLASSES=%ROOT%\build\classes"
set "LIB_DIR=%ROOT%\build\libs\deps"
set "EXTRACT_DIR=%ROOT%\build\extract"
set "JAR_OUT=%ROOT%\build\libs\novacore-engine.jar"
set "MANIFEST=%ROOT%\build\MANIFEST.MF"
set "SOURCES_FILE=%ROOT%\build\sources.txt"

echo [Build] novacore-engine v%VERSION% - NovaStepStudios
echo.
java -version >NUL 2>&1
if %ERRORLEVEL% neq 0 ( echo [ERROR] Java no encontrado. & exit /b 1 )
for /f "tokens=3" %%v in ('java -version 2^>^&1 ^| findstr /i "version"') do echo [Info] Java: %%v
echo.
if not exist "%BUILD_CLASSES%" mkdir "%BUILD_CLASSES%"
if not exist "%LIB_DIR%"       mkdir "%LIB_DIR%"
if not exist "%ROOT%\build\libs" mkdir "%ROOT%\build\libs"

echo [Deps] Verificando dependencias...
set "WRAPPER_JAR=%ROOT%\gradle\wrapper\gradle-wrapper.jar"
if not exist "%WRAPPER_JAR%" (
    echo Descargando Gradle Wrapper...
    powershell -Command "Invoke-WebRequest -Uri 'https://github.com/gradle/gradle/raw/v8.0.2/gradle/wrapper/gradle-wrapper.jar' -OutFile '%WRAPPER_JAR%' -UseBasicParsing"
    if not exist "%WRAPPER_JAR%" ( echo [ERROR] No se pudo descargar Gradle Wrapper. & exit /b 1 )
)
call .\gradlew.bat build --quiet
if %ERRORLEVEL% neq 0 ( echo [ERROR] Gradle build fallido. & exit /b 1 )
echo [Deps] OK.
echo.

echo [Package] Creando fat JAR...
call .\gradlew.bat shadowJar --quiet
if %ERRORLEVEL% neq 0 ( echo [ERROR] Shadow JAR fallido. & exit /b 1 )
echo [Package] OK.

if not exist "%JAR_OUT%" ( echo [ERROR] No se creo el JAR. & exit /b 1 )
echo [Package] OK.
echo.
echo ====================================================
echo  Build exitoso!
echo  JAR: build\libs\novacore-engine.jar
echo ====================================================
echo.
echo Para ejecutar:
echo    java -jar build\libs\novacore-engine.jar
echo    java -jar build\libs\novacore-engine.jar --port 7878 --ws-port 7879 --threads 32