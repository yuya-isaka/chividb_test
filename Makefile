# Goコマンドのパスを設定
GO_CMD := go

# テストコマンドのパスを設定
GO_TEST := gotest

# ビルドするバイナリの名前を設定
BINARY_NAME := chibidb

# ビルドするOSとアーキテクチャを設定
BUILD_OS := darwin
BUILD_ARCH := amd64

# ビルド用のフラグを設定
BUILD_FLAGS := -v

# テスト用のフラグを設定
TEST_FLAGS := -v -race -coverprofile=cover.out -covermode=atomic

# カバレッジレポート用のフラグを設定
COVER_FLAGS := -html=cover.out -o cover.html

# 静的解析とテストを実行
test:
	$(GO_CMD) vet ./...
	$(GO_TEST) $(TEST_FLAGS) ./...

# カバレッジレポートを生成してブラウザで開く
check:
	$(GO_CMD) vet ./...
	$(GO_TEST) $(TEST_FLAGS) ./...
	$(GO_CMD) tool cover $(COVER_FLAGS)
	open cover.html

# ビルドしたバイナリを実行
build:
	$(GO_CMD) build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/main.go
	./$(BINARY_NAME)

# ビルドしたバイナリとカバレッジレポートを削除
clean:
	rm -f $(BINARY_NAME) cover.out cover.html