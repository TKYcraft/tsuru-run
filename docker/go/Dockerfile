# ビルド用の環境
FROM golang:1.18-alpine as builder

# go installにgitを使っているようなのでadd
RUN apk update && apk --no-cache add git

# ディレクトリの作成
RUN mkdir -p /go_usr/app/src /go_usr/app/build

# ワーキングディレクトリの設定
WORKDIR /go_usr/app

RUN go mod init TKYcraft/tsuru-run

# 使用しているパッケージのインストール
RUN --mount=type=cache,target=${GOPATH}/src \
    go get github.com/bwmarrin/discordgo
# RUN go get github.com/bwmarrin/dgvoice

# ホストのファイルをコンテナの作業ディレクトリに移行
COPY ./app/src /go_usr/app/src

# buildを超えてキャッシュを保持し、ファイルが変更されていなければキャッシュの利用ができる
RUN --mount=type=cache,target=/.cache/go-build \
    CGO_ENABLED=0 \
    go build -installsuffix cgo -o build/main.build src/main.go

# 実行ファイルのみの環境
FROM gcr.io/distroless/static as dev

COPY ./app/file /go_usr/file

# ビルド用の環境から実行ファイルをコピー
COPY --from=builder /go_usr/app/build/main.build /go_usr/app/build/main.build

# 実行
ENTRYPOINT  [ "/go_usr/app/build/main.build" ]
