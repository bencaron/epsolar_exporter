

build:
	go build . 


buildarm:
	GOARCH=arm go build -o epsolar_exporter.arm . 


topi: buildarm
	scp epsolar_exporter.arm baron.cookie:~/epsolar_exporter