# generate_cert:
# 	cd insecure && go run "$$(go env GOROOT)/src/crypto/tls/generate_cert.go" \
# 		--host=localhost,127.0.0.1 \
# 		--ecdsa-curve=P256 \
# 		--ca=true

serve_local:
	docker-compose up --build