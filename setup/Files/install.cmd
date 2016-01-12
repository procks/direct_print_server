netsh firewall add portopening TCP 9188 "Direct Print Service" ENABLE ALL
netsh firewall add portopening UDP 9188 "Direct Print Service" ENABLE ALL
prunsrv //US//DirectPrintService --DisplayName="Direct Print Service" ^
        --Install="%~dp0prunsrv.exe" --Jvm=auto --StartMode=jvm --StopMode=jvm ^
		--StartClass=com.solvaig.print.PrintServer --StartParams=start ^
        --StopClass=com.solvaig.print.PrintServer --StopParams=stop ^
        --Startup=auto --Classpath=%CLASSPATH%;PrintServer.jar ^
		--ServiceUser=%1 --ServicePassword=%2
net start DirectPrintService