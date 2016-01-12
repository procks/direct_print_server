:: BatchGotAdmin
:-------------------------------------
REM  --> Check for permissions
>nul 2>&1 "%SYSTEMROOT%\system32\cacls.exe" "%SYSTEMROOT%\system32\config\system"

REM --> If error flag set, we do not have admin.
if '%errorlevel%' NEQ '0' (
    echo Requesting administrative privileges...
    goto UACPrompt
) else ( goto gotAdmin )

:UACPrompt
    echo Set UAC = CreateObject^("Shell.Application"^) > "%temp%\getadmin.vbs"
    echo UAC.ShellExecute "%~s0", "", "", "runas", 1 >> "%temp%\getadmin.vbs"

    "%temp%\getadmin.vbs"
    exit /B

:gotAdmin
    if exist "%temp%\getadmin.vbs" ( del "%temp%\getadmin.vbs" )
    pushd "%CD%"
    CD /D "%~dp0"
:--------------------------------------
netsh firewall add portopening TCP 9188 "Direct Print Service" ENABLE ALL
netsh firewall add portopening UDP 9188 "Direct Print Service" ENABLE ALL
prunsrv //US//DirectPrintService --DisplayName="Direct Print Service" ^
        --Install="%~dp0prunsrv.exe" --Jvm=auto --StartMode=jvm --StopMode=jvm ^
		--StartClass=com.solvaig.print.PrintServer --StartParams=start ^
        --StopClass=com.solvaig.print.PrintServer --StopParams=stop ^
        --Startup=auto --Classpath=%CLASSPATH%;PrintServer.jar ^
		--LogPrefix=%SERVICE_NAME% ^
		--LogPath="%APPDATA%\DirectPrintServiceLogs" ^
		--StdOutput=auto ^
		--StdError=auto ^
		--LogLevel=INFO
net start DirectPrintService
pause