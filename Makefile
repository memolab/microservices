genproto:
	protoc --proto_path=proto/vendor --proto_path=proto/v1 --go_out=common/pb/v1 --go-grpc_out=common/pb/v1 \
	--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
	--doc_out=services/api_gateway_service/static/doc --doc_opt=html,index.html \
	proto/v1/*.proto


.PHONY: gen