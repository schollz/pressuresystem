
run:
	go-bindata -o bindata.go static/...
	go build -v 
	./pressuresystem.exe 
	
pre:
	go install -v github.com/jteeuwen/go-bindata/go-bindata
