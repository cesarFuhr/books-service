up:
	docker compose up -d --build

down:
	docker compose down

test: test-env-up	
	go test -p=1 ./... 

test-env-up:
	docker compose -f docker-compose.test.yml up -d

test-env-down:
	docker compose -f docker-compose.test.yml down


 