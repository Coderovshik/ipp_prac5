gen-swagger-doc:
	@swagger generate spec -o ./swagger.yml main.go doc.go