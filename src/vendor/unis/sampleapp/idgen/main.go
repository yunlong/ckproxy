package main

import (
	"unis"
)

/*
	Start server : ./idgen -f ./app.ini

	Test commands after the server startup :

	$ curl http://localhost:9360/idgen?count=100
	$ curl http://localhost:9360/status.html
*/

func main() {
	fw := unis.DefaultFramework
	fw.RegisterModule("idgen", new(IdGenModule))
	err := fw.Initialize()
	if err != nil {
		panic(err.Error())
	}

	fw.Run()
}
