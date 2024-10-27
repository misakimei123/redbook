.PHONY: docker
docker:
	@rm redbook || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -tags=k8s -o redbook .
	@docker rmi -f misakimei123/redbook:0.0.1
	@docker build -t misakimei123/redbook:0.0.1 .

.PHONY: mock
mock:
	@mockgen -source=./internal/web/ginadaptor/types.go -package=ginadaptormocks -destination=./internal/web/ginadaptor/mocks/types.mock.go
	@mockgen -source=./internal/web/jwt/types.go -package=jwtmocks -destination=./internal/web/jwt/mocks/types.mock.go
	@mockgen -source=./internal/service/article.go -package=svcmocks -destination=./internal/service/mocks/article.mock.go
	@mockgen -source=./internal/service/user.go -package=svcmocks -destination=./internal/service/mocks/user.mock.go
	@mockgen -source=./internal/service/code.go -package=svcmocks -destination=./internal/service/mocks/code.mock.go
	@mockgen -source=./internal/repository/user.go -package=repomocks -destination=./internal/repository/mocks/user.mock.go
	@mockgen -source=./internal/repository/code.go -package=repomocks -destination=./internal/repository/mocks/code.mock.go
	@mockgen -source=./internal/repository/article.go -package=repomocks -destination=./internal/repository/mocks/article.mock.go
	@mockgen -source=./internal/repository/article_author.go -package=repomocks -destination=./internal/repository/mocks/article_author.mock.go
	@mockgen -source=./internal/repository/article_reader.go -package=repomocks -destination=./internal/repository/mocks/article_reader.mock.go
	@mockgen -source=./internal/repository/dao/user.go -package=daomocks -destination=./internal/repository/dao/mocks/user.mock.go
	@mockgen -source=./internal/repository/dao/profile.go -package=daomocks -destination=./internal/repository/dao/mocks/profile.mock.go
	@mockgen -source=./internal/repository/dao/article.go -package=daomocks -destination=./internal/repository/dao/mocks/article.mock.go
	@mockgen -source=./internal/repository/dao/article_author.go -package=daomocks -destination=./internal/repository/dao/mocks/article_author.mock.go
	@mockgen -source=./internal/repository/dao/article_reader.go -package=daomocks -destination=./internal/repository/dao/mocks/article_reader.mock.go
	@mockgen -source=./internal/repository/cache/user.go -package=cachemocks -destination=./internal/repository/cache/mocks/user.mock.go
	@mockgen -source=./internal/repository/cache/code/rediscode.go -package=rediscodemocks -destination=./internal/repository/cache/code/mocks/rediscode.mock.go
	@mockgen -package=redismocks -destination=./internal/repository/cache/redismocks/cmd.mock.go github.com/redis/go-redis/v9 Cmdable
	@mockgen -source=./internal/service/sms/type.go -package=smsmocks -destination=./internal/service/sms/mocks/sms.mock.go
	@mockgen -source=./pkg/limiter/types.go -package=limitmocks -destination=./pkg/limiter/mocks/limit.mock.go
	@mockgen -source=./internal/service/sms/ratelimit/limiter.go -package=limitersmsmocks -destination=./internal/service/sms/ratelimit/mock/limiter.mock.go
	@mockgen -source=./internal/service/sms/repo/sms.go -package=smsrepomocks -destination=./internal/service/sms/repo/mock/sms.mock.go
	@mockgen -source=./internal/service/sms/repo/dao/sms.go -package=smsdao -destination=./internal/service/sms/repo/dao/mock/sms.mock.go

.PHONY: grpc
grpc:
	@buf generate api/proto