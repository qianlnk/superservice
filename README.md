
	 SSSSSSSSSS     UUUU    UUUU    PPPPPPPPPPP 	    PPPPPPPPPPP     EEEEEEEEEEEE    RRRRRRRRRRR
	SSSS   SSSSS    UUUU    UUUU    PPPP    PPPP     PPPP    PPPP    EEEE            RRRR  RRRRRR
	 SSSSS          UUUU    UUUU    PPPP    PPPP     PPPP    PPPP    EEEE            RRRR    RRRR
	  SSSSSSSS      UUUU    UUUU    PPPP  PPPPPP     PPPP  PPPPPP    EEEEEEEEEEEE    RRRRRRRRRRR
	     SSSSSS     UUUU    UUUU    PPPPPPPPPP 		PPPPPPPPPP      EEEE            RRRRRRRRRR
	SSSS    SSSS    UUUU    UUUU    PPPP 			PPPP            EEEE            RRRR  RRRR
	SSSSSSSSSSS     UUUUUUUUUUUU    PPPP 			PPPP            EEEE            RRRR  RRRRRR
	  SSSSSSSS        UUUUUUUU      PPPP             PPPP            EEEEEEEEEEEE    RRRR     RRRR SERVICE
　　　
# superservice
----
Super service is a project witch collection of release, monitoring, alarm and operation service.
## operation service
You can release version by superservicectl cmd `relase`, and then add service to superservice, it will create by config.
It will start itself when AutoStart true, and restart when program exit use to AutoRestart.
Support command:
list					list all services, include detail information
start servicename/all
stop servicename/all	
restart servicename/all
delete servicename/all
update servicename/all
add servicename/all
log servicename			tailf service's log
## release
cmd release will send your program to service machine specify.
## monitor
It will send some machine info to superservicectl when you set monitor true.
## alarm
Send email according to the condition you set.