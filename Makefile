migrate:
	go run github.com/prisma/prisma-client-go db push
	go run github.com/prisma/prisma-client-go generate

build:
	sudo docker build -t bendi/image-host .

run:
	go run .