
.PHONY: release
release:
	@echo "start release server"
	@echo "pkg/api/ : 保存 gate 的对象"
	@mkdir -p build/engine-gate/bin
	@go build -o build/engine-gate/bin/server ./cmd/server
	@tar zcvf engine-gate.tar.gz build/

.PHONY: clean
clean:
	@echo "start clean"
	@rm build/* -r