.PHONY: cover
cover: ## Run all the tests with race detection and opens the coverage report
	go test . -coverprofile=coverage.out -race
	go tool cover -html=coverage.out